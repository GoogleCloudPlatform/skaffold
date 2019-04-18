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
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/build"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/build/cache"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/build/kaniko"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/build/local"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/build/tag"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/color"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/config"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/constants"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/deploy"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/env"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/event"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/kubernetes"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/plugins/builders"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/plugins/environments"
	runcontext "github.com/GoogleContainerTools/skaffold/pkg/skaffold/runner/context"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/sync"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/sync/kubectl"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/test"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/version"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/watch"
)

// SkaffoldRunner is responsible for running the skaffold build and deploy config.
type SkaffoldRunner struct {
	build.Builder
	env.EnvBuilder
	deploy.Deployer
	test.Tester
	tag.Tagger
	sync.Syncer
	watch.Watcher

	cache             *cache.Cache
	runCtx            *runcontext.RunContext
	labellers         []deploy.Labeller
	builds            []build.Artifact
	hasBuilt          bool
	hasDeployed       bool
	imageList         *kubernetes.ImageList
	RPCServerShutdown func() error
}

// NewForConfig returns a new SkaffoldRunner for a SkaffoldConfig
func NewForConfig(opts *config.SkaffoldOptions, cfg *latest.SkaffoldConfig) (*SkaffoldRunner, error) {
	runCtx, err := runcontext.GetRunContext(opts, &cfg.Pipeline)
	if err != nil {
		return nil, errors.Wrap(err, "getting run context")
	}

	tagger, err := getTagger(cfg.Build.TagPolicy, opts.CustomTag)
	if err != nil {
		return nil, errors.Wrap(err, "parsing tag config")
	}

	// Env Plugin Flow only if executionEnv set to googleCloudPlugin
	executor, builder, err := executeEnvPluginFlow(runCtx)
	if err != nil {
		return nil, errors.Wrap(err, "could not execute EnvPluin flow")
	}

	artifactCache := cache.NewCache(builder, runCtx)
	tester, err := getTester(runCtx)
	if err != nil {
		return nil, errors.Wrap(err, "parsing test config")
	}

	deployer, err := getDeployer(runCtx)
	if err != nil {
		return nil, errors.Wrap(err, "parsing deploy config")
	}

	labellers := []deploy.Labeller{opts, builder, deployer, tagger}

	builder, tester, deployer = WithTimings(builder, tester, deployer, opts.CacheArtifacts)
	if opts.Notification {
		deployer = WithNotification(deployer)
	}

	trigger, err := watch.NewTrigger(opts)
	if err != nil {
		return nil, errors.Wrap(err, "creating watch trigger")
	}

	shutdown, err := event.InitializeState(runCtx)
	if err != nil {
		return nil, errors.Wrap(err, "initializing skaffold event handler")
	}

	event.LogSkaffoldMetadata(version.Get())
	return &SkaffoldRunner{
		Builder:           builder,
		EnvBuilder:        executor,
		Tester:            tester,
		Deployer:          deployer,
		Tagger:            tagger,
		Syncer:            kubectl.NewSyncer(runCtx.Namespaces),
		Watcher:           watch.NewWatcher(trigger),
		labellers:         labellers,
		imageList:         kubernetes.NewImageList(),
		cache:             artifactCache,
		runCtx:            runCtx,
		RPCServerShutdown: shutdown,
	}, nil
}

func executeEnvPluginFlow(runCtx *runcontext.RunContext) (env.EnvBuilder, build.Builder, error) {
	executor, err := getEnv(runCtx)
	if err != nil {
		return nil, nil, errors.Wrap(err, "execution env not supported")
	}

	var builder build.Builder
	if executor == nil {
		builder, err = getBuilder(runCtx)
		if err != nil {
			return nil, nil, errors.Wrap(err, "parsing build config")
		}
	}
	return executor, builder, err
}

func getEnv(runCtx *runcontext.RunContext) (env.EnvBuilder, error) {
	switch {
	case runCtx.Cfg.Build.GoogleCloudBuild != nil:
		return environments.RegisteredEnvPlugins(runCtx)
	case runCtx.Cfg.Build.ExecutionEnvironment.Name == constants.GoogleCloudBuild:
		return environments.RegisteredEnvPlugins(runCtx)
	}
	return nil, nil
}

func getBuilder(runCtx *runcontext.RunContext) (build.Builder, error) {
	switch {
	case runCtx.Plugin:
		logrus.Debugln("Using builder plugins")
		return builders.RegisteredBuilderPlugins(runCtx)

	case len(runCtx.Opts.PreBuiltImages) > 0:
		logrus.Debugln("Using pre-built images")
		return build.NewPreBuiltImagesBuilder(runCtx), nil

	case runCtx.Cfg.Build.LocalBuild != nil:
		logrus.Debugln("Using builder: local")
		return local.NewBuilder(runCtx)

	case runCtx.Cfg.Build.GoogleCloudBuild != nil:
		logrus.Debugln("Using builder: google cloud.")
		return builders.RegisteredBuilderPlugins(runCtx)

	case runCtx.Cfg.Build.Cluster != nil:
		logrus.Debugln("Using builder: kaniko")
		return kaniko.NewBuilder(runCtx)

	default:
		return nil, fmt.Errorf("unknown builder for config %+v", runCtx.Cfg.Build)
	}
}

func getTester(runCtx *runcontext.RunContext) (test.Tester, error) {
	return test.NewTester(runCtx)
}

func getDeployer(runCtx *runcontext.RunContext) (deploy.Deployer, error) {
	switch {
	case runCtx.Cfg.Deploy.HelmDeploy != nil:
		return deploy.NewHelmDeployer(runCtx), nil

	case runCtx.Cfg.Deploy.KubectlDeploy != nil:
		return deploy.NewKubectlDeployer(runCtx), nil

	case runCtx.Cfg.Deploy.KustomizeDeploy != nil:
		return deploy.NewKustomizeDeployer(runCtx), nil

	default:
		return nil, fmt.Errorf("unknown deployer for config %+v", runCtx.Cfg.Deploy)
	}
}

func getTagger(t latest.TagPolicy, customTag string) (tag.Tagger, error) {
	switch {
	case customTag != "":
		return &tag.CustomTag{
			Tag: customTag,
		}, nil

	case t.EnvTemplateTagger != nil:
		return tag.NewEnvTemplateTagger(t.EnvTemplateTagger.Template)

	case t.ShaTagger != nil:
		return &tag.ChecksumTagger{}, nil

	case t.GitTagger != nil:
		return &tag.GitCommit{}, nil

	case t.DateTimeTagger != nil:
		return tag.NewDateTimeTagger(t.DateTimeTagger.Format, t.DateTimeTagger.TimeZone), nil

	default:
		return nil, fmt.Errorf("unknown tagger for strategy %+v", t)
	}
}

func (r *SkaffoldRunner) newLogger(out io.Writer, artifacts []*latest.Artifact) *kubernetes.LogAggregator {
	var imageNames []string
	for _, artifact := range artifacts {
		imageNames = append(imageNames, artifact.ImageName)
	}

	return kubernetes.NewLogAggregator(out, imageNames, r.imageList, r.runCtx.Namespaces)
}

// HasDeployed returns true if this runner has deployed something.
func (r *SkaffoldRunner) HasDeployed() bool {
	return r.hasDeployed
}

// HasBuilt returns true if this runner has built something.
func (r *SkaffoldRunner) HasBuilt() bool {
	return r.hasBuilt
}

func (r *SkaffoldRunner) buildTestDeploy(ctx context.Context, out io.Writer, artifacts []*latest.Artifact) error {
	bRes, err := r.BuildAndTest(ctx, out, artifacts)
	if err != nil {
		return err
	}

	// Update which images are logged.
	for _, build := range bRes {
		r.imageList.Add(build.Tag)
	}

	// Make sure all artifacts are redeployed. Not only those that were just built.
	r.builds = build.MergeWithPreviousBuilds(bRes, r.builds)

	if err := r.Deploy(ctx, out, r.builds); err != nil {
		return errors.Wrap(err, "deploy failed")
	}

	return nil
}

// Run builds artifacts, runs tests on built artifacts, and then deploys them.
func (r *SkaffoldRunner) Run(ctx context.Context, out io.Writer, artifacts []*latest.Artifact) error {
	if err := r.buildTestDeploy(ctx, out, artifacts); err != nil {
		return err
	}

	if r.runCtx.Opts.Tail {
		logger := r.newLogger(out, artifacts)
		if err := logger.Start(ctx); err != nil {
			return errors.Wrap(err, "starting logger")
		}
		<-ctx.Done()
	}

	return nil
}

type tagErr struct {
	tag string
	err error
}

// imageTags generates tags for a list of artifacts
func (r *SkaffoldRunner) imageTags(ctx context.Context, out io.Writer, artifacts []*latest.Artifact) (tag.ImageTags, error) {
	start := time.Now()
	color.Default.Fprintln(out, "Generating tags...")

	tagErrs := make([]chan tagErr, len(artifacts))

	for i := range artifacts {
		tagErrs[i] = make(chan tagErr, 1)

		i := i
		go func() {
			tag, err := r.Tagger.GenerateFullyQualifiedImageName(artifacts[i].Workspace, artifacts[i].ImageName)
			tagErrs[i] <- tagErr{tag: tag, err: err}
		}()
	}

	imageTags := make(tag.ImageTags, len(artifacts))

	for i, artifact := range artifacts {
		imageName := artifact.ImageName
		color.Default.Fprintf(out, " - %s -> ", imageName)

		select {
		case <-ctx.Done():
			return nil, context.Canceled

		case t := <-tagErrs[i]:
			tag := t.tag
			err := t.err
			if err != nil {
				return nil, errors.Wrapf(err, "generating tag for %s", imageName)
			}

			fmt.Fprintln(out, tag)

			imageTags[imageName] = tag
		}
	}

	color.Default.Fprintln(out, "Tags generated in", time.Since(start))
	return imageTags, nil
}

// BuildAndTest builds artifacts and runs tests on built artifacts
func (r *SkaffoldRunner) BuildAndTest(ctx context.Context, out io.Writer, artifacts []*latest.Artifact) ([]build.Artifact, error) {
	tags, err := r.imageTags(ctx, out, artifacts)
	if err != nil {
		return nil, errors.Wrap(err, "generating tag")
	}
	r.hasBuilt = true

	artifactsToBuild, res, err := r.cache.RetrieveCachedArtifacts(ctx, out, artifacts)
	if err != nil {
		return nil, errors.Wrap(err, "retrieving cached artifacts")
	}

	// if EnvPlugin detcted, we need to build in Execution env. Since this only exists for `googlecloudbuild`,
	// we hard code this check here where envBuilder is not nil
	// this should go away in real implementation.
	fmt.Println("===================================================")
	fmt.Println(r.EnvBuilder)
	if r.EnvBuilder != nil {
		bRes := []build.Artifact{}
		logrus.Debug("=====================In Env Plugin Flow\n")
		// Use EnvBuilder Plugin
		for _, a := range artifacts {
			// get builder Plugin impl for this artifact
			builder, err := builders.InitBuilderPluginForArtifact(r.runCtx, a)
			if err != nil {
				return bRes, err
			}
			desc, _ := builder.BuildDescription(tags, a)
			a, err := r.EnvBuilder.ExecuteArtifactBuild(ctx, out, tags[a.ImageName], a, *desc)
			if err != nil {
				return bRes, err
			}
			bRes = append(bRes, a)
			fmt.Println(a.ImageName, builder)
		}
		return bRes, nil
	}
	logrus.Debug("=====================Not In Env Plugin Flow\n")
	bRes, err := r.Build(ctx, out, tags, artifactsToBuild)
	if err != nil {
		return nil, errors.Wrap(err, "build failed")
	}
	r.cache.RetagLocalImages(ctx, out, artifactsToBuild, bRes)
	bRes = append(bRes, res...)
	if err := r.cache.CacheArtifacts(ctx, artifacts, bRes); err != nil {
		logrus.Warnf("error caching artifacts: %v", err)
	}
	if !r.runCtx.Opts.SkipTests {
		if err = r.Test(ctx, out, bRes); err != nil {
			return nil, errors.Wrap(err, "test failed")
		}
	}
	return bRes, err

}

// Deploy deploys the given artifacts
func (r *SkaffoldRunner) Deploy(ctx context.Context, out io.Writer, artifacts []build.Artifact) error {
	err := r.Deployer.Deploy(ctx, out, artifacts, r.labellers)
	r.hasDeployed = true
	return err
}

// TailLogs prints the logs for deployed artifacts.
func (r *SkaffoldRunner) TailLogs(ctx context.Context, out io.Writer, artifacts []*latest.Artifact, bRes []build.Artifact) error {
	if !r.runCtx.Opts.Tail {
		return nil
	}

	for _, b := range bRes {
		r.imageList.Add(b.Tag)
	}

	logger := r.newLogger(out, artifacts)
	if err := logger.Start(ctx); err != nil {
		return errors.Wrap(err, "starting logger")
	}

	<-ctx.Done()
	return nil
}
