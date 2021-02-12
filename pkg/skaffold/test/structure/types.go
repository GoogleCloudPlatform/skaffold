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

import "github.com/GoogleContainerTools/skaffold/pkg/skaffold/docker"

type Runner struct {
	structureTests []string
	testWorkingDir string
	localDaemon    docker.LocalDaemon
	imagesAreLocal func(imageName string) (bool, error)
}

// NewRunner creates a new structure.Runner.
func NewRunner(tc []string, workingDir string, localDaemon docker.LocalDaemon,
	imagesAreLocal func(imageName string) (bool, error)) *Runner {
	return &Runner{
		structureTests: tc,
		testWorkingDir: workingDir,
		localDaemon:    localDaemon,
		imagesAreLocal: imagesAreLocal,
	}
}
