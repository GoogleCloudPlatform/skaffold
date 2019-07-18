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

package cmd

import (
	"context"
	"io"
	"io/ioutil"
	"testing"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/config"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/runner"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
	"github.com/GoogleContainerTools/skaffold/testutil"
)

type mockDevRunner struct {
	runner.Runner
	hasBuilt    bool
	hasDeployed bool
	errDev      error
	calls       []string
}

func (r *mockDevRunner) Dev(context.Context, io.Writer, []*latest.Artifact) error {
	r.calls = append(r.calls, "Dev")
	return r.errDev
}

func (r *mockDevRunner) HasBuilt() bool {
	r.calls = append(r.calls, "HasBuilt")
	return r.hasBuilt
}

func (r *mockDevRunner) HasDeployed() bool {
	r.calls = append(r.calls, "HasDeployed")
	return r.hasDeployed
}

func (r *mockDevRunner) Prune(context.Context, io.Writer) error {
	r.calls = append(r.calls, "Prune")
	return nil
}

func (r *mockDevRunner) Cleanup(context.Context, io.Writer) error {
	r.calls = append(r.calls, "Cleanup")
	return nil
}

func TestDoDev(t *testing.T) {
	tests := []struct {
		description   string
		hasBuilt      bool
		hasDeployed   bool
		expectedCalls []string
	}{
		{
			description:   "cleanup and then prune",
			hasBuilt:      true,
			hasDeployed:   true,
			expectedCalls: []string{"Dev", "HasDeployed", "HasBuilt", "Cleanup", "Prune"},
		},
		{
			description:   "hasn't deployed",
			hasBuilt:      true,
			hasDeployed:   false,
			expectedCalls: []string{"Dev", "HasDeployed", "HasBuilt", "Prune"},
		},
		{
			description:   "hasn't built",
			hasBuilt:      false,
			hasDeployed:   false,
			expectedCalls: []string{"Dev", "HasDeployed", "HasBuilt"},
		},
	}

	for _, test := range tests {
		testutil.Run(t, test.description, func(t *testutil.T) {
			mockRunner := &mockDevRunner{
				hasBuilt:    test.hasBuilt,
				hasDeployed: test.hasDeployed,
				errDev:      context.Canceled,
			}
			t.Override(&createRunner, func(*config.SkaffoldOptions) (runner.Runner, *latest.SkaffoldConfig, error) {
				return mockRunner, &latest.SkaffoldConfig{}, nil
			})
			t.Override(&opts, &config.SkaffoldOptions{
				Cleanup: true,
				NoPrune: false,
			})

			err := doDev(context.Background(), ioutil.Discard)

			t.CheckDeepEqual(test.expectedCalls, mockRunner.calls)
			t.CheckDeepEqual(true, err == context.Canceled)
		})
	}
}
