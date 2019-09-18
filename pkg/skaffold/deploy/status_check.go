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
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/color"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/deploy/resource"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/kubectl"
	pkgkubernetes "github.com/GoogleContainerTools/skaffold/pkg/skaffold/kubernetes"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/runner/runcontext"
)

var (
	defaultStatusCheckDeadline = time.Duration(10) * time.Minute

	// Poll period for checking set to 100 milliseconds
	defaultPollPeriodInMilliseconds = 100

	// report resource status for pending resources 0.5 second.
	reportStatusTime = 500 * time.Millisecond

	// For testing
	executeRolloutStatus = getRollOutStatus
)

const (
	tabHeader = " -"
)

type counter struct {
	total   int
	pending int32
}

func StatusCheck(ctx context.Context, defaultLabeller *DefaultLabeller, runCtx *runcontext.RunContext, out io.Writer) error {
	client, err := pkgkubernetes.Client()
	if err != nil {
		return errors.Wrap(err, "getting kubernetes client")
	}

	deadline := getDeadline(runCtx.Cfg.Deploy.StatusCheckDeadlineSeconds)
	deployments, err := getDeployments(client, runCtx.Opts.Namespace, defaultLabeller, deadline)
	if err != nil {
		return errors.Wrap(err, "could not fetch deployments")
	}

	wg := sync.WaitGroup{}

	c := newCounter(len(deployments))

	for _, d := range deployments {
		wg.Add(1)
		go func(d *resource.Deployment) {
			defer wg.Done()
			pollDeploymentRolloutStatus(ctx, kubectl.NewFromRunContext(runCtx), d)
			pending := c.markProcessed()
			printStatusCheckSummary(out, d, pending, c.total)
		}(d)
	}

	// Retrieve pending resource states
	go func() {
		printResourceStatus(ctx, out, deployments, deadline)
	}()

	// Wait for all deployment status to be fetched
	wg.Wait()
	return getSkaffoldDeployStatus(deployments)
}

func getDeployments(client kubernetes.Interface, ns string, l *DefaultLabeller, deadlineDuration time.Duration) ([]*resource.Deployment, error) {
	deps, err := client.AppsV1().Deployments(ns).List(metav1.ListOptions{
		LabelSelector: l.RunIDKeyValueString(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not fetch deployments")
	}

	deployments := make([]*resource.Deployment, 0, len(deps.Items))
	for _, d := range deps.Items {
		var deadline time.Duration
		if d.Spec.ProgressDeadlineSeconds == nil || *d.Spec.ProgressDeadlineSeconds > int32(deadlineDuration.Seconds()) {
			deadline = deadlineDuration
		} else {
			deadline = time.Duration(*d.Spec.ProgressDeadlineSeconds) * time.Second
		}
		deployments = append(deployments, resource.NewDeployment(d.Name, d.Namespace, deadline))
	}

	return deployments, nil
}

func pollDeploymentRolloutStatus(ctx context.Context, k *kubectl.CLI, d *resource.Deployment) {
	pollDuration := time.Duration(defaultPollPeriodInMilliseconds) * time.Millisecond
	// Add poll duration to account for one last attempt after progressDeadlineSeconds.
	timeoutContext, cancel := context.WithTimeout(ctx, d.Deadline()+pollDuration)
	logrus.Debugf("checking rollout status %s", d.String())
	defer cancel()
	for {
		select {
		case <-timeoutContext.Done():
			err := errors.Wrap(timeoutContext.Err(), fmt.Sprintf("deployment rollout status could not be fetched within %v", d.Deadline()))
			d.UpdateStatus(err.Error(), err)
			d.MarkDone()
			return
		case <-time.After(pollDuration):
			status, err := executeRolloutStatus(timeoutContext, k, d.Name())
			d.UpdateStatus(status, err)
			if err != nil || strings.Contains(status, "successfully rolled out") {
				d.MarkDone()
				return
			}
		}
	}
}

func getSkaffoldDeployStatus(deployments []*resource.Deployment) error {
	var errorStrings []string
	for _, d := range deployments {
		if err := d.Status().Error(); err != nil {
			errorStrings = append(errorStrings, fmt.Sprintf("deployment %s failed due to %s", d, err.Error()))
		}
	}
	if len(errorStrings) == 0 {
		return nil
	}
	return fmt.Errorf("following deployments are not stable:\n%s", strings.Join(errorStrings, "\n"))
}

func getRollOutStatus(ctx context.Context, k *kubectl.CLI, dName string) (string, error) {
	b, err := k.RunOut(ctx, "rollout", "status", "deployment", dName, "--watch=false")
	return string(b), err
}

func getDeadline(d int) time.Duration {
	if d > 0 {
		return time.Duration(d) * time.Second
	}
	return defaultStatusCheckDeadline
}

func printStatusCheckSummary(out io.Writer, d *resource.Deployment, pending int, total int) {
	status := fmt.Sprintf("%s %s", tabHeader, d)
	if err := d.Status().Error(); err != nil {
		status = fmt.Sprintf("%s failed.%s Error: %s.",
			status,
			trimNewLine(getPendingMessage(pending, total)),
			trimNewLine(err.Error()),
		)
	} else {
		status = fmt.Sprintf("%s is ready.%s", status, getPendingMessage(pending, total))
	}
	color.Default.Fprintln(out, status)
}

// Print resource statuses until all status check are completed or context is cancelled.
func printResourceStatus(ctx context.Context, out io.Writer, deps []*resource.Deployment, deadline time.Duration) {
	timeoutContext, cancel := context.WithTimeout(ctx, deadline)
	defer cancel()
	for {
		var allResourcesCheckComplete bool
		select {
		case <-timeoutContext.Done():
			return
		case <-time.After(reportStatusTime):
			allResourcesCheckComplete = printStatus(deps, out)
		}
		if allResourcesCheckComplete {
			return
		}
	}
}

func printStatus(deps []*resource.Deployment, out io.Writer) bool {
	allResourcesCheckComplete := true
	for _, d := range deps {
		if d.IsDone() {
			continue
		}
		allResourcesCheckComplete = false
		if str := d.ReportSinceLastUpdated(); str != "" {
			color.Default.Fprintln(out, tabHeader, trimNewLine(str))
		}
	}
	return allResourcesCheckComplete
}

func getPendingMessage(pending int, total int) string {
	if pending > 0 {
		return fmt.Sprintf(" [%d/%d deployment(s) still pending]", pending, total)
	}
	return ""
}

func trimNewLine(msg string) string {
	return strings.TrimSuffix(msg, "\n")
}

func newCounter(i int) *counter {
	return &counter{
		total:   i,
		pending: int32(i),
	}
}

func (c *counter) markProcessed() int {
	return int(atomic.AddInt32(&c.pending, -1))
}
