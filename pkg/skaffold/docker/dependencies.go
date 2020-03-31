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
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/docker/docker/builder/dockerignore"
	"github.com/docker/docker/pkg/fileutils"
	"github.com/karrick/godirwalk"
)

// NormalizeDockerfilePath returns the absolute path to the dockerfile.
func NormalizeDockerfilePath(context, dockerfile string) (string, error) {
	if filepath.IsAbs(dockerfile) {
		return dockerfile, nil
	}

	if !strings.HasPrefix(dockerfile, context) {
		dockerfile = filepath.Join(context, dockerfile)
	}
	return filepath.Abs(dockerfile)
}

// GetDependencies finds the sources dependencies for the given docker artifact.
// All paths are relative to the workspace.
func GetDependencies(ctx context.Context, workspace string, dockerfilePath string, buildArgs map[string]*string, insecureRegistries map[string]bool) ([]string, error) {
	absDockerfilePath, err := NormalizeDockerfilePath(workspace, dockerfilePath)
	if err != nil {
		return nil, fmt.Errorf("normalizing dockerfile path: %w", err)
	}

	fts, err := readCopyCmdsFromDockerfile(false, absDockerfilePath, workspace, buildArgs, insecureRegistries)
	if err != nil {
		return nil, err
	}

	excludes, err := readDockerignore(workspace, absDockerfilePath)
	if err != nil {
		return nil, fmt.Errorf("reading .dockerignore: %w", err)
	}

	deps := make([]string, 0, len(fts))
	for _, ft := range fts {
		deps = append(deps, ft.from)
	}

	files, err := WalkWorkspace(workspace, excludes, deps)
	if err != nil {
		return nil, fmt.Errorf("walking workspace: %w", err)
	}

	// Always add dockerfile even if it's .dockerignored. The daemon will need it anyways.
	if !filepath.IsAbs(dockerfilePath) {
		files[dockerfilePath] = true
	} else {
		files[absDockerfilePath] = true
	}

	// Ignore .dockerignore
	delete(files, ".dockerignore")

	var dependencies []string
	for file := range files {
		dependencies = append(dependencies, file)
	}
	sort.Strings(dependencies)

	return dependencies, nil
}

// readDockerignore reads patterns to ignore
func readDockerignore(workspace string, absDockerfilePath string) ([]string, error) {
	var excludes []string
	dockerignorePaths := []string{
		absDockerfilePath + ".dockerignore",
		filepath.Join(workspace, ".dockerignore"),
	}
	for _, dockerignorePath := range dockerignorePaths {
		if _, err := os.Stat(dockerignorePath); !os.IsNotExist(err) {
			r, err := os.Open(dockerignorePath)
			if err != nil {
				return nil, err
			}
			defer r.Close()

			excludes, err = dockerignore.ReadAll(r)
			if err != nil {
				return nil, err
			}
			return excludes, nil
		}
	}
	return nil, nil
}

// WalkWorkspace walks the given host directories and records all files found.
// Note: if you change this function, you might also want to modify `walkWorkspaceWithDestinations`.
func WalkWorkspace(workspace string, excludes, deps []string) (map[string]bool, error) {
	pExclude, err := fileutils.NewPatternMatcher(excludes)
	if err != nil {
		return nil, fmt.Errorf("invalid exclude patterns: %w", err)
	}

	// Walk the workspace
	files := make(map[string]bool)
	for _, dep := range deps {
		dep = filepath.Clean(dep)
		absDep := filepath.Join(workspace, dep)

		fi, err := os.Stat(absDep)
		if err != nil {
			return nil, fmt.Errorf("stating file %q: %w", absDep, err)
		}

		switch mode := fi.Mode(); {
		case mode.IsDir():
			if err := godirwalk.Walk(absDep, &godirwalk.Options{
				Unsorted: true,
				Callback: func(fpath string, info *godirwalk.Dirent) error {
					if fpath == absDep {
						return nil
					}

					relPath, err := filepath.Rel(workspace, fpath)
					if err != nil {
						return err
					}

					ignored, err := pExclude.Matches(relPath)
					if err != nil {
						return err
					}

					if info.IsDir() {
						if !ignored {
							return nil
						}
						// exclusion handling closely follows vendor/github.com/docker/docker/pkg/archive/archive.go
						// No exceptions (!...) in patterns so just skip dir
						if !pExclude.Exclusions() {
							return filepath.SkipDir
						}

						dirSlash := relPath + string(filepath.Separator)

						for _, pat := range pExclude.Patterns() {
							if !pat.Exclusion() {
								continue
							}
							if strings.HasPrefix(pat.String()+string(filepath.Separator), dirSlash) {
								// found a match - so can't skip this dir
								return nil
							}
						}

						return filepath.SkipDir
					} else if !ignored {
						files[relPath] = true
					}

					return nil
				},
			}); err != nil {
				return nil, fmt.Errorf("walking folder %q: %w", absDep, err)
			}
		case mode.IsRegular():
			ignored, err := pExclude.Matches(dep)
			if err != nil {
				return nil, err
			}

			if !ignored {
				files[dep] = true
			}
		}
	}
	return files, nil
}
