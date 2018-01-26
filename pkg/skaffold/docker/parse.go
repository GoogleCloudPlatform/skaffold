/*
Copyright 2018 Google LLC

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
	"io"
	"path"

	"github.com/moby/moby/builder/dockerfile/parser"
	"github.com/pkg/errors"
)

const (
	add  = "add"
	copy = "copy"
)

// GetDockerfileDependencies parses a dockerfile and returns the full paths
// of all the source files that the resulting docker image depends on.
// GetDockerfileDependencies does not expand paths and may contain both a directory and
// files within it.
func GetDockerfileDependencies(workspace string, r io.Reader) ([]string, error) {
	res, err := parser.Parse(r)
	if err != nil {
		return nil, errors.Wrap(err, "parsing dockerfile")
	}
	deps := []string{}
	seen := map[string]struct{}{}
	for _, value := range res.AST.Children {
		switch value.Value {
		case add, copy:
			src := value.Next.Value
			// If flags are present, we are dealing with a multi-stage dockerfile
			// Adding a dependency from a different stage does not imply a source dependency
			if len(value.Flags) != 0 {
				continue
			}
			depPath := path.Join(workspace, src)
			if _, ok := seen[depPath]; ok {
				// If we've already seen this file, only add it once.
				continue
			}
			seen[depPath] = struct{}{}
			deps = append(deps, depPath)
		}
	}
	return deps, nil
}
