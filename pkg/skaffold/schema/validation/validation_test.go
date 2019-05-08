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
			err := Process(tt.cfg)
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
		{
			name: "unexported fields",
			input: struct {
				a emptyStruct
			}{
				a: emptyStruct{},
			},
			expectedErrs: 1,
		},
		{
			name: "exported and unexported fields",
			input: struct {
				a, A, b emptyStruct
			}{
				a: emptyStruct{},
				A: emptyStruct{},
				b: emptyStruct{},
			},
			expectedErrs: 2,
		},
		{
			name: "unexported nil ptr fields",
			input: struct {
				a *emptyStruct
			}{
				a: nil,
			},
			expectedErrs: 1,
		},
		{
			name: "unexported ptr fields",
			input: struct {
				a *emptyStruct
			}{
				a: &emptyStruct{},
			},
			expectedErrs: 1,
		},
		{
			name: "unexported and exported ptr fields",
			input: struct {
				a, A, b *emptyStruct
			}{
				a: &emptyStruct{},
				A: &emptyStruct{},
				b: &emptyStruct{},
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

func TestValidateNetworkMode(t *testing.T) {
	tests := []struct {
		name      string
		artifacts []*latest.Artifact
		shouldErr bool
	}{
		{
			name: "not a docker artifact",
			artifacts: []*latest.Artifact{
				{
					ImageName: "image/bazel",
					ArtifactType: latest.ArtifactType{
						BazelArtifact: &latest.BazelArtifact{},
					},
				},
			},
		},
		{
			name: "no networkmode",
			artifacts: []*latest.Artifact{
				{
					ImageName: "image/no-network",
					ArtifactType: latest.ArtifactType{
						DockerArtifact: &latest.DockerArtifact{},
					},
				},
			},
		},
		{
			name: "bridge",
			artifacts: []*latest.Artifact{
				{
					ImageName: "image/bridge",
					ArtifactType: latest.ArtifactType{
						DockerArtifact: &latest.DockerArtifact{
							NetworkMode: "Bridge",
						},
					},
				},
			},
		},
		{
			name: "none",
			artifacts: []*latest.Artifact{
				{
					ImageName: "image/none",
					ArtifactType: latest.ArtifactType{
						DockerArtifact: &latest.DockerArtifact{
							NetworkMode: "None",
						},
					},
				},
			},
		},
		{
			name: "host",
			artifacts: []*latest.Artifact{
				{
					ImageName: "image/host",
					ArtifactType: latest.ArtifactType{
						DockerArtifact: &latest.DockerArtifact{
							NetworkMode: "Host",
						},
					},
				},
			},
		},
		{
			name:      "invalid networkmode",
			shouldErr: true,
			artifacts: []*latest.Artifact{
				{
					ImageName: "image/bad",
					ArtifactType: latest.ArtifactType{
						DockerArtifact: &latest.DockerArtifact{
							NetworkMode: "Bad",
						},
					},
				},
			},
		},
		{
			name: "case insensitive",
			artifacts: []*latest.Artifact{
				{
					ImageName: "image/case-insensitive",
					ArtifactType: latest.ArtifactType{
						DockerArtifact: &latest.DockerArtifact{
							NetworkMode: "bRiDgE",
						},
					},
				},
			},
		},
	}

	origValidateYamlTags := validateYamltags
	validateYamltags = func(_ interface{}) error { return nil }
	defer func() { validateYamltags = origValidateYamlTags }()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := Process(
				&latest.SkaffoldConfig{
					Pipeline: latest.Pipeline{
						Build: latest.BuildConfig{
							Artifacts: test.artifacts,
						},
					},
				})
			testutil.CheckError(t, test.shouldErr, err)
		})
	}
}

func TestValidateCustomDependencies(t *testing.T) {
	tests := []struct {
		description    string
		dependencies   *latest.CustomDependencies
		expectedErrors int
	}{
		{
			description: "no errors",
			dependencies: &latest.CustomDependencies{
				Paths:  []string{"somepath"},
				Ignore: []string{"anotherpath"},
			},
		}, {
			description: "ignore in conjunction with dockerfile",
			dependencies: &latest.CustomDependencies{
				Dockerfile: &latest.DockerfileDependency{
					Path: "some/path",
				},
				Ignore: []string{"ignoreme"},
			},
			expectedErrors: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			artifact := &latest.Artifact{
				ArtifactType: latest.ArtifactType{
					CustomArtifact: &latest.CustomArtifact{
						Dependencies: test.dependencies,
					},
				},
			}

			errs := validateCustomDependencies([]*latest.Artifact{artifact})
			if len(errs) != test.expectedErrors {
				t.Fatalf("got incorrect number of errors. got: %d \n expected: %d \n", len(errs), test.expectedErrors)
			}
		})
	}
}
