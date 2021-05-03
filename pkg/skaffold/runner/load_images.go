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

package runner

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/docker/distribution/reference"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/color"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/config"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/graph"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/kubectl"
	kubectx "github.com/GoogleContainerTools/skaffold/pkg/skaffold/kubernetes/context"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/runner/runcontext"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/util"
	"k8s.io/client-go/tools/clientcmd/api"
)

type ImageLoader struct {
	builds     *[]graph.Artifact
	runCtx     *runcontext.RunContext
	kubectlCLI *kubectl.CLI
}

func NewImageLoader(builds *[]graph.Artifact, runCtx *runcontext.RunContext, kubectlCLI *kubectl.CLI) *ImageLoader {
	return &ImageLoader{
		builds: builds, runCtx: runCtx, kubectlCLI: kubectlCLI,
	}
}

// loadImagesInKindNodes loads artifact images into every node of a kind cluster.
func (i *ImageLoader) LoadImagesInKindNodes(ctx context.Context, out io.Writer, kindCluster string, artifacts []graph.Artifact) error {
	color.Default.Fprintln(out, "Loading images into kind cluster nodes...")
	return i.loadImages(ctx, out, artifacts, func(tag string) *exec.Cmd {
		return exec.CommandContext(ctx, "kind", "load", "docker-image", "--name", kindCluster, tag)
	})
}

// loadImagesInK3dNodes loads artifact images into every node of a k3s cluster.
func (i *ImageLoader) LoadImagesInK3dNodes(ctx context.Context, out io.Writer, k3dCluster string, artifacts []graph.Artifact) error {
	color.Default.Fprintln(out, "Loading images into k3d cluster nodes...")
	return i.loadImages(ctx, out, artifacts, func(tag string) *exec.Cmd {
		return exec.CommandContext(ctx, "k3d", "image", "import", "--cluster", k3dCluster, tag)
	})
}

func (i *ImageLoader) loadImages(ctx context.Context, out io.Writer, artifacts []graph.Artifact, createCmd func(tag string) *exec.Cmd) error {
	start := time.Now()

	var knownImages []string

	for _, artifact := range artifacts {
		// Only load images that this runner built
		if !i.wasBuilt(artifact.Tag) {
			continue
		}

		color.Default.Fprintf(out, " - %s -> ", artifact.Tag)

		// Only load images that are unknown to the node
		if knownImages == nil {
			var err error
			if knownImages, err = findKnownImages(ctx, i.kubectlCLI); err != nil {
				return fmt.Errorf("unable to retrieve node's images: %w", err)
			}
		}
		normalizedImageRef, err := reference.ParseNormalizedNamed(artifact.Tag)
		if err != nil {
			return err
		}
		if util.StrSliceContains(knownImages, normalizedImageRef.String()) {
			color.Green.Fprintln(out, "Found")
			continue
		}

		cmd := createCmd(artifact.Tag)
		if output, err := util.RunCmdOut(cmd); err != nil {
			color.Red.Fprintln(out, "Failed")
			return fmt.Errorf("unable to load image %q into cluster: %w, %s", artifact.Tag, err, output)
		}

		color.Green.Fprintln(out, "Loaded")
	}

	color.Default.Fprintln(out, "Images loaded in", util.ShowHumanizeTime(time.Since(start)))
	return nil
}

func findKnownImages(ctx context.Context, cli *kubectl.CLI) ([]string, error) {
	nodeGetOut, err := cli.RunOut(ctx, "get", "nodes", `-ojsonpath='{@.items[*].status.images[*].names[*]}'`)
	if err != nil {
		return nil, fmt.Errorf("unable to inspect the nodes: %w", err)
	}

	knownImages := strings.Split(string(nodeGetOut), " ")
	return knownImages, nil
}

func (i *ImageLoader) wasBuilt(tag string) bool {
	for _, built := range *i.builds {
		if built.Tag == tag {
			return true
		}
	}

	return false
}

func (i *ImageLoader) getCurrentContext() (*api.Context, error) {
	currentCfg, err := kubectx.CurrentConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to get kubernetes config: %w", err)
	}

	currentContext, present := currentCfg.Contexts[i.runCtx.GetKubeContext()]
	if !present {
		return nil, fmt.Errorf("unable to get current kubernetes context: %w", err)
	}
	return currentContext, nil
}

func (i *ImageLoader) LoadImagesIntoCluster(ctx context.Context, out io.Writer, artifacts []graph.Artifact) error {
	currentContext, err := i.getCurrentContext()
	if err != nil {
		return err
	}

	if config.IsKindCluster(i.runCtx.GetKubeContext()) {
		kindCluster := config.KindClusterName(currentContext.Cluster)

		// With `kind`, docker images have to be loaded with the `kind` CLI.
		if err := i.LoadImagesInKindNodes(ctx, out, kindCluster, artifacts); err != nil {
			return fmt.Errorf("loading images into kind nodes: %w", err)
		}
	}

	if config.IsK3dCluster(i.runCtx.GetKubeContext()) {
		k3dCluster := config.K3dClusterName(currentContext.Cluster)

		// With `k3d`, docker images have to be loaded with the `k3d` CLI.
		if err := i.LoadImagesInK3dNodes(ctx, out, k3dCluster, artifacts); err != nil {
			return fmt.Errorf("loading images into k3d nodes: %w", err)
		}
	}

	return nil
}
