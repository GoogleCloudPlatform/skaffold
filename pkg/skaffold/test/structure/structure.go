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

package structure

import (
	"context"
	"io"
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/build"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/util"
)

// Test is the entrypoint for running structure tests
func (tr *Runner) Test(ctx context.Context, out io.Writer, imageName string, bRes []build.Artifact) error {
	fqn, err := getImage(ctx, out, imageName, bRes, tr.localDaemon, tr.imagesAreLocal)
	if err != nil {
		return err
	}
	if fqn == "" {
		return nil
	}

	files, err := tr.TestDependencies()
	if err != nil {
		return err
	}

	logrus.Infof("Running structure tests for files %v", files)

	args := []string{"test", "-v", "warn", "--image", fqn}
	for _, f := range files {
		args = append(args, "--config", f)
	}

	cmd := exec.CommandContext(ctx, "container-structure-test", args...)
	cmd.Stdout = out
	cmd.Stderr = out
	cmd.Env = tr.env()

	if err := util.RunCmd(cmd); err != nil {
		return containerStructureTestErr(err)
	}

	return nil
}

// env returns a merged environment of the current process environment and any extra environment.
// This ensures that the correct docker environment configuration is passed to container-structure-test,
// for example when running on minikube.
func (tr *Runner) env() []string {
	extraEnv := tr.localDaemon.ExtraEnv()
	if extraEnv == nil {
		return nil
	}

	parentEnv := os.Environ()
	mergedEnv := make([]string, len(parentEnv), len(parentEnv)+len(extraEnv))
	copy(mergedEnv, parentEnv)
	return append(mergedEnv, extraEnv...)
}
