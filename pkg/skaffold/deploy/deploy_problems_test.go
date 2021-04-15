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

package deploy

import (
	"fmt"
	"testing"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/config"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/constants"
	sErrors "github.com/GoogleContainerTools/skaffold/pkg/skaffold/errors"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
	"github.com/GoogleContainerTools/skaffold/proto/v1"
	"github.com/GoogleContainerTools/skaffold/testutil"
)

func TestSuggestDeployFailedAction(t *testing.T) {
	tests := []struct {
		description string
		context     string
		err         error
		isMinikube  bool
		expected    string
		expectedAE  *proto.ActionableErr
	}{
		{
			description: "minikube status",
			context:     "minikube",
			isMinikube:  true,
			err:         fmt.Errorf("exiting dev mode because first deploy failed: unable to connect to Kubernetes: Get \"https://192.168.64.3:8443/version?timeout=32s\": net/http: TLS handshake timeout"),
			expected:    "Deploy Failed. Could not connect to cluster due to \"https://192.168.64.3:8443/version?timeout=32s\": net/http: TLS handshake timeout. Check if minikube is running using \"minikube status\" command and try again.",
			expectedAE: &proto.ActionableErr{
				ErrCode: proto.StatusCode_DEPLOY_CLUSTER_CONNECTION_ERR,
				Message: "exiting dev mode because first deploy failed: unable to connect to Kubernetes: Get \"https://192.168.64.3:8443/version?timeout=32s\": net/http: TLS handshake timeout",
				Suggestions: []*proto.Suggestion{{
					SuggestionCode: proto.SuggestionCode_CHECK_MINIKUBE_STATUS,
					Action:         "Check if minikube is running using \"minikube status\" command and try again",
				}},
			},
		},
		{
			description: "minikube status named ctx",
			context:     "test_cluster",
			isMinikube:  true,
			err:         fmt.Errorf("exiting dev mode because first deploy failed: unable to connect to Kubernetes: Get \"https://192.168.64.3:8443/version?timeout=32s\": net/http: TLS handshake timeout"),
			expected:    "Deploy Failed. Could not connect to cluster due to \"https://192.168.64.3:8443/version?timeout=32s\": net/http: TLS handshake timeout. Check if minikube is running using \"minikube status -p test_cluster\" command and try again.",
			expectedAE: &proto.ActionableErr{
				ErrCode: proto.StatusCode_DEPLOY_CLUSTER_CONNECTION_ERR,
				Message: "exiting dev mode because first deploy failed: unable to connect to Kubernetes: Get \"https://192.168.64.3:8443/version?timeout=32s\": net/http: TLS handshake timeout",
				Suggestions: []*proto.Suggestion{{
					SuggestionCode: proto.SuggestionCode_CHECK_MINIKUBE_STATUS,
					Action:         "Check if minikube is running using \"minikube status -p test_cluster\" command and try again",
				}},
			},
		},
		{
			description: "gke cluster",
			context:     "gke_test",
			err:         fmt.Errorf("exiting dev mode because first deploy failed: unable to connect to Kubernetes: Get \"https://192.168.64.3:8443/version?timeout=32s\": net/http: TLS handshake timeout"),
			expected:    "Deploy Failed. Could not connect to cluster due to \"https://192.168.64.3:8443/version?timeout=32s\": net/http: TLS handshake timeout. Check your connection for the cluster.",
			expectedAE: &proto.ActionableErr{
				ErrCode: proto.StatusCode_DEPLOY_CLUSTER_CONNECTION_ERR,
				Message: "exiting dev mode because first deploy failed: unable to connect to Kubernetes: Get \"https://192.168.64.3:8443/version?timeout=32s\": net/http: TLS handshake timeout",
				Suggestions: []*proto.Suggestion{{
					SuggestionCode: proto.SuggestionCode_CHECK_CLUSTER_CONNECTION,
					Action:         "Check your connection for the cluster",
				}},
			},
		},
		{
			description: "internal system error for minikube cluster",
			context:     "minikube",
			isMinikube:  true,
			err:         fmt.Errorf("Error: (Internal Server Error: the server is currently unable to handle the request)"),
			expected:    "Deploy Failed. Error: (Internal Server Error: the server is currently unable to handle the request). Check if minikube is running using \"minikube status\" command and try again or open an issue at https://github.com/GoogleContainerTools/skaffold/issues/new.",
			expectedAE: &proto.ActionableErr{
				ErrCode: proto.StatusCode_DEPLOY_CLUSTER_INTERNAL_SYSTEM_ERR,
				Message: "Error: (Internal Server Error: the server is currently unable to handle the request)",
				Suggestions: []*proto.Suggestion{{
					SuggestionCode: proto.SuggestionCode_CHECK_MINIKUBE_STATUS,
					Action:         "Check if minikube is running using \"minikube status\" command and try again",
				},
					{
						SuggestionCode: proto.SuggestionCode_OPEN_ISSUE,
						Action:         fmt.Sprintf("open an issue at %s", constants.GithubIssueLink),
					},
				},
			},
		},
		{
			description: "internal system error for k8s cluster",
			context:     "test_cluster",
			err:         fmt.Errorf("Error: (Internal Server Error: the server is currently unable to handle the request)"),
			expected:    "Deploy Failed. Error: (Internal Server Error: the server is currently unable to handle the request). Something went wrong with your cluster \"test_cluster\". Try again.\nIf this keeps happening please open an issue at https://github.com/GoogleContainerTools/skaffold/issues/new.",
			expectedAE: &proto.ActionableErr{
				ErrCode: proto.StatusCode_DEPLOY_CLUSTER_INTERNAL_SYSTEM_ERR,
				Message: "Error: (Internal Server Error: the server is currently unable to handle the request)",
				Suggestions: []*proto.Suggestion{{
					SuggestionCode: proto.SuggestionCode_OPEN_ISSUE,
					Action:         "Something went wrong with your cluster \"test_cluster\". Try again.\nIf this keeps happening please open an issue at https://github.com/GoogleContainerTools/skaffold/issues/new",
				}},
			},
		},
	}
	for _, test := range tests {
		testutil.Run(t, test.description, func(t *testutil.T) {
			cfg := mockConfig{kubeContext: test.context}
			if test.isMinikube {
				cfg.minikube = test.context
			}
			actual := sErrors.ShowAIError(&cfg, test.err)
			fmt.Println(actual.Error(), "\n", test.expected)
			t.CheckDeepEqual(test.expected, actual.Error())
			actualAE := sErrors.ActionableErr(&cfg, constants.Deploy, test.err)
			t.CheckDeepEqual(test.expectedAE, actualAE)
		})
	}
}

type mockConfig struct {
	minikube    string
	kubeContext string
}

func (m mockConfig) MinikubeProfile() string                { return m.minikube }
func (m mockConfig) GetPipelines() []latest.Pipeline        { return []latest.Pipeline{} }
func (m mockConfig) GetWorkingDir() string                  { return "" }
func (m mockConfig) GlobalConfig() string                   { return "" }
func (m mockConfig) ConfigurationFile() string              { return "" }
func (m mockConfig) DefaultRepo() *string                   { return &m.minikube }
func (m mockConfig) SkipRender() bool                       { return true }
func (m mockConfig) Prune() bool                            { return true }
func (m mockConfig) GetKubeContext() string                 { return m.kubeContext }
func (m mockConfig) GetInsecureRegistries() map[string]bool { return map[string]bool{} }
func (m mockConfig) Mode() config.RunMode                   { return config.RunModes.Dev }
