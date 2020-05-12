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
	"io"
	"testing"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/config"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/runner"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
	"github.com/GoogleContainerTools/skaffold/testutil"
)

// Verify workaround so that Dev and Debug can have separate defaults for Auto{Build,Deploy,Sync} 
// https://github.com/GoogleContainerTools/skaffold/issues/4129 
// https://github.com/spf13/pflag/issues/257
func TestDebugIndependentFromDev(t *testing.T) {
	mockRunner := &mockDevRunner{}
	testutil.Run(t, "DevDebug", func(t *testutil.T) {
		t.Override(&createRunner, func(config.SkaffoldOptions) (runner.Runner, *latest.SkaffoldConfig, error) {
			return mockRunner, &latest.SkaffoldConfig{}, nil
		})
		t.Override(&opts, config.SkaffoldOptions{})
		t.Override(&doDev, func(_ context.Context, _ io.Writer) error {
			if !opts.AutoBuild {
				t.Error("opts.AutoBuild should be true for dev")
			}
			if !opts.AutoDeploy {
				t.Error("opts.AutoDeploy should be true for dev")
			}
			if !opts.AutoSync {
				t.Error("opts.AutoSync should be true for dev")
			}
			return nil
		})
		t.Override(&doDebug, func(_ context.Context, _ io.Writer) error {
			if opts.AutoBuild {
				t.Error("opts.AutoBuild should be false for `debug`")
			}
			if opts.AutoDeploy {
				t.Error("opts.AutoDeploy should be false for `debug`")
			}
			if opts.AutoSync {
				t.Error("opts.AutoSync should be false for `debug`")
			}
			return nil
		})

		// debug should be independent of dev
		dev := NewCmdDev()
		debug := NewCmdDebug()

		t.Log("Round 1")
		dev.Execute()
		debug.Execute()
		t.Log("Round 2")
		dev.Execute()
		debug.Execute()
	})
}
