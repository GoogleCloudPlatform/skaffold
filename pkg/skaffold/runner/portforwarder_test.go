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
	"testing"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
	"github.com/GoogleContainerTools/skaffold/testutil"
)

func TestCreateLabelQuery(t *testing.T) {
	tests := []struct {
		description string
		deploy      *latest.HelmDeploy
		expected    string
	}{
		{
			description: "one helm release",
			deploy: &latest.HelmDeploy{
				Releases: []latest.HelmRelease{{Name: "foo"}},
			},
			expected: "release in (foo)",
		}, {
			description: "multiple helm release",
			deploy: &latest.HelmDeploy{
				Releases: []latest.HelmRelease{{Name: "foo"}, {Name: "bar"}},
			},
			expected: "release in (foo,bar)",
		}, {
			description: "no helm release",
			expected:    "K8ManagedBy=test",
		},
	}
	for _, test := range tests {
		testutil.Run(t, test.description, func(t *testutil.T) {
			actual := createLabelQuery(test.deploy, "K8ManagedBy=test")
			t.CheckDeepEqual(actual, test.expected)
		})
	}
}
