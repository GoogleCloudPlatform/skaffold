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
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
)

// SyncRules creates a map of sync rules by looking at the COPY/ADD commands in the Dockerfile.
// All keys are relative to the artifact's workspace, the destinations are absolute container paths.
// TODO(corneliusweig) destinations are not resolved across stages in multistage dockerfiles. Is there a use-case for that?
func SyncRules(workspace string, dockerfilePath string, buildArgs map[string]*string, insecureRegistries map[string]bool) ([]*latest.SyncRule, error) {
	absDockerfilePath, err := NormalizeDockerfilePath(workspace, dockerfilePath)
	if err != nil {
		return nil, errors.Wrap(err, "normalizing dockerfile path")
	}

	// only the COPY/ADD commands from the last image are syncable
	copyCommands, err := readCopyCmdsFromDockerfile(true, absDockerfilePath, buildArgs, insecureRegistries)
	if err != nil {
		return nil, err
	}

	var syncRules []*latest.SyncRule

	for _, copyCommand := range copyCommands {
		for _, copySrc := range copyCommand.srcs {
			var src, strip string

			if strings.Contains(copySrc, "*") {
				// Glob
				src = copySrc
				strip = path.Dir(copySrc)
			} else {
				fi, err := os.Stat(filepath.Join(workspace, copySrc))
				if err != nil && !os.IsNotExist(err) {
					return nil, err
				}

				if os.IsNotExist(err) || fi.Mode().IsRegular() {
					// File
					src = copySrc
					strip = path.Dir(copySrc)
				} else {
					// Directory
					src = path.Join(copySrc, "**")
					strip = copySrc
				}
			}

			if strip == "." {
				strip = ""
			}

			syncRules = append(syncRules, &latest.SyncRule{
				Src:   src,
				Dest:  copyCommand.dest,
				Strip: strip,
			})
		}
	}

	return syncRules, nil
}
