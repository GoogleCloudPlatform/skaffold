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

package docker

import (
	"context"
	"io"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/build"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/build/tag"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/config"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/constants"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/docker"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/event"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/plugin/environments/gcb"
	runcontext "github.com/GoogleContainerTools/skaffold/pkg/skaffold/runner/context"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/defaults"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

// Builder builds artifacts with Docker.
type Builder struct {
	opts *config.SkaffoldOptions
	env  *latest.ExecutionEnvironment
	*latest.LocalBuild
	LocalDocker  docker.LocalDaemon
	LocalCluster bool
	PushImages   bool
	// TODO: remove once old docker build functionality is removed (priyawadhwa@)
	PluginMode  bool
	KubeContext string
	builtImages []string
}

// NewBuilder creates a new Builder that builds artifacts with Docker.
func NewBuilder() *Builder {
	return &Builder{
		PluginMode: true,
	}
}

// Init stores skaffold options and the execution environment
func (b *Builder) Init(ctx *runcontext.RunContext) error {
	if b.PluginMode {
		if err := event.SetupRPCClient(ctx.Opts); err != nil {
			logrus.Warn("error establishing gRPC connection to skaffold process; events will not be handled correctly")
			logrus.Warn(err.Error())
			return err
		}
	}
	b.opts = ctx.Opts
	b.env = ctx.Cfg.Build.ExecutionEnvironment
	logrus.Debugf("initialized plugin with %+v", ctx)
	return nil
}

// Labels are labels specific to Docker.
func (b *Builder) Labels() map[string]string {
	return map[string]string{
		constants.Labels.Builder: "docker",
	}
}

// DependenciesForArtifact returns the dependencies for this docker artifact
func (b *Builder) DependenciesForArtifact(ctx context.Context, artifact *latest.Artifact) ([]string, error) {
	if err := setArtifact(artifact); err != nil {
		return nil, err
	}
	paths, err := docker.GetDependencies(ctx, artifact.Workspace, artifact.DockerArtifact.DockerfilePath, artifact.DockerArtifact.BuildArgs)
	if err != nil {
		return nil, errors.Wrapf(err, "getting dependencies for %s", artifact.ImageName)
	}
	return util.AbsolutePaths(artifact.Workspace, paths), nil
}

// Build is responsible for building artifacts in their respective execution environments
// The builder plugin is also responsible for setting any necessary defaults
func (b *Builder) Build(ctx context.Context, out io.Writer, tags tag.ImageTags, artifacts []*latest.Artifact) ([]build.Artifact, error) {
	switch b.env.Name {
	case constants.GoogleCloudBuild:
		return b.googleCloudBuild(ctx, out, tags, artifacts)
	case constants.Local:
		return b.local(ctx, out, tags, artifacts)
	default:
		return nil, errors.Errorf("%s is not a supported environment for builder docker", b.env.Name)
	}
}

func (b *Builder) Prune(ctx context.Context, out io.Writer) error {
	switch b.env.Name {
	case constants.GoogleCloudBuild:
		return nil // noop
	case constants.Local:
		return b.prune(ctx, out)
	default:
		return errors.Errorf("%s is not a supported environment for builder docker", b.env.Name)
	}
}

// googleCloudBuild sets any necessary defaults and then builds artifacts with docker in GCB
func (b *Builder) googleCloudBuild(ctx context.Context, out io.Writer, tags tag.ImageTags, artifacts []*latest.Artifact) ([]build.Artifact, error) {
	var g *latest.GoogleCloudBuild
	if err := util.CloneThroughJSON(b.env.Properties, &g); err != nil {
		return nil, errors.Wrap(err, "converting execution environment to googleCloudBuild struct")
	}
	defaults.SetDefaultCloudBuildDockerImage(g)
	for _, a := range artifacts {
		if err := setArtifact(a); err != nil {
			return nil, err
		}
	}
	runCtx := &runcontext.RunContext{
		Opts: b.opts,
		Cfg: &latest.Pipeline{
			Build: latest.BuildConfig{
				BuildType: latest.BuildType{
					GoogleCloudBuild: g,
				},
			},
		},
	}
	return gcb.NewBuilder(runCtx).Build(ctx, out, tags, artifacts)
}

func setArtifact(artifact *latest.Artifact) error {
	if artifact.ArtifactType.DockerArtifact != nil {
		return nil
	}
	var a *latest.DockerArtifact
	if err := yaml.UnmarshalStrict(artifact.BuilderPlugin.Contents, &a); err != nil {
		return errors.Wrap(err, "unmarshalling docker artifact")
	}
	if a == nil {
		a = &latest.DockerArtifact{}
	}
	defaults.SetDefaultDockerArtifact(a)
	artifact.ArtifactType.DockerArtifact = a
	return nil
}
