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

// GetDependenciesMaven finds the source dependencies for the given jib-maven artifact.
// All paths are absolute.
// TODO(coollog): Add support for multi-module projects.
func GetDependenciesMaven(workspace string, a *v1alpha3.JibMavenArtifact) ([]string, error) {
	deps, err := getDependencies(getCommandMaven(workspace, a))
	if err != nil {
		return nil, errors.Wrapf(err, "getting jib-maven dependencies")
	}
	return deps, nil
}

func getCommandMaven(workspace string, a *v1alpha3.JibMavenArtifact) *exec.Cmd {
	args := []string{"jib:_skaffold-files", "-q"}
	if a.Profile != "" {
		args = append(args, "-P", a.Profile)
	}

	return getCommand(workspace, "mvn", getWrapperMaven(), args)
}
