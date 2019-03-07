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
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/util"
	"github.com/GoogleContainerTools/skaffold/testutil"
	"github.com/pkg/errors"
)

func TestGradleWrapperDefinition(t *testing.T) {
	testutil.CheckDeepEqual(t, "gradle", GradleCommand.Executable)
	testutil.CheckDeepEqual(t, "gradlew", GradleCommand.Wrapper)
}

func TestGetDependenciesGradle(t *testing.T) {
	tmpDir, cleanup := testutil.NewTempDir(t)
	defer cleanup()

	tmpDir.Write("build", "")
	tmpDir.Write("dep1", "")
	tmpDir.Write("dep2", "")

	build := tmpDir.Path("build")
	dep1 := tmpDir.Path("dep1")
	dep2 := tmpDir.Path("dep2")

	ctx := context.Background()

	var tests = []struct {
		description string
		stdout      string
		modTime     time.Time
		expected    []string
		err         error
	}{
		{
			description: "failure",
			stdout:      "",
			modTime:     time.Unix(0, 0),
			err:         errors.New("error"),
		},
		{
			description: "success",
			stdout:      fmt.Sprintf("BEGIN JIB JSON\n{\"build\":[\"%s\"],\"inputs\":[\"%s\"],\"ignore\":[]}", build, dep1),
			modTime:     time.Unix(0, 0),
			expected:    []string{build, dep1},
		},
		{
			// Expected output differs from stdout since build file hasn't change, thus gradle command won't run
			description: "success",
			stdout:      fmt.Sprintf("BEGIN JIB JSON\n{\"build\":[\"%s\"],\"inputs\":[\"%s\", \"%s\"],\"ignore\":[]}", build, dep1, dep2),
			modTime:     time.Unix(0, 0),
			expected:    []string{build, dep1},
		},
		{
			description: "success",
			stdout:      fmt.Sprintf("BEGIN JIB JSON\n{\"build\":[\"%s\"],\"inputs\":[\"%s\", \"%s\"],\"ignore\":[]}", build, dep1, dep2),
			modTime:     time.Unix(10000, 0),
			expected:    []string{build, dep1, dep2},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			defer func(c util.Command) { util.DefaultExecCommand = c }(util.DefaultExecCommand)
			util.DefaultExecCommand = testutil.NewFakeCmd(t).WithRunOutErr(
				strings.Join(getCommandGradle(ctx, tmpDir.Root(), &latest.JibGradleArtifact{Project: "gradle-test"}).Args, " "),
				test.stdout,
				test.err,
			)

			// Change build file mod time
			os.Chtimes(build, test.modTime, test.modTime)

			deps, err := GetDependenciesGradle(ctx, tmpDir.Root(), &latest.JibGradleArtifact{Project: "gradle-test"})
			if test.err != nil {
				testutil.CheckErrorAndDeepEqual(t, true, err, "getting jibGradle dependencies: failed to get Jib dependencies; it's possible you are using an old version of Jib (Skaffold requires Jib v1.0.2+): "+test.err.Error(), err.Error())
			} else {
				testutil.CheckDeepEqual(t, test.expected, deps)
			}
		})
	}
}

func TestGetCommandGradle(t *testing.T) {
	ctx := context.Background()

	var tests = []struct {
		description       string
		jibGradleArtifact latest.JibGradleArtifact
		filesInWorkspace  []string
		expectedCmd       func(workspace string) *exec.Cmd
	}{
		{
			description:       "gradle default",
			jibGradleArtifact: latest.JibGradleArtifact{},
			filesInWorkspace:  []string{},
			expectedCmd: func(workspace string) *exec.Cmd {
				return GradleCommand.CreateCommand(ctx, workspace, []string{":_jibSkaffoldFilesV2", "-q"})
			},
		},
		{
			description:       "gradle default with project",
			jibGradleArtifact: latest.JibGradleArtifact{Project: "project"},
			filesInWorkspace:  []string{},
			expectedCmd: func(workspace string) *exec.Cmd {
				return GradleCommand.CreateCommand(ctx, workspace, []string{":project:_jibSkaffoldFilesV2", "-q"})
			},
		},
		{
			description:       "gradle with wrapper",
			jibGradleArtifact: latest.JibGradleArtifact{},
			filesInWorkspace:  []string{"gradlew", "gradlew.cmd"},
			expectedCmd: func(workspace string) *exec.Cmd {
				return GradleCommand.CreateCommand(ctx, workspace, []string{":_jibSkaffoldFilesV2", "-q"})
			},
		},
		{
			description:       "gradle with wrapper and project",
			jibGradleArtifact: latest.JibGradleArtifact{Project: "project"},
			filesInWorkspace:  []string{"gradlew", "gradlew.cmd"},
			expectedCmd: func(workspace string) *exec.Cmd {
				return GradleCommand.CreateCommand(ctx, workspace, []string{":project:_jibSkaffoldFilesV2", "-q"})
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

			cmd := getCommandGradle(ctx, tmpDir.Root(), &test.jibGradleArtifact)
			expectedCmd := test.expectedCmd(tmpDir.Root())
			testutil.CheckDeepEqual(t, expectedCmd.Path, cmd.Path)
			testutil.CheckDeepEqual(t, expectedCmd.Args, cmd.Args)
			testutil.CheckDeepEqual(t, expectedCmd.Dir, cmd.Dir)
		})
	}
}

func TestGenerateGradleArgs(t *testing.T) {
	var testCases = []struct {
		in        latest.JibGradleArtifact
		skipTests bool
		out       []string
	}{
		{latest.JibGradleArtifact{}, false, []string{":task", "--image=image"}},
		{latest.JibGradleArtifact{Flags: []string{"-extra", "args"}}, false, []string{":task", "--image=image", "-extra", "args"}},
		{latest.JibGradleArtifact{}, true, []string{":task", "--image=image", "-x", "test"}},
		{latest.JibGradleArtifact{Project: "project"}, false, []string{":project:task", "--image=image"}},
		{latest.JibGradleArtifact{Project: "project"}, true, []string{":project:task", "--image=image", "-x", "test"}},
	}

	for _, tt := range testCases {
		command := GenerateGradleArgs("task", "image", &tt.in, tt.skipTests)

		testutil.CheckDeepEqual(t, tt.out, command)
	}
}
