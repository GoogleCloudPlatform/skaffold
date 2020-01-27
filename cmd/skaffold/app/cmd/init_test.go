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

package cmd

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/initializer"

	"github.com/GoogleContainerTools/skaffold/testutil"
)

func TestFlagsToConfigVersion(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedConfig initializer.Config
		initResult     error
		shouldErr      bool
	}{
		{
			name:       "error + default values of flags mapped to config values",
			initResult: errors.New("test error"),
			shouldErr:  true,
			expectedConfig: initializer.Config{
				ComposeFile:         "",
				CliArtifacts:        nil,
				SkipBuild:           false,
				SkipDeploy:          false,
				Force:               false,
				Analyze:             false,
				EnableJibInit:       false,
				EnableBuildpackInit: false,
				Opts:                opts,
			},
		},
		{
			name: "no error + non-default values for flags mapped to config values",
			args: []string{
				"--compose-file=a-compose-file",
				"--artifact", "a1=b1",
				"-a", "a2=b2",
				"--skip-build",
				"--skip-deploy",
				"--force",
				"--analyze",
				"--XXenableJibInit",
				"--XXenableBuildpackInit",
			},
			expectedConfig: initializer.Config{
				ComposeFile:         "a-compose-file",
				CliArtifacts:        []string{"a1=b1", "a2=b2"},
				SkipBuild:           true,
				SkipDeploy:          true,
				Force:               true,
				Analyze:             true,
				EnableJibInit:       true,
				EnableBuildpackInit: true,
				Opts:                opts,
			},
		},
	}
	for _, test := range tests {
		testutil.Run(t, test.name, func(t *testutil.T) {
			var capturedConfig initializer.Config
			mockFunc := func(ctx context.Context, out io.Writer, c initializer.Config) error {
				capturedConfig = c
				return test.initResult
			}
			t.Override(&initEntrypoint, mockFunc)
			initArgs := append([]string{"init"}, test.args...)
			os.Args = initArgs
			init := NewCmdInit()
			err := init.Execute()
			// we ignore Skaffold options
			test.expectedConfig.Opts = capturedConfig.Opts
			t.CheckErrorAndDeepEqual(test.shouldErr, err, test.expectedConfig, capturedConfig)
		})
	}
}
