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
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/build"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/config"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/docker"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/event"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/runner/runcontext"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
	schemautil "github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/util"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/util"
	"github.com/GoogleContainerTools/skaffold/testutil"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
)

var testBuilds = []build.Artifact{{
	ImageName: "skaffold-helm",
	Tag:       "docker.io:5000/skaffold-helm:3605e7bc17cf46e53f4d81c4cbc24e5b4c495184",
}}

var testBuildsFoo = []build.Artifact{{
	ImageName: "foo",
	Tag:       "foo:3605e7bc17cf46e53f4d81c4cbc24e5b4c495184",
}}

var testDeployConfig = latest.HelmDeploy{
	Releases: []latest.HelmRelease{{
		Name:      "skaffold-helm",
		ChartPath: "examples/test",
		Values: map[string]string{
			"image": "skaffold-helm",
		},
		Overrides: schemautil.HelmOverrides{Values: map[string]interface{}{"foo": "bar"}},
		SetValues: map[string]string{
			"some.key": "somevalue",
		},
	}},
}

var testDeployRecreatePodsConfig = latest.HelmDeploy{
	Releases: []latest.HelmRelease{{
		Name:      "skaffold-helm",
		ChartPath: "examples/test",
		Values: map[string]string{
			"image": "skaffold-helm",
		},
		Overrides: schemautil.HelmOverrides{Values: map[string]interface{}{"foo": "bar"}},
		SetValues: map[string]string{
			"some.key": "somevalue",
		},
		RecreatePods: true,
	}},
}

var testDeploySkipBuildDependenciesConfig = latest.HelmDeploy{
	Releases: []latest.HelmRelease{{
		Name:      "skaffold-helm",
		ChartPath: "examples/test",
		Values: map[string]string{
			"image": "skaffold-helm",
		},
		Overrides: schemautil.HelmOverrides{Values: map[string]interface{}{"foo": "bar"}},
		SetValues: map[string]string{
			"some.key": "somevalue",
		},
		SkipBuildDependencies: true,
	}},
}

var testDeployHelmStyleConfig = latest.HelmDeploy{
	Releases: []latest.HelmRelease{{
		Name:      "skaffold-helm",
		ChartPath: "examples/test",
		Values: map[string]string{
			"image": "skaffold-helm",
		},
		Overrides: schemautil.HelmOverrides{Values: map[string]interface{}{"foo": "bar"}},
		SetValues: map[string]string{
			"some.key": "somevalue",
		},
		ImageStrategy: latest.HelmImageStrategy{
			HelmImageConfig: latest.HelmImageConfig{
				HelmConventionConfig: &latest.HelmConventionConfig{},
			},
		},
	}},
}

var testDeployHelmExplicitRegistryStyleConfig = latest.HelmDeploy{
	Releases: []latest.HelmRelease{
		{
			Name:      "skaffold-helm",
			ChartPath: "examples/test",
			Values: map[string]string{
				"image": "skaffold-helm",
			},
			Overrides: schemautil.HelmOverrides{Values: map[string]interface{}{"foo": "bar"}},
			SetValues: map[string]string{
				"some.key": "somevalue",
			},
			ImageStrategy: latest.HelmImageStrategy{
				HelmImageConfig: latest.HelmImageConfig{
					HelmConventionConfig: &latest.HelmConventionConfig{
						ExplicitRegistry: true,
					},
				},
			},
		},
	},
}

var testDeployConfigParameterUnmatched = latest.HelmDeploy{
	Releases: []latest.HelmRelease{{
		Name:      "skaffold-helm",
		ChartPath: "examples/test",
		Values: map[string]string{
			"image": "skaffold-helm-unmatched",
		}},
	},
}

var testDeployFooWithPackaged = latest.HelmDeploy{
	Releases: []latest.HelmRelease{{
		Name:      "foo",
		ChartPath: "testdata/foo",
		Values: map[string]string{
			"image": "foo",
		},
		Packaged: &latest.HelmPackaged{
			Version:    "0.1.2",
			AppVersion: "1.2.3",
		},
	}},
}

var testDeployWithTemplatedName = latest.HelmDeploy{
	Releases: []latest.HelmRelease{{
		Name:      "{{.USER}}-skaffold-helm",
		ChartPath: "examples/test",
		Values: map[string]string{
			"image.tag": "skaffold-helm",
		},
		Overrides: schemautil.HelmOverrides{Values: map[string]interface{}{"foo": "bar"}},
		SetValues: map[string]string{
			"some.key": "somevalue",
		}},
	},
}

var testDeploySkipBuildDependencies = latest.HelmDeploy{
	Releases: []latest.HelmRelease{{
		Name:                  "skaffold-helm",
		ChartPath:             "stable/chartmuseum",
		SkipBuildDependencies: true,
	}},
}

var testDeployRemoteChart = latest.HelmDeploy{
	Releases: []latest.HelmRelease{{
		Name:                  "skaffold-helm-remote",
		ChartPath:             "stable/chartmuseum",
		SkipBuildDependencies: false,
	}},
}

var testNamespace = "testNamespace"

var validDeployYaml = `
# Source: skaffold-helm/templates/deployment.yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: skaffold-helm
  labels:
    app: skaffold-helm
    chart: skaffold-helm-0.1.0
    release: skaffold-helm
    heritage: Tiller
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: skaffold-helm
        release: skaffold-helm
    spec:
      containers:
        - name: skaffold-helm
          image: gcr.io/nick-cloudbuild/skaffold-helm:f759510436c8fd6f7ffa13dd9e9d85e64bec8d2bfd12c5aa3fb9af1288eccdab
          imagePullPolicy:
          command: ["/bin/bash", "-c", "--" ]
          args: ["while true; do sleep 30; done;"]
          resources:
            {}
`

var validServiceYaml = `
# Source: skaffold-helm/templates/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: skaffold-helm-skaffold-helm
  labels:
    app: skaffold-helm
    chart: skaffold-helm-0.1.0
    release: skaffold-helm
    heritage: Tiller
spec:
  type: ClusterIP
  ports:
    - port: 80
      targetPort: 80
      protocol: TCP
      name: nginx
  selector:
    app: skaffold-helm
    release: skaffold-helm
`

var invalidDeployYaml = `REVISION: 2
RELEASED: Tue Jun 12 15:40:18 2018
CHART: skaffold-helm-0.1.0
USER-SUPPLIED VALUES:
image: gcr.io/nick-cloudbuild/skaffold-helm:f759510436c8fd6f7ffa13dd9e9d85e64bec8d2bfd12c5aa3fb9af1288eccdab

COMPUTED VALUES:
image: gcr.io/nick-cloudbuild/skaffold-helm:f759510436c8fd6f7ffa13dd9e9d85e64bec8d2bfd12c5aa3fb9af1288eccdab
ingress:
  annotations: null
  enabled: false
  hosts:
  - chart-example.local
  tls: null
replicaCount: 1
resources: {}
service:
  externalPort: 80
  internalPort: 80
  name: nginx
  type: ClusterIP

HOOKS:
MANIFEST:
`

// TestMain disables logrus output before running tests.
func TestMain(m *testing.M) {
	logrus.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

func TestHelmDeploy(t *testing.T) {
	tests := []struct {
		description string
		cmd         *MockHelm
		runContext  *runcontext.RunContext
		builds      []build.Artifact
		shouldErr   bool
	}{
		{
			description: "deploy success",
			cmd:         &MockHelm{},
			runContext:  makeRunContext(testDeployConfig, false),
			builds:      testBuilds,
		},
		{
			description: "deploy success with recreatePods",
			cmd:         &MockHelm{},
			runContext:  makeRunContext(testDeployRecreatePodsConfig, false),
			builds:      testBuilds,
		},
		{
			description: "deploy success with skipBuildDependencies",
			cmd:         &MockHelm{},
			runContext:  makeRunContext(testDeploySkipBuildDependenciesConfig, false),
			builds:      testBuilds,
		},
		{
			description: "deploy should not error for unmatched parameter when no builds present",
			cmd:         &MockHelm{},
			runContext:  makeRunContext(testDeployConfigParameterUnmatched, false),
			builds:      nil,
		},
		{
			description: "deploy should error for unmatched parameter when builds present",
			cmd:         &MockHelm{},
			runContext:  makeRunContext(testDeployConfigParameterUnmatched, false),
			builds:      testBuilds,
			shouldErr:   true,
		},
		{
			description: "deploy success remote chart with skipBuildDependencies",
			cmd:         &MockHelm{},
			runContext:  makeRunContext(testDeploySkipBuildDependencies, false),
			builds:      testBuilds,
		},
		{
			description: "deploy error remote chart without skipBuildDependencies",
			cmd: &MockHelm{
				depResult: fmt.Errorf("unexpected error"),
			},
			runContext: makeRunContext(testDeployRemoteChart, false),
			builds:     testBuilds,
			shouldErr:  true,
		},
		{
			description: "get failure should install not upgrade",
			cmd: &MockHelm{
				getResult: fmt.Errorf("not found"),
				installMatcher: func(cmd *exec.Cmd) bool {
					expected := fmt.Sprintf("image=%s", testBuilds[0].Tag)
					return util.StrSliceContains(cmd.Args, expected)
				},
				upgradeResult: fmt.Errorf("should not have called upgrade"),
			},
			runContext: makeRunContext(testDeployConfig, false),
			builds:     testBuilds,
		},
		{
			description: "get failure should install not upgrade with helm image strategy",
			cmd: &MockHelm{
				getResult: fmt.Errorf("not found"),
				installMatcher: func(cmd *exec.Cmd) bool {
					dockerRef, err := docker.ParseReference(testBuilds[0].Tag)
					if err != nil {
						return false
					}

					expected := fmt.Sprintf("image.repository=%s,image.tag=%s", dockerRef.BaseName, dockerRef.Tag)
					return util.StrSliceContains(cmd.Args, expected)
				},
				upgradeResult: fmt.Errorf("should not have called upgrade"),
			},
			runContext: makeRunContext(testDeployHelmStyleConfig, false),
			builds:     testBuilds,
		},
		{
			description: "helm image strategy with explicit registry should set the Helm registry value",
			cmd: &MockHelm{
				getResult: fmt.Errorf("not found"),
				installMatcher: func(cmd *exec.Cmd) bool {
					expected := fmt.Sprintf("image.registry=%s,image.repository=%s,image.tag=%s", "docker.io:5000", "skaffold-helm", "3605e7bc17cf46e53f4d81c4cbc24e5b4c495184")
					return util.StrSliceContains(cmd.Args, expected)
				},
				upgradeResult: fmt.Errorf("should not have called upgrade"),
			},
			runContext: makeRunContext(testDeployHelmExplicitRegistryStyleConfig, false),
			builds:     testBuilds,
		},
		{
			description: "get success should upgrade by force, not install",
			cmd: &MockHelm{
				upgradeMatcher: func(cmd *exec.Cmd) bool {
					return util.StrSliceContains(cmd.Args, "--force")
				},
				installResult: fmt.Errorf("should not have called install"),
			},
			runContext: makeRunContext(testDeployConfig, true),
			builds:     testBuilds,
		},
		{
			description: "get success should upgrade without force, not install",
			cmd: &MockHelm{
				upgradeMatcher: func(cmd *exec.Cmd) bool {
					return !util.StrSliceContains(cmd.Args, "--force")
				},
				installResult: fmt.Errorf("should not have called install"),
			},
			runContext: makeRunContext(testDeployConfig, false),
			builds:     testBuilds,
		},
		{
			description: "deploy error",
			cmd: &MockHelm{
				upgradeResult: fmt.Errorf("unexpected error"),
			},
			shouldErr:  true,
			runContext: makeRunContext(testDeployConfig, false),
			builds:     testBuilds,
		},
		{
			description: "dep build error",
			cmd: &MockHelm{
				depResult: fmt.Errorf("unexpected error"),
			},
			shouldErr:  true,
			runContext: makeRunContext(testDeployConfig, false),
			builds:     testBuilds,
		},
		{
			description: "should package chart and deploy",
			cmd: &MockHelm{
				packageOut: bytes.NewBufferString("Packaged to " + os.TempDir() + "foo-0.1.2.tgz"),
			},
			shouldErr:  false,
			runContext: makeRunContext(testDeployFooWithPackaged, false),
			builds:     testBuildsFoo,
		},
		{
			description: "should fail to deploy when packaging fails",
			cmd: &MockHelm{
				packageResult: fmt.Errorf("packaging failed"),
			},
			shouldErr:  true,
			runContext: makeRunContext(testDeployFooWithPackaged, false),
			builds:     testBuildsFoo,
		},
		{
			description: "deploy and get templated release name",
			cmd:         &MockHelm{},
			runContext:  makeRunContext(testDeployWithTemplatedName, false),
			builds:      testBuilds,
		},
	}
	for _, test := range tests {
		testutil.Run(t, test.description, func(t *testutil.T) {
			t.Override(&util.DefaultExecCommand, test.cmd.ForTest(t))

			event.InitializeState(test.runContext.Cfg.Build)
			err := NewHelmDeployer(test.runContext).Deploy(context.Background(), ioutil.Discard, test.builds, nil).GetError()

			t.CheckError(test.shouldErr, err)
		})
	}
}

type CommandMatcher func(*exec.Cmd) bool

type MockHelm struct {
	t *testutil.T

	getResult      error
	installResult  error
	installMatcher CommandMatcher
	upgradeResult  error
	upgradeMatcher CommandMatcher
	depResult      error

	packageOut    io.Reader
	packageResult error
}

func (m *MockHelm) ForTest(t *testutil.T) *MockHelm {
	m.t = t
	return m
}

func (m *MockHelm) RunCmdOut(c *exec.Cmd) ([]byte, error) {
	m.t.Error("Shouldn't be used")
	return nil, nil
}

func (m *MockHelm) RunCmd(c *exec.Cmd) error {
	if len(c.Args) < 3 {
		m.t.Errorf("Not enough args in command %v", c)
	}

	if c.Args[1] != "--kube-context" || c.Args[2] != testKubeContext {
		m.t.Errorf("Invalid kubernetes context %v", c)
	}

	if c.Args[3] == "get" || c.Args[3] == "upgrade" {
		if releaseName := c.Args[4]; strings.Contains(releaseName, "{{") {
			m.t.Errorf("Invalid release name: %v", releaseName)
		}
	}

	switch c.Args[3] {
	case "get":
		return m.getResult
	case "install":
		if m.upgradeMatcher != nil && !m.installMatcher(c) {
			m.t.Errorf("install matcher failed to match cmd: %+v", c.Args)
		}
		return m.installResult
	case "upgrade":
		if m.upgradeMatcher != nil && !m.upgradeMatcher(c) {
			m.t.Errorf("upgrade matcher failed to match cmd: %+v", c.Args)
		}
		return m.upgradeResult
	case "dep":
		return m.depResult
	case "package":
		if m.packageOut != nil {
			if _, err := io.Copy(c.Stdout, m.packageOut); err != nil {
				m.t.Errorf("Failed to copy stdout")
			}
		}
		return m.packageResult
	default:
		m.t.Errorf("Unknown helm command: %+v", c)
		return nil
	}
}

func TestParseHelmRelease(t *testing.T) {
	tests := []struct {
		description string
		yaml        []byte
		shouldErr   bool
	}{
		{
			description: "parse valid deployment yaml",
			yaml:        []byte(validDeployYaml),
		},
		{
			description: "parse valid service yaml",
			yaml:        []byte(validServiceYaml),
		},
		{
			description: "parse invalid deployment yaml",
			yaml:        []byte(invalidDeployYaml),
			shouldErr:   true,
		},
	}
	for _, test := range tests {
		testutil.Run(t, test.description, func(t *testutil.T) {
			_, err := parseRuntimeObject(testNamespace, test.yaml)

			t.CheckError(test.shouldErr, err)
		})
	}
}

func TestExtractChartFilename(t *testing.T) {
	out, err := extractChartFilename(
		"Successfully packaged chart and saved it to: /var/folders/gm/rrs_712142x8vymmd7xq7h340000gn/T/foo-1.2.3-dirty.tgz\n",
		"/var/folders/gm/rrs_712142x8vymmd7xq7h340000gn/T/",
	)

	testutil.CheckErrorAndDeepEqual(t, false, err, "foo-1.2.3-dirty.tgz", out)
}

func TestHelmDependencies(t *testing.T) {
	tests := []struct {
		description           string
		files                 []string
		valuesFiles           []string
		skipBuildDependencies bool
		remote                bool
		expected              func(folder *testutil.TempDir) []string
	}{
		{
			description:           "charts dir is included when skipBuildDependencies is true",
			files:                 []string{"Chart.yaml", "charts/xyz.tar", "templates/deploy.yaml"},
			skipBuildDependencies: true,
			expected: func(folder *testutil.TempDir) []string {
				return []string{folder.Path("Chart.yaml"), folder.Path("charts/xyz.tar"), folder.Path("templates/deploy.yaml")}
			},
		},
		{
			description:           "charts dir is excluded when skipBuildDependencies is false",
			files:                 []string{"Chart.yaml", "charts/xyz.tar", "templates/deploy.yaml"},
			skipBuildDependencies: false,
			expected: func(folder *testutil.TempDir) []string {
				return []string{folder.Path("Chart.yaml"), folder.Path("templates/deploy.yaml")}
			},
		},
		{
			description:           "values file is included",
			skipBuildDependencies: false,
			files:                 []string{"Chart.yaml"},
			valuesFiles:           []string{"/folder/values.yaml"},
			expected: func(folder *testutil.TempDir) []string {
				return []string{"/folder/values.yaml", folder.Path("Chart.yaml")}
			},
		},
		{
			description:           "no deps for remote chart path",
			skipBuildDependencies: false,
			files:                 []string{"Chart.yaml"},
			remote:                true,
			expected: func(folder *testutil.TempDir) []string {
				return nil
			},
		},
	}
	for _, test := range tests {
		testutil.Run(t, test.description, func(t *testutil.T) {
			tmpDir := t.NewTempDir().
				Touch(test.files...)

			deployer := NewHelmDeployer(makeRunContext(latest.HelmDeploy{
				Releases: []latest.HelmRelease{{
					Name:                  "skaffold-helm",
					ChartPath:             tmpDir.Root(),
					ValuesFiles:           test.valuesFiles,
					Values:                map[string]string{"image": "skaffold-helm"},
					Overrides:             schemautil.HelmOverrides{Values: map[string]interface{}{"foo": "bar"}},
					SetValues:             map[string]string{"some.key": "somevalue"},
					SkipBuildDependencies: test.skipBuildDependencies,
					Remote:                test.remote,
				},
				},
			}, false))

			deps, err := deployer.Dependencies()

			t.CheckNoError(err)
			t.CheckDeepEqual(test.expected(tmpDir), deps)
		})
	}
}

func TestExpandPaths(t *testing.T) {
	homedir.DisableCache = true // for testing only

	tests := []struct {
		description  string
		paths        []string
		unixExpanded []string // unix expands path with forward slashes, windows with backward slashes
		winExpanded  []string
		env          map[string]string
	}{
		{
			description:  "expand paths on unix",
			paths:        []string{"~/path/with/tilde/values.yaml", "/some/absolute/path/values.yaml"},
			unixExpanded: []string{"/home/path/with/tilde/values.yaml", "/some/absolute/path/values.yaml"},
			winExpanded:  []string{`\home\path\with\tilde\values.yaml`, "/some/absolute/path/values.yaml"},
			env:          map[string]string{"HOME": "/home"},
		},
		{
			description:  "expand paths on windows",
			paths:        []string{"~/path/with/tilde/values.yaml", `C:\Users\SomeUser\path\values.yaml`},
			unixExpanded: []string{`C:\Users\SomeUser/path/with/tilde/values.yaml`, `C:\Users\SomeUser\path\values.yaml`},
			winExpanded:  []string{`C:\Users\SomeUser\path\with\tilde\values.yaml`, `C:\Users\SomeUser\path\values.yaml`},
			env:          map[string]string{"HOME": `C:\Users\SomeUser`},
		},
	}

	for _, test := range tests {
		testutil.Run(t, test.description, func(t *testutil.T) {
			t.SetEnvs(test.env)
			expanded := expandPaths(test.paths)

			if runtime.GOOS == "windows" {
				t.CheckDeepEqual(test.winExpanded, expanded)
			} else {
				t.CheckDeepEqual(test.unixExpanded, expanded)
			}
		})
	}
}

func makeRunContext(deploy latest.HelmDeploy, force bool) *runcontext.RunContext {
	pipeline := latest.Pipeline{}
	pipeline.Deploy.DeployType.HelmDeploy = &deploy

	return &runcontext.RunContext{
		Cfg:         pipeline,
		KubeContext: testKubeContext,
		Opts: config.SkaffoldOptions{
			Namespace: testNamespace,
			Force:     force,
		},
	}
}
