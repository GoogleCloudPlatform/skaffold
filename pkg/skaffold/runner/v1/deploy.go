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

package v1

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/constants"
	deployutil "github.com/GoogleContainerTools/skaffold/pkg/skaffold/deploy/util"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/event"
	eventV2 "github.com/GoogleContainerTools/skaffold/pkg/skaffold/event/v2"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/graph"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/instrumentation"
	kubernetesclient "github.com/GoogleContainerTools/skaffold/pkg/skaffold/kubernetes/client"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/output"
)

// DeployAndLog deploys a list of already built artifacts and optionally show the logs.
func (r *SkaffoldRunner) DeployAndLog(ctx context.Context, out io.Writer, artifacts []graph.Artifact) error {
	defer r.deployer.GetLogger().Stop()

	// Logs should be retrieved up to just before the deploy
	r.deployer.GetLogger().SetSince(time.Now())
	// First deploy
	if err := r.Deploy(ctx, out, artifacts); err != nil {
		return err
	}

	defer r.deployer.GetAccessor().Stop()

	if err := r.deployer.GetAccessor().Start(ctx, out, r.runCtx.GetNamespaces()); err != nil {
		logrus.Warnln("Error starting port forwarding:", err)
	}

	// Start printing the logs after deploy is finished
	if err := r.deployer.GetLogger().Start(ctx, out, r.runCtx.GetNamespaces()); err != nil {
		return fmt.Errorf("starting logger: %w", err)
	}

	if r.runCtx.Tail() || r.runCtx.PortForward() {
		output.Yellow.Fprintln(out, "Press Ctrl+C to exit")
		<-ctx.Done()
	}

	return nil
}

func (r *SkaffoldRunner) Deploy(ctx context.Context, out io.Writer, artifacts []graph.Artifact) error {
	if r.runCtx.RenderOnly() {
		return r.Render(ctx, out, artifacts, false, r.runCtx.RenderOutput())
	}
	defer r.deployer.GetStatusMonitor().Reset()

	out = output.WithEventContext(out, constants.Deploy, eventV2.SubtaskIDNone, "skaffold")

	output.Default.Fprintln(out, "Tags used in deployment:")

	for _, artifact := range artifacts {
		output.Default.Fprintf(out, " - %s -> ", artifact.ImageName)
		fmt.Fprintln(out, artifact.Tag)
	}

	var localImages []graph.Artifact
	for _, a := range artifacts {
		if isLocal, err := r.isLocalImage(a.ImageName); err != nil {
			return err
		} else if isLocal {
			localImages = append(localImages, a)
		}
	}

	if len(localImages) > 0 {
		logrus.Debugln(`Local images can't be referenced by digest.
They are tagged and referenced by a unique, local only, tag instead.
See https://skaffold.dev/docs/pipeline-stages/taggers/#how-tagging-works`)
	}

	// Check that the cluster is reachable.
	// This gives a better error message when the cluster can't
	// be reached.
	if err := failIfClusterIsNotReachable(); err != nil {
		return fmt.Errorf("unable to connect to Kubernetes: %w", err)
	}

	if err := r.deployer.GetImageLoader().LoadImages(ctx, out, localImages); err != nil {
		return err
	}

	deployOut, postDeployFn, err := deployutil.WithLogFile(time.Now().Format(deployutil.TimeFormat)+".log", out, r.runCtx.Muted())
	if err != nil {
		return err
	}

	event.DeployInProgress()
	eventV2.TaskInProgress(constants.Deploy, "Deploy to cluster")
	ctx, endTrace := instrumentation.StartTrace(ctx, "Deploy_Deploying")
	defer endTrace()

	namespaces, err := r.deployer.Deploy(ctx, deployOut, artifacts)
	postDeployFn()
	if err != nil {
		event.DeployFailed(err)
		eventV2.TaskFailed(constants.Deploy, err)
		endTrace(instrumentation.TraceEndError(err))
		return err
	}

	r.hasDeployed = true

	statusCheckOut, postStatusCheckFn, err := deployutil.WithStatusCheckLogFile(time.Now().Format(deployutil.TimeFormat)+".log", out, r.runCtx.Muted())
	defer postStatusCheckFn()
	if err != nil {
		endTrace(instrumentation.TraceEndError(err))
		return err
	}

	event.DeployComplete()
	eventV2.TaskSucceeded(constants.Deploy)
	r.runCtx.UpdateNamespaces(namespaces)
	if !r.runCtx.Opts.IterativeStatusCheck {
		// run final aggregated status check only if iterative status check is turned off.
		return r.deployer.GetStatusMonitor().Check(ctx, statusCheckOut)
	}
	return nil
}

// failIfClusterIsNotReachable checks that Kubernetes is reachable.
// This gives a clear early error when the cluster can't be reached.
func failIfClusterIsNotReachable() error {
	client, err := kubernetesclient.Client()
	if err != nil {
		return err
	}

	_, err = client.Discovery().ServerVersion()
	return err
}
