/*
Copyright 2018 The Skaffold Authors

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
	"context"
	"io"
	"os/exec"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/build"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/constants"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/deploy/kubectl"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/v1alpha2"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/util"
	"github.com/pkg/errors"
)

type KustomizeDeployer struct {
	*v1alpha2.KustomizeDeploy

	kubectl kubectl.CLI
}

func NewKustomizeDeployer(cfg *v1alpha2.KustomizeDeploy, kubeContext string, namespace string) *KustomizeDeployer {
	return &KustomizeDeployer{
		KustomizeDeploy: cfg,
		kubectl: kubectl.CLI{
			Namespace:   namespace,
			KubeContext: kubeContext,
			Flags:       cfg.Flags,
		},
	}
}

func (k *KustomizeDeployer) Labels() map[string]string {
	return map[string]string{
		constants.Labels.Deployer: "kustomize",
	}
}

func (k *KustomizeDeployer) Deploy(ctx context.Context, out io.Writer, builds []build.Artifact) ([]Artifact, error) {
	manifests, err := k.readManifests(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "reading manifests")
	}

	if len(manifests) == 0 {
		return nil, nil
	}

	manifests, err = manifests.ReplaceImages(builds)
	if err != nil {
		return nil, errors.Wrap(err, "replacing images in manifests")
	}

	updated, err := k.kubectl.Apply(ctx, out, manifests)
	if err != nil {
		return nil, errors.Wrap(err, "apply")
	}

	return parseManifestsForDeploys(updated)
}

func (k *KustomizeDeployer) Cleanup(ctx context.Context, out io.Writer) error {
	manifests, err := k.readManifests(ctx)
	if err != nil {
		return errors.Wrap(err, "reading manifests")
	}

	if err := k.kubectl.Delete(ctx, out, manifests); err != nil {
		return errors.Wrap(err, "delete")
	}

	return nil
}

func (k *KustomizeDeployer) Dependencies() ([]string, error) {
	// TODO(r2d4): parse kustomization yaml and add base and patches as dependencies
	return []string{k.KustomizePath}, nil
}

func (k *KustomizeDeployer) readManifests(ctx context.Context) (kubectl.ManifestList, error) {
	cmd := exec.CommandContext(ctx, "kustomize", "build", k.KustomizePath)
	out, err := util.RunCmdOut(cmd)
	if err != nil {
		return nil, errors.Wrap(err, "kustomize build")
	}

	var manifests kubectl.ManifestList
	manifests.Append(out)
	return manifests, nil
}
