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

package validation

import (
	"fmt"
	"testing"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
	"github.com/GoogleContainerTools/skaffold/testutil"
)

var (
	cfgWithErrors = &latest.SkaffoldConfig{
		Pipeline: latest.Pipeline{
			Build: latest.BuildConfig{
				Artifacts: []*latest.Artifact{
					{
						ArtifactType: latest.ArtifactType{
							DockerArtifact: &latest.DockerArtifact{},
							BazelArtifact:  &latest.BazelArtifact{},
						},
					},
					{
						ArtifactType: latest.ArtifactType{
							BazelArtifact:  &latest.BazelArtifact{},
							KanikoArtifact: &latest.KanikoArtifact{},
						},
					},
				},
			},
			Deploy: latest.DeployConfig{
				DeployType: latest.DeployType{
					HelmDeploy:    &latest.HelmDeploy{},
					KubectlDeploy: &latest.KubectlDeploy{},
				},
			},
		},
	}
)

func TestValidateSchema(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *latest.SkaffoldConfig
		shouldErr bool
	}{
		{
			name:      "config with errors",
			cfg:       cfgWithErrors,
			shouldErr: true,
		},
		{
			name:      "empty config",
			cfg:       &latest.SkaffoldConfig{},
			shouldErr: true,
		},
		{
			name: "minimal config",
			cfg: &latest.SkaffoldConfig{
				APIVersion: "foo",
				Kind:       "bar",
			},
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSchema(tt.cfg)
			testutil.CheckError(t, tt.shouldErr, err)
		})
	}
}

func alwaysErr(_ interface{}) error {
	return fmt.Errorf("always fail")
}

type emptyStruct struct{}
type nestedEmptyStruct struct {
	N emptyStruct
}

func TestVisitStructs(t *testing.T) {
	tests := []struct {
		name         string
		input        interface{}
		expectedErrs int
	}{
		{
			name:         "single struct to validate",
			input:        emptyStruct{},
			expectedErrs: 1,
		},
		{
			name:         "recurse into nested struct",
			input:        nestedEmptyStruct{},
			expectedErrs: 2,
		},
		{
			name: "check all slice items",
			input: struct {
				A []emptyStruct
			}{
				A: []emptyStruct{{}, {}},
			},
			expectedErrs: 3,
		},
		{
			name: "recurse into slices",
			input: struct {
				A []nestedEmptyStruct
			}{
				A: []nestedEmptyStruct{
					{
						N: emptyStruct{},
					},
				},
			},
			expectedErrs: 3,
		},
		{
			name: "recurse into ptr slices",
			input: struct {
				A []*nestedEmptyStruct
			}{
				A: []*nestedEmptyStruct{
					{
						N: emptyStruct{},
					},
				},
			},
			expectedErrs: 3,
		},
		{
			name: "ignore empty slices",
			input: struct {
				A []emptyStruct
			}{},
			expectedErrs: 1,
		},
		{
			name: "ignore nil pointers",
			input: struct {
				A *struct{}
			}{},
			expectedErrs: 1,
		},
		{
			name: "recurse into members",
			input: struct {
				A, B emptyStruct
			}{
				A: emptyStruct{},
				B: emptyStruct{},
			},
			expectedErrs: 3,
		},
		{
			name: "recurse into ptr members",
			input: struct {
				A, B *emptyStruct
			}{
				A: &emptyStruct{},
				B: &emptyStruct{},
			},
			expectedErrs: 3,
		},
		{
			name: "ignore other fields",
			input: struct {
				A emptyStruct
				C int
			}{
				A: emptyStruct{},
				C: 2,
			},
			expectedErrs: 2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := visitStructs(test.input, alwaysErr)

			testutil.CheckDeepEqual(t, test.expectedErrs, len(actual))
		})
	}
}
