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
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/v1alpha3"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/util"
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GetDependenciesMaven finds the source dependencies for the given jib-maven artifact.
// All paths are absolute.
// TODO(coollog): Add support for multi-module projects.
func GetDependenciesMaven(workspace string, a *v1alpha3.JibMavenArtifact, isWindows bool) ([]string, error) {
	executable, subCommand := getCommandMaven(workspace, a, isWindows)
	return getDependencies(workspace, "pom.xml", executable, subCommand, "jib-maven")
}

// GetDependenciesGradle finds the source dependencies for the given jib-gradle artifact.
// All paths are absolute.
// TODO(coollog): Add support for multi-module projects.
func GetDependenciesGradle(workspace string, a *v1alpha3.JibGradleArtifact, isWindows bool) ([]string, error) {
	executable, subCommand := getCommandGradle(workspace, a, isWindows)
	return getDependencies(workspace, "build.gradle", executable, subCommand, "jib-gradle")
}

func exists(workspace string, filename string) bool {
	_, err := os.Stat(filepath.Join(workspace, filename))
	return err == nil
}

func getDependencies(workspace string, buildFile string, executable string, subCommand []string, artifactName string) ([]string, error) {
	if !exists(workspace, buildFile) {
		return nil, errors.Errorf("no %s found", buildFile)
	}

	cmd := exec.Command(executable, subCommand...)
	cmd.Dir = workspace
	stdout, err := util.RunCmdOut(cmd)
	if err != nil {
		return nil, errors.Wrapf(err, "getting %s dependencies", artifactName)
	}

	return getDepsFromStdout(string(stdout)), nil
}

const (
	mavenExecutable     = "mvn"
	mavenWrapper        = "mvnw"
	mavenWindowsWrapper = "mvnw.cmd"

	gradleExecutable     = "gradle"
	gradleWrapper        = "gradlew"
	gradleWindowsWrapper = "gradlew.bat"
)

func getCommandMaven(workspace string, a *v1alpha3.JibMavenArtifact, isWindows bool) (executable string, subCommand []string) {
	subCommand = []string{"jib:_skaffold-files", "-q"}
	if a.Profile != "" {
		subCommand = append(subCommand, "-P", a.Profile)
	}

	return getCommand(workspace, mavenExecutable, subCommand, mavenWrapper, mavenWindowsWrapper, isWindows)
}

func getCommandGradle(workspace string, _ /* a */ *v1alpha3.JibGradleArtifact, isWindows bool) (executable string, subCommand []string) {
	return getCommand(workspace, gradleExecutable, []string{"_jibSkaffoldFiles", "-q"}, gradleWrapper, gradleWindowsWrapper, isWindows)
}

func getCommand(workspace string, defaultExecutable string, defaultSubCommand []string, wrapper string, windowsWrapper string, isWindows bool) (executable string, subCommand []string) {
	executable = defaultExecutable
	subCommand = defaultSubCommand

	if isWindows {
		if exists(workspace, windowsWrapper) {
			executable = "cmd.exe"
			subCommand = append([]string{windowsWrapper}, subCommand...)
			subCommand = append([]string{"/C"}, subCommand...)
		}
	} else {
		wrapperExecutable := wrapper
		if exists(workspace, wrapperExecutable) {
			executable = "./" + wrapperExecutable
		}
	}

	return executable, subCommand
}

func getDepsFromStdout(stdout string) []string {
	lines := strings.Split(stdout, "\n")
	deps := []string{}
	for _, l := range lines {
		if l == "" {
			continue
		}
		deps = append(deps, l)
	}
	return deps
}
