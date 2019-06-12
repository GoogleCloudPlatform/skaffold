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
	"fmt"
	"path/filepath"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/constants"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
)

// Dockerfile is the path to a dockerfile
type Dockerfile string

// GetPrompt returns the initBuilder's string representation, used when prompting the user to choose a builder.
func (d Dockerfile) GetPrompt() string {
	return fmt.Sprintf("Docker (%s)", d)
}

// GetArtifact returns the Artifact used to generate the Build Config.
func (d Dockerfile) GetArtifact(manifestImage string) *latest.Artifact {
	path := string(d)
	workspace := filepath.Dir(path)
	a := &latest.Artifact{ImageName: manifestImage}
	if workspace != "." {
		a.Workspace = workspace
	}
	if filepath.Base(path) != constants.DefaultDockerfilePath {
		a.ArtifactType = latest.ArtifactType{
			DockerArtifact: &latest.DockerArtifact{DockerfilePath: path},
		}
	}

	return a
}

// GetConfiguredImage returns the target image configured by the builder
func (d Dockerfile) GetConfiguredImage() string {
	// Target image is not configured in dockerfiles
	return ""
}
