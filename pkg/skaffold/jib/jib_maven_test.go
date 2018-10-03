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
	"strings"
	"testing"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/v1alpha3"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/util"
	"github.com/GoogleContainerTools/skaffold/testutil"
	"github.com/pkg/errors"
)

func TestGetDependenciesMaven(t *testing.T) {
	var tests = []struct {
		description string
		stdout      string
		err         error
	}{
		{
			description: "success",
			stdout:      "dep1\ndep2\n\n\n",
			err:         nil,
		},
		{
			description: "failure",
			stdout:      "",
			err:         errors.New("error"),
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			tmpDir, cleanup := testutil.NewTempDir(t)
			defer cleanup()

			defer func(c util.Command) { util.DefaultExecCommand = c }(util.DefaultExecCommand)
			util.DefaultExecCommand = testutil.NewFakeCmdOut(
				strings.Join(getCommandMaven(tmpDir.Root(), &v1alpha3.JibMavenArtifact{}).Args, " "),
				test.stdout,
				test.err,
			)

			deps, err := GetDependenciesMaven(tmpDir.Root(), &v1alpha3.JibMavenArtifact{})
			if test.err != nil {
				testutil.CheckErrorAndDeepEqual(t, true, err, "getting jib-maven dependencies: "+test.err.Error(), err.Error())
			} else {
				testutil.CheckDeepEqual(t, []string{"dep1", "dep2"}, deps)
			}
		})
	}
}

func TestGetCommandMaven(t *testing.T) {
	var tests = []struct {
		description      string
		jibMavenArtifact v1alpha3.JibMavenArtifact
		filesInWorkspace []string
		expectedCmd      func(workspace string) *exec.Cmd
	}{
		{
			description:      "maven no profile",
			jibMavenArtifact: v1alpha3.JibMavenArtifact{},
			filesInWorkspace: []string{},
			expectedCmd: func(workspace string) *exec.Cmd {
				return getCommand(workspace, "mvn", "ignored", []string{"jib:_skaffold-files", "-q"})
			},
		},
		{
			description:      "maven with profile",
			jibMavenArtifact: v1alpha3.JibMavenArtifact{Profile: "profile"},
			filesInWorkspace: []string{},
			expectedCmd: func(workspace string) *exec.Cmd {
				return getCommand(workspace, "mvn", "ignored", []string{"jib:_skaffold-files", "-q", "-P", "profile"})
			},
		},
		{
			description:      "maven with wrapper no profile",
			jibMavenArtifact: v1alpha3.JibMavenArtifact{},
			filesInWorkspace: []string{getWrapperMaven()},
			expectedCmd: func(workspace string) *exec.Cmd {
				return getCommand(workspace, "ignored", getWrapperMaven(), []string{"jib:_skaffold-files", "-q"})
			},
		},
		{
			description:      "maven with wrapper and profile",
			jibMavenArtifact: v1alpha3.JibMavenArtifact{Profile: "profile"},
			filesInWorkspace: []string{getWrapperMaven()},
			expectedCmd: func(workspace string) *exec.Cmd {
				return getCommand(workspace, "ignored", getWrapperMaven(), []string{"jib:_skaffold-files", "-q", "-P", "profile"})
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			tmpDir, cleanup := testutil.NewTempDir(t)
			defer cleanup()

			for _, file := range test.filesInWorkspace {
				tmpDir.Write(file, "")
			}

			cmd := getCommandMaven(tmpDir.Root(), &test.jibMavenArtifact)
			expectedCmd := test.expectedCmd(tmpDir.Root())
			testutil.CheckDeepEqual(t, expectedCmd.Path, cmd.Path)
			testutil.CheckDeepEqual(t, expectedCmd.Args, cmd.Args)
			testutil.CheckDeepEqual(t, expectedCmd.Dir, cmd.Dir)
		})
	}
}
