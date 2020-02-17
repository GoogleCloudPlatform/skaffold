/*
Copyright 2020 The Skaffold Authors

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

package analyze

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/karrick/godirwalk"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/initializer/build"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/initializer/config"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/util"
)

// analyzer is following the visitor pattern. It is called on every file
// as the analysis.analyze function walks the directory structure recursively.
// It can manage state and react to walking events assuming a breadth first search.
type analyzer interface {
	enterDir(dir string)
	analyzeFile(file string) error
	exitDir(dir string)
}

type ProjectAnalysis struct {
	configAnalyzer  *skaffoldConfigAnalyzer
	kubeAnalyzer    *kubeAnalyzer
	builderAnalyzer *builderAnalyzer
}

func (a *ProjectAnalysis) Builders() []build.InitBuilder {
	return a.builderAnalyzer.foundBuilders
}

func (a *ProjectAnalysis) Manifests() []string {
	return a.kubeAnalyzer.kubernetesManifests
}

func (a *ProjectAnalysis) analyzers() []analyzer {
	return []analyzer{
		a.kubeAnalyzer,
		a.configAnalyzer,
		a.builderAnalyzer,
	}
}

// newAnalysis sets up the analysis of the directory based on the initializer configuration
func NewAnalyzer(c config.Config) *ProjectAnalysis {
	return &ProjectAnalysis{
		kubeAnalyzer: &kubeAnalyzer{},
		builderAnalyzer: &builderAnalyzer{
			findBuilders:         !c.SkipBuild,
			enableJibInit:        c.EnableJibInit,
			enableBuildpacksInit: c.EnableBuildpacksInit,
			buildpacksBuilder:    c.BuildpacksBuilder,
		},
		configAnalyzer: &skaffoldConfigAnalyzer{
			force:        c.Force,
			analyzeMode:  c.Analyze,
			targetConfig: c.Opts.ConfigurationFile,
		},
	}
}

// analyze recursively walks a directory and notifies the analyzers of files and enterDir and exitDir events
// at the end of the analyze function the analysis struct's analyzers should contain the state that we can
// use to do further computation.
func (a *ProjectAnalysis) Analyze(dir string) error {
	for _, analyzer := range a.analyzers() {
		analyzer.enterDir(dir)
	}
	dirents, err := godirwalk.ReadDirents(dir, nil)
	if err != nil {
		return err
	}

	var subdirectories []*godirwalk.Dirent
	//this is for deterministic results - given the same directory structure
	//init should have the same results
	sort.Sort(dirents)

	// Traverse files
	for _, file := range dirents {
		if util.IsHiddenFile(file.Name()) || util.IsHiddenDir(file.Name()) {
			continue
		}

		// If we found a directory, keep track of it until we've gone through all the files first
		if file.IsDir() {
			subdirectories = append(subdirectories, file)
			continue
		}

		filePath := filepath.Join(dir, file.Name())
		for _, analyzer := range a.analyzers() {
			// to make skaffold.yaml more portable across OS-es we should generate always / based filePaths
			filePath = strings.ReplaceAll(filePath, string(os.PathSeparator), "/")
			if err := analyzer.analyzeFile(filePath); err != nil {
				return err
			}
		}
	}

	// Recurse into subdirectories
	for _, subdir := range subdirectories {
		if err = a.Analyze(filepath.Join(dir, subdir.Name())); err != nil {
			return err
		}
	}

	for _, analyzer := range a.analyzers() {
		analyzer.exitDir(dir)
	}
	return nil
}
