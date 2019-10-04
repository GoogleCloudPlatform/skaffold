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

package runner

import (
	"context"
	"io/ioutil"
	"testing"

	v1 "k8s.io/api/core/v1"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/build"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
	"github.com/GoogleContainerTools/skaffold/testutil"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/clientcmd/api"
)

func TestDeployLogging(t *testing.T) {
	testutil.Run(t, "should list pods for logging", func(t *testutil.T) {
		t.SetupFakeKubernetesContext(api.Config{CurrentContext: "cluster1"})

		ctx := context.Background()
		artifacts := []build.Artifact{{
			ImageName: "img",
			Tag:       "img:1",
		}}

		bench := &TestBench{}
		runner := createRunner(t, bench, nil)
		runner.runCtx.Opts.Tail = true

		err := runner.DeployAndLog(ctx, ioutil.Discard, artifacts)

		t.CheckErrorAndDeepEqual(false, err, []Actions{{Deployed: []string{"img:1"}}}, bench.Actions())
		t.CheckDeepEqual(true, runner.podSelector.Select(&v1.Pod{
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Image: "img:1",
					},
				},
			},
		}))
	})
}
func TestBuildTestDeploy(t *testing.T) {
	tests := []struct {
		description     string
		testBench       *TestBench
		shouldErr       bool
		expectedActions []Actions
	}{
		{
			description: "run no error",
			testBench:   &TestBench{},
			expectedActions: []Actions{{
				Built:    []string{"img:1"},
				Tested:   []string{"img:1"},
				Deployed: []string{"img:1"},
			}},
		},
		{
			description:     "run build error",
			testBench:       &TestBench{buildErrors: []error{errors.New("")}},
			shouldErr:       true,
			expectedActions: []Actions{{}},
		},
		{
			description: "run test error",
			testBench:   &TestBench{testErrors: []error{errors.New("")}},
			shouldErr:   true,
			expectedActions: []Actions{{
				Built: []string{"img:1"},
			}},
		},
		{
			description: "run deploy error",
			testBench:   &TestBench{deployErrors: []error{errors.New("")}},
			shouldErr:   true,
			expectedActions: []Actions{{
				Built:  []string{"img:1"},
				Tested: []string{"img:1"},
			}},
		},
	}
	for _, test := range tests {
		testutil.Run(t, test.description, func(t *testutil.T) {
			t.SetupFakeKubernetesContext(api.Config{CurrentContext: "cluster1"})

			ctx := context.Background()
			artifacts := []*latest.Artifact{{
				ImageName: "img",
			}}

			runner := createRunner(t, test.testBench, nil)
			bRes, err := runner.BuildAndTest(ctx, ioutil.Discard, artifacts)
			if err == nil {
				err = runner.DeployAndLog(ctx, ioutil.Discard, bRes)
			}

			t.CheckErrorAndDeepEqual(test.shouldErr, err, test.expectedActions, test.testBench.Actions())
		})
	}
}
