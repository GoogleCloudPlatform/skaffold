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

package jib

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/util"
	"github.com/karrick/godirwalk"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// filesLists contains cached build/input dependencies
type filesLists struct {
	// BuildDefinitions lists paths to build definitions that trigger a call out to Jib to refresh the pathMap, as well as a rebuild, upon changing
	BuildDefinitions []string `json:"build"`

	// Inputs lists paths to build dependencies that trigger a rebuild upon changing
	Inputs []string `json:"inputs"`

	// Results lists paths to files that should be ignored when checking for changes to rebuild
	Results []string `json:"ignore"`

	// BuildFileTimes keeps track of the last modification time of each build file
	BuildFileTimes map[string]time.Time
}

// watchedFiles maps from project name to watched files
var watchedFiles = map[string]filesLists{}

// getDependencies returns a list of files to watch for changes to rebuild
func getDependencies(cmd *exec.Cmd, projectName string) ([]string, error) {
	var dependencyList []string
	files, ok := watchedFiles[projectName]
	if !ok {
		files = filesLists{}
	}

	if len(files.Inputs) == 0 && len(files.BuildDefinitions) == 0 {
		// Make sure build file modification time map is setup
		if files.BuildFileTimes == nil {
			files.BuildFileTimes = make(map[string]time.Time)
		}

		// Refresh dependency list if empty
		if err := refreshDependencyList(&files, cmd); err != nil {
			return nil, errors.Wrap(err, "initial Jib dependency refresh failed")
		}

	} else if err := walkFiles(files.BuildDefinitions, files.Results, func(path string, info os.FileInfo) error {
		// Walk build files to check for changes
		if val, ok := files.BuildFileTimes[path]; !ok || info.ModTime() != val {
			return refreshDependencyList(&files, cmd)
		}
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "failed to walk Jib build files for changes")
	}

	// Walk updated files to build dependency list
	if err := walkFiles(files.Inputs, files.Results, func(path string, info os.FileInfo) error {
		dependencyList = append(dependencyList, path)
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "failed to walk Jib input files to build dependency list")
	}
	if err := walkFiles(files.BuildDefinitions, files.Results, func(path string, info os.FileInfo) error {
		dependencyList = append(dependencyList, path)
		files.BuildFileTimes[path] = info.ModTime()
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "failed to walk Jib build files to build dependency list")
	}

	// Store updated files list information
	watchedFiles[projectName] = files

	sort.Strings(dependencyList)
	return dependencyList, nil
}

// refreshDependencyList calls out to Jib to update files with the latest list of files/directories to watch.
func refreshDependencyList(files *filesLists, cmd *exec.Cmd) error {
	stdout, err := util.RunCmdOut(cmd)
	if err != nil {
		return errors.Wrap(err, "failed to get Jib dependencies; it's possible you are using an old version of Jib (Skaffold requires Jib v1.0.2+)")
	}

	// Search for Jib's output JSON. Jib's Maven/Gradle output takes the following form:
	// ...
	// BEGIN JIB JSON
	// {"build":["/paths","/to","/buildFiles"],"inputs":["/paths","/to","/inputs"],"ignore":["/paths","/to","/ignore"]}
	// ...
	// To parse the output, search for "BEGIN JIB JSON", then unmarshal the next line into the pathMap struct.
	matches := regexp.MustCompile(`BEGIN JIB JSON\r?\n({.*})`).FindSubmatch(stdout)
	if len(matches) == 0 {
		return errors.New("failed to get Jib dependencies")
	}

	line := bytes.Replace(matches[1], []byte(`\`), []byte(`\\`), -1)
	return json.Unmarshal([]byte(line), &files)
}

// walkFiles walks through a list of files and directories and performs a callback on each of the files
func walkFiles(watchedFiles []string, ignoredFiles []string, callback func(path string, info os.FileInfo) error) error {
	for _, dep := range watchedFiles {
		if isIgnored(dep, ignoredFiles) {
			continue
		}

		// Resolves directories recursively.
		info, err := os.Stat(dep)
		if err != nil {
			if os.IsNotExist(err) {
				logrus.Debugf("could not stat dependency: %s", err)
				continue // Ignore files that don't exist
			}
			return errors.Wrapf(err, "unable to stat file %s", dep)
		}

		// Process file
		if !info.IsDir() {
			if err := callback(dep, info); err != nil {
				return err
			}
			continue
		}

		// Process directory
		if err = godirwalk.Walk(dep, &godirwalk.Options{
			Unsorted: true,
			Callback: func(path string, _ *godirwalk.Dirent) error {
				if isIgnored(path, ignoredFiles) {
					return filepath.SkipDir
				}
				pathInfo, err := os.Stat(dep)
				if err != nil {
					if os.IsNotExist(err) {
						logrus.Debugf("could not stat dependency: %s", err)
						return filepath.SkipDir
					}
					return errors.Wrapf(err, "unable to stat file %s", dep)
				}
				return callback(path, pathInfo)
			},
		}); err != nil {
			return errors.Wrap(err, "filepath walk")
		}
	}
	return nil
}

// isIgnored tests a path for whether or not it should be ignored according to a list of ignored files/directories
func isIgnored(path string, ignoredFiles []string) bool {
	for _, ignored := range ignoredFiles {
		if strings.HasPrefix(path, ignored) {
			return true
		}
	}
	return false
}
