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

package jib

import (
	"os/exec"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/v1alpha3"
	"github.com/pkg/errors"
)

// GetDependenciesGradle finds the source dependencies for the given jib-gradle artifact.
// All paths are absolute.
func GetDependenciesGradle(workspace string, a *v1alpha3.JibGradleArtifact) ([]string, error) {
	cmd := getCommandGradle(workspace, a)
	deps, err := getDependencies(cmd)
	if err != nil {
		return nil, errors.Wrapf(err, "getting jib-gradle dependencies")
	}
	return deps, nil
}

func getCommandGradle(workspace string, _ /* a */ *v1alpha3.JibGradleArtifact) *exec.Cmd {
	return getCommand(workspace, "gradle", getWrapperGradle(), []string{"_jibSkaffoldFiles", "-q"})
}
