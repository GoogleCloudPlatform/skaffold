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
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/util"
)

func CreateDockerTarContext(ctx context.Context, w io.Writer, workspace string, a *latest.DockerArtifact, cfg Config) error {
	paths, err := GetDependenciesCached(ctx, workspace, a.DockerfilePath, a.BuildArgs, cfg)
	if err != nil {
		return fmt.Errorf("getting relative tar paths: %w", err)
	}

	var p []string
	for _, path := range paths {
		p = append(p, filepath.Join(workspace, path))
	}

	if err := util.CreateTar(w, workspace, p); err != nil {
		return fmt.Errorf("creating tar gz: %w", err)
	}

	return nil
}
