/*
Copyright 2019 The Skaffold Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package deploy

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/blang/semver"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/build"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/color"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/constants"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/docker"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/event"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/runner/runcontext"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/util"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/warnings"
)

type HelmDeployer struct {
	*latest.HelmDeploy

	kubeContext string
	kubeConfig  string
	namespace   string
	forceDeploy bool
	// bV is the helm binary version
	bV semver.Version
}

// NewHelmDeployer returns a new HelmDeployer for a DeployConfig filled
// with the needed configuration for `helm`
func NewHelmDeployer(runCtx *runcontext.RunContext) *HelmDeployer {
	logrus.Infof("NEW!")

	return &HelmDeployer{
		HelmDeploy:  runCtx.Cfg.Deploy.HelmDeploy,
		kubeContext: runCtx.KubeContext,
		kubeConfig:  runCtx.Opts.KubeConfig,
		namespace:   runCtx.Opts.Namespace,
		forceDeploy: runCtx.Opts.Force,
	}
}

func (h *HelmDeployer) Labels() map[string]string {
	return map[string]string{
		constants.Labels.Deployer: "helm",
	}
}

func (h *HelmDeployer) Deploy(ctx context.Context, out io.Writer, builds []build.Artifact, labellers []Labeller) *Result {
	logrus.Infof("DEPLOY!")
	event.DeployInProgress()

	var dRes []Artifact
	nsMap := map[string]struct{}{}
	valuesSet := map[string]bool{}

	// Deploy every release
	for _, r := range h.Releases {
		results, err := h.deployRelease(ctx, out, r, builds, valuesSet)
		if err != nil {
			releaseName, _ := expandTemplate(r.Name)

			event.DeployFailed(err)
			return NewDeployErrorResult(errors.Wrapf(err, "deploying %s", releaseName))
		}

		// collect namespaces
		for _, r := range results {
			if trimmed := strings.TrimSpace(r.Namespace); trimmed != "" {
				nsMap[trimmed] = struct{}{}
			}
		}

		dRes = append(dRes, results...)
	}

	// Let's make sure that every image tag is set with `--set`.
	// Otherwise, templates have no way to use the images that were built.
	for _, build := range builds {
		if !valuesSet[build.Tag] {
			warnings.Printf("image [%s] is not used.", build.Tag)
			warnings.Printf("image [%s] is used instead.", build.ImageName)
			warnings.Printf("See helm sample for how to replace image names with their actual tags: https://github.com/GoogleContainerTools/skaffold/blob/master/examples/helm-deployment/skaffold.yaml")
		}
	}

	event.DeployComplete()

	labels := merge(h, labellers...)
	labelDeployResults(labels, dRes)

	// Collect namespaces in a string
	namespaces := make([]string, 0, len(nsMap))
	for ns := range nsMap {
		namespaces = append(namespaces, ns)
	}

	return NewDeploySuccessResult(namespaces)
}

func (h *HelmDeployer) Dependencies() ([]string, error) {
	var deps []string
	for _, release := range h.Releases {
		deps = append(deps, release.ValuesFiles...)

		if release.Remote {
			// chart path is only a dependency if it exists on the local filesystem
			continue
		}

		chartDepsDir := filepath.Join(release.ChartPath, "charts")
		err := filepath.Walk(release.ChartPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return errors.Wrapf(err, "failure accessing path '%s'", path)
			}

			if !info.IsDir() {
				if !strings.HasPrefix(path, chartDepsDir) || release.SkipBuildDependencies {
					// We can always add a dependency if it is not contained in our chartDepsDir.
					// However, if the file is in  our chartDepsDir, we can only include the file
					// if we are not running the helm dep build phase, as that modifies files inside
					// the chartDepsDir and results in an infinite build loop.
					deps = append(deps, path)
				}
			}

			return nil
		})
		if err != nil {
			return deps, errors.Wrap(err, "issue walking releases")
		}
	}
	sort.Strings(deps)
	return deps, nil
}

// Cleanup deletes what was deployed by calling Deploy.
func (h *HelmDeployer) Cleanup(ctx context.Context, out io.Writer) error {
	for _, r := range h.Releases {
		if err := h.deleteRelease(ctx, out, r); err != nil {
			releaseName, _ := expandTemplate(r.Name)
			return errors.Wrapf(err, "deploying %s", releaseName)
		}
	}
	return nil
}

func (h *HelmDeployer) helm(ctx context.Context, out io.Writer, useSecrets bool, arg ...string) error {
	logrus.Infof("helm: %v", arg)
	args := append([]string{"--kube-context", h.kubeContext}, arg...)
	args = append(args, h.Flags.Global...)
	if h.kubeConfig != "" {
		args = append(args, "--kubeconfig", h.kubeConfig)
	}

	if useSecrets {
		args = append([]string{"secrets"}, args...)
	}

	cmd := exec.CommandContext(ctx, "helm", args...)
	cmd.Stdout = out
	cmd.Stderr = out

	return util.RunCmd(cmd)
}

func (h *HelmDeployer) deployRelease(ctx context.Context, out io.Writer, r latest.HelmRelease, builds []build.Artifact, valuesSet map[string]bool) ([]Artifact, error) {

	releaseName, err := expandTemplate(r.Name)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse the release name template")
	}
	hv, err := h.binVer(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "binary version")
	}

	o := installOpts{
		releaseName: releaseName,
		upgrade:     true,
		helmVersion: hv,
	}

	if err := h.helm(ctx, ioutil.Discard, false, "get", releaseName); err != nil {
		color.Yellow.Fprintf(out, "Helm release %s not installed. Installing...\n", releaseName)
		o.upgrade = false
	}

	// Dependency builds should be skipped when trying to install a chart
	// with local dependencies in the chart folder, e.g. the istio helm chart.
	// This decision is left to the user.
	// Dep builds should also be skipped whenever a remote chart path is specified.
	if !r.SkipBuildDependencies && !r.Remote {
		// First build dependencies.
		logrus.Infof("Building helm dependencies...")
		if err := h.helm(ctx, out, false, "dep", "build", r.ChartPath); err != nil {
			return nil, errors.Wrap(err, "building helm dependencies")
		}
	}

	if h.namespace != "" {
		o.namespace = h.namespace
	} else if r.Namespace != "" {
		o.namespace = r.Namespace
	}

	// Overrides.Values
	if len(r.Overrides.Values) != 0 {
		overrides, err := yaml.Marshal(r.Overrides)
		if err != nil {
			return nil, errors.Wrap(err, "cannot marshal overrides to create overrides values.yaml")
		}

		if err := ioutil.WriteFile(constants.HelmOverridesFilename, overrides, 0666); err != nil {
			return nil, errors.Wrapf(err, "cannot create file %s", constants.HelmOverridesFilename)
		}
		defer func() {
			os.Remove(constants.HelmOverridesFilename)
		}()
	}

	// There are 2 strategies:
	// 1) Deploy chart directly from filesystem path or from repository
	//    (like stable/kubernetes-dashboard). Version only applies to a
	//    chart from repository.
	// 2) Package chart into a .tgz archive with specific version and then deploy
	//    that packaged chart. This way user can apply any version and appVersion
	//    for the chart.
	if r.Packaged != nil {
		chartPath, err := h.packageChart(ctx, r)
		if err != nil {
			return nil, errors.WithMessage(err, "cannot package chart")
		}
		o.chartPath = chartPath
	}

	args, err := h.installArgs(r, builds, valuesSet, o)
	if err != nil {
		return nil, errors.Wrap(err, "release args")
	}
	err = h.helm(ctx, out, r.UseHelmSecrets, args...)
	return h.getDeployResults(ctx, o.namespace, releaseName), err
}

type installOpts struct {
	releaseName string
	namespace   string
	chartPath   string
	upgrade     bool
	helmVersion semver.Version
}

// installArgs calculates what arguments to pass in for deployment installation
func (h *HelmDeployer) installArgs(r latest.HelmRelease, builds []build.Artifact, valuesSet map[string]bool, o installOpts) ([]string, error) {

	var args []string
	if !o.upgrade {
		args = append(args, "install")
		if o.helmVersion.LT(semver.MustParse("3.0.0")) {
			args = append(args, "--name")
		}
		args = append(args, o.releaseName)
		args = append(args, h.Flags.Install...)
	} else {
		args = append(args, "upgrade", o.releaseName)
		args = append(args, h.Flags.Upgrade...)
		if h.forceDeploy {
			args = append(args, "--force")
		}
		if r.RecreatePods {
			args = append(args, "--recreate-pods")
		}
	}

	if o.namespace != "" {
		args = append(args, "--namespace", o.namespace)
	}

	// TODO(dgageot): we should merge `Values`, `SetValues` and `SetValueTemplates`
	// as much as possible.

	// Values
	params, err := h.joinTagsToBuildResult(builds, r.Values)
	if err != nil {
		return nil, errors.Wrap(err, "matching build results to chart values")
	}

	// Overrides.Values
	if len(r.Overrides.Values) != 0 {
		args = append(args, "-f", constants.HelmOverridesFilename)
	}

	for k, v := range params {
		var value string

		cfg := r.ImageStrategy.HelmImageConfig.HelmConventionConfig
		value, err = getImageSetValueFromHelmStrategy(cfg, k, v.Tag)
		if err != nil {
			return nil, err
		}

		valuesSet[v.Tag] = true
		args = append(args, "--set-string", value)
	}

	// SetValues
	for k, v := range r.SetValues {
		valuesSet[v] = true
		args = append(args, "--set", fmt.Sprintf("%s=%s", k, v))
	}

	// SetFiles
	args = append(args, generateGetFilesArgs(r.SetFiles, valuesSet)...)

	envMap := map[string]string{}
	for idx, b := range builds {
		suffix := ""
		if idx > 0 {
			suffix = strconv.Itoa(idx + 1)
		}

		for k, v := range createEnvVarMap(b.ImageName, b.Tag) {
			envMap[k+suffix] = v
		}
	}
	logrus.Debugf("EnvVarMap: %#v\n", envMap)

	for k, v := range r.SetValueTemplates {
		v, err := templatedField(v, envMap)
		if err != nil {
			return nil, err
		}
		valuesSet[v] = true
		args = append(args, "--set", fmt.Sprintf("%s=%s", k, v))
	}

	// ValuesFiles
	for _, v := range expandPaths(r.ValuesFiles) {
		v, err := templatedField(v, envMap)
		if err != nil {
			return nil, err
		}
		args = append(args, "-f", v)
	}

	if r.Wait {
		args = append(args, "--wait")
	}

	if r.Version != "" {
		args = append(args, "--version", r.Version)
	}
	args = append(args, r.ChartPath)
	return args, err
}

func templatedField(tmpl string, envMap map[string]string) (string, error) {
	t, err := util.ParseEnvTemplate(tmpl)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse template")
	}
	v, err := util.ExecuteEnvTemplate(t, envMap)
	if err != nil {
		return "", errors.Wrapf(err, "failed to generate template")
	}
	return v, nil
}

func createEnvVarMap(imageName string, fqn string) map[string]string {
	customMap := map[string]string{
		"IMAGE_NAME": imageName,
		"DIGEST":     fqn, // The `DIGEST` name is kept for compatibility reasons
	}
	if fqn != "" {
		// DIGEST_ALGO and DIGEST_HEX are deprecated and will contain non sense values
		names := strings.SplitN(fqn, ":", 2)
		if len(names) >= 2 {
			customMap["DIGEST_ALGO"] = names[0]
			customMap["DIGEST_HEX"] = names[1]
		} else {
			customMap["DIGEST_HEX"] = fqn
		}
	}
	return customMap
}

// packageChart packages the chart and returns path to the chart archive file.
// If this function returns an error, it will always be wrapped.
func (h *HelmDeployer) packageChart(ctx context.Context, r latest.HelmRelease) (string, error) {
	tmp := os.TempDir()
	packageArgs := []string{"package", r.ChartPath, "--destination", tmp}
	if r.Packaged.Version != "" {
		v, err := expandTemplate(r.Packaged.Version)
		if err != nil {
			return "", errors.Wrap(err, `concretize "packaged.version" template`)
		}
		packageArgs = append(packageArgs, "--version", v)
	}
	if r.Packaged.AppVersion != "" {
		av, err := expandTemplate(r.Packaged.AppVersion)
		if err != nil {
			return "", errors.Wrap(err, `concretize "packaged.appVersion" template`)
		}
		packageArgs = append(packageArgs, "--app-version", av)
	}

	buf := &bytes.Buffer{}
	err := h.helm(ctx, buf, false, packageArgs...)
	output := strings.TrimSpace(buf.String())
	if err != nil {
		return "", errors.Wrapf(err, "package chart into a .tgz archive (%s)", output)
	}

	fpath, err := extractChartFilename(output, tmp)
	if err != nil {
		return "", err
	}

	return filepath.Join(tmp, fpath), nil
}

func (h *HelmDeployer) getReleaseInfo(ctx context.Context, release string) (*bufio.Reader, error) {
	var releaseInfo bytes.Buffer
	if err := h.helm(ctx, &releaseInfo, false, "get", release); err != nil {
		return nil, fmt.Errorf("error retrieving helm deployment info: %s", releaseInfo.String())
	}
	return bufio.NewReader(&releaseInfo), nil
}

func getImageSetValueFromHelmStrategy(cfg *latest.HelmConventionConfig, valueName string, tag string) (string, error) {
	if cfg != nil {
		dockerRef, err := docker.ParseReference(tag)
		if err != nil {
			return "", errors.Wrapf(err, "cannot parse the image reference %s", tag)
		}

		var imageTag string
		if dockerRef.Digest != "" {
			imageTag = fmt.Sprintf("%s@%s", dockerRef.Tag, dockerRef.Digest)
		} else {
			imageTag = dockerRef.Tag
		}

		if cfg.ExplicitRegistry {
			if dockerRef.Domain == "" {
				return "", errors.New(fmt.Sprintf("image reference %s has no domain", tag))
			}
			return fmt.Sprintf(
				"%[1]s.registry=%[2]s,%[1]s.repository=%[3]s,%[1]s.tag=%[4]s",
				valueName,
				dockerRef.Domain,
				dockerRef.Path,
				imageTag,
			), nil
		}
		return fmt.Sprintf(
			"%[1]s.repository=%[2]s,%[1]s.tag=%[3]s",
			valueName, dockerRef.BaseName,
			imageTag,
		), nil
	}
	return fmt.Sprintf("%s=%s", valueName, tag), nil
}

// Retrieve info about all releases using helm get
// Skaffold labels will be applied to each deployed k8s object
// Since helm isn't always consistent with retrieving results, don't return errors here
func (h *HelmDeployer) getDeployResults(ctx context.Context, namespace string, release string) []Artifact {
	b, err := h.getReleaseInfo(ctx, release)
	if err != nil {
		logrus.Warnf(err.Error())
		return nil
	}
	return parseReleaseInfo(namespace, b)
}

func (h *HelmDeployer) deleteRelease(ctx context.Context, out io.Writer, r latest.HelmRelease) error {
	hv, err := h.binVer(ctx)
	if err != nil {
		return errors.Wrap(err, "binary version")
	}

	args, err := h.deleteArgs(r.Name, hv)
	if err != nil {
		return errors.Wrap(err, "delete args")
	}

	if err := h.helm(ctx, out, false, args...); err != nil {
		logrus.Debugf("deleting release %s: %v\n", r.Name, err)
	}
	return nil
}

// deleteArgs returns the arguments to be used for deleting a release
func (h *HelmDeployer) deleteArgs(name string, hv semver.Version) ([]string, error) {
	exp, err := expandTemplate(name)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse the release name template")
	}
	args := []string{"delete", exp}
	if hv.LT(semver.MustParse("3.0.0")) {
		args = append(args, "--purge")
	}
	return args, nil
}

func (h *HelmDeployer) joinTagsToBuildResult(builds []build.Artifact, params map[string]string) (map[string]build.Artifact, error) {
	imageToBuildResult := map[string]build.Artifact{}
	for _, b := range builds {
		imageToBuildResult[b.ImageName] = b
	}

	paramToBuildResult := map[string]build.Artifact{}

	for param, imageName := range params {
		b, ok := imageToBuildResult[imageName]
		if !ok {
			return nil, fmt.Errorf("no build present for %s", imageName)
		}

		paramToBuildResult[param] = b
	}

	return paramToBuildResult, nil
}

// binVer returns the version of the helm binary found in PATH. May be cached.
func (h *HelmDeployer) binVer(ctx context.Context) (semver.Version, error) {
	if h.bV.Major != 0 {
		return h.bV, nil
	}

	var b bytes.Buffer
	if err := h.helm(ctx, &b, false, "version"); err != nil {
		return semver.Version{}, errors.Wrap(err, "helm version")
	}
	bs := b.Bytes()
	logrus.Warnf("helm binary version: %s", bs)

	bi := struct {
		Version string `json:"Version"`
	}{}
	if err := json.Unmarshal(bs, &bi); err != nil {
		return semver.Version{}, errors.Wrap(err, "unmarshal")
	}
	logrus.Warnf("struct: %+v", bi)

	v, err := semver.Make(bi.Version)
	if err != nil {
		return semver.Version{}, errors.Wrap(err, "semver make")
	}

	h.bV = v
	logrus.Warnf("set bV to %s", v)
	return h.bV, nil
}

func generateGetFilesArgs(m map[string]string, valuesSet map[string]bool) []string {
	args := make([]string, 0, len(m))
	for k, v := range m {
		valuesSet[v] = true
		args = append(args, "--set-file", fmt.Sprintf("%s=%s", k, v))
	}
	return args
}

func (h *HelmDeployer) Render(context.Context, io.Writer, []build.Artifact, []Labeller, string) error {
	return errors.New("not yet implemented")
}

// expandTemplate parses and executes template s with OS environment variables.
// If s is not a template but a simple string, returns unchanged s.
func expandTemplate(s string) (string, error) {
	tmpl, err := util.ParseEnvTemplate(s)
	if err != nil {
		return "", errors.Wrap(err, "parsing template")
	}

	return util.ExecuteEnvTemplate(tmpl, nil)
}

func extractChartFilename(s, tmp string) (string, error) {
	s = strings.TrimSpace(s)
	idx := strings.Index(s, tmp)
	if idx == -1 {
		return "", errors.New("cannot locate packaged chart archive")
	}

	return s[idx+len(tmp):], nil
}

func expandPaths(paths []string) []string {
	for i, path := range paths {
		expanded, err := homedir.Expand(path)
		if err == nil {
			paths[i] = expanded
		}
	}

	return paths
}
