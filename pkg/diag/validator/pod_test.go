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

package validator

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fakekubeclientset "k8s.io/client-go/kubernetes/fake"

	"github.com/GoogleContainerTools/skaffold/proto"
	"github.com/GoogleContainerTools/skaffold/testutil"
)

func TestRun(t *testing.T) {
	type mockLogOutput struct {
		output []byte
		err    error
	}
	before := time.Now()
	after := before.Add(3 * time.Second)
	tests := []struct {
		description string
		pods        []*v1.Pod
		logOutput   mockLogOutput
		events      []v1.Event
		expected    []Resource
	}{
		{
			description: "pod don't exist in test namespace",
			pods: []*v1.Pod{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "foo-ns",
				},
				TypeMeta: metav1.TypeMeta{Kind: "Pod"}},
			},
			expected: nil,
		},
		{
			description: "pod is Waiting conditions with error",
			pods: []*v1.Pod{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "test",
				},
				TypeMeta: metav1.TypeMeta{Kind: "Pod"},
				Status: v1.PodStatus{
					Phase:      v1.PodPending,
					Conditions: []v1.PodCondition{{Type: v1.PodScheduled, Status: v1.ConditionTrue}},
					ContainerStatuses: []v1.ContainerStatus{
						{
							Name:  "foo-container",
							Image: "foo-image",
							State: v1.ContainerState{
								Waiting: &v1.ContainerStateWaiting{
									Reason:  "ErrImagePull",
									Message: "rpc error: code = Unknown desc = Error response from daemon: pull access denied for leeroy-web1, repository does not exist or may require 'docker login': denied: requested access to the resource is denied",
								},
							},
						},
					},
				},
			}},
			expected: []Resource{NewResource("test", "Pod", "foo", "Pending",
				fmt.Errorf("container foo-container is waiting to start: foo-image can't be pulled"),
				proto.StatusCode_STATUSCHECK_IMAGE_PULL_ERR, nil)},
		},
		{
			description: "pod is Waiting condition due to ErrImageBackOffPullErr",
			pods: []*v1.Pod{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "test",
				},
				TypeMeta: metav1.TypeMeta{Kind: "Pod"},
				Status: v1.PodStatus{
					Phase:      v1.PodPending,
					Conditions: []v1.PodCondition{{Type: v1.PodScheduled, Status: v1.ConditionTrue}},
					ContainerStatuses: []v1.ContainerStatus{
						{
							Name:  "foo-container",
							Image: "foo-image",
							State: v1.ContainerState{
								Waiting: &v1.ContainerStateWaiting{
									Reason:  "ErrImagePullBackOff",
									Message: "rpc error: code = Unknown desc = Error response from daemon: pull access denied for leeroy-web1, repository does not exist or may require 'docker login': denied: requested access to the resource is denied",
								},
							},
						},
					},
				},
			}},
			expected: []Resource{NewResource("test", "Pod", "foo", "Pending",
				fmt.Errorf("container foo-container is waiting to start: foo-image can't be pulled"),
				proto.StatusCode_STATUSCHECK_IMAGE_PULL_ERR, nil)},
		},
		{
			description: "pod is Waiting due to Image Backoff Pull error",
			pods: []*v1.Pod{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "test",
				},
				TypeMeta: metav1.TypeMeta{Kind: "Pod"},
				Status: v1.PodStatus{
					Phase:      v1.PodPending,
					Conditions: []v1.PodCondition{{Type: v1.PodScheduled, Status: v1.ConditionTrue}},
					ContainerStatuses: []v1.ContainerStatus{
						{
							Name:  "foo-container",
							Image: "foo-image",
							State: v1.ContainerState{
								Waiting: &v1.ContainerStateWaiting{
									Reason:  "ImagePullBackOff",
									Message: "rpc error: code = Unknown desc = Error response from daemon: pull access denied for leeroy-web1, repository does not exist or may require 'docker login': denied: requested access to the resource is denied",
								},
							},
						},
					},
				},
			}},
			expected: []Resource{NewResource("test", "Pod", "foo", "Pending",
				fmt.Errorf("container foo-container is waiting to start: foo-image can't be pulled"),
				proto.StatusCode_STATUSCHECK_IMAGE_PULL_ERR, nil)},
		},
		{
			description: "pod is in Terminated State",
			pods: []*v1.Pod{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "test",
				},
				TypeMeta: metav1.TypeMeta{Kind: "Pod"},
				Status: v1.PodStatus{
					Phase:      v1.PodSucceeded,
					Conditions: []v1.PodCondition{{Type: v1.PodScheduled, Status: v1.ConditionTrue}},
				},
			}},
			expected: []Resource{NewResource("test", "Pod", "foo", "Succeeded", nil,
				proto.StatusCode_STATUSCHECK_SUCCESS, nil)},
		},
		{
			description: "pod is in Stable State",
			pods: []*v1.Pod{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "test",
				},
				TypeMeta: metav1.TypeMeta{Kind: "Pod"},
				Status: v1.PodStatus{
					Phase:      v1.PodRunning,
					Conditions: []v1.PodCondition{{Type: v1.PodScheduled, Status: v1.ConditionTrue}},
					ContainerStatuses: []v1.ContainerStatus{
						{
							Name:  "foo-container",
							State: v1.ContainerState{Running: &v1.ContainerStateRunning{}},
						},
					},
				},
			}},
			expected: []Resource{NewResource("test", "Pod", "foo", "Running", nil,
				proto.StatusCode_STATUSCHECK_SUCCESS, nil)},
		},
		{
			description: "pod condition unknown",
			pods: []*v1.Pod{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "test",
				},
				TypeMeta: metav1.TypeMeta{Kind: "Pod"},
				Status: v1.PodStatus{
					Phase: v1.PodPending,
					Conditions: []v1.PodCondition{{
						Type:    v1.PodScheduled,
						Status:  v1.ConditionUnknown,
						Message: "could not determine",
					}},
				},
			}},
			expected: []Resource{NewResource("test", "Pod", "foo", "Pending",
				fmt.Errorf("could not determine"), proto.StatusCode_STATUSCHECK_UNKNOWN, nil)},
		},
		{
			description: "pod could not be scheduled",
			pods: []*v1.Pod{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "test",
				},
				TypeMeta: metav1.TypeMeta{Kind: "Pod"},
				Status: v1.PodStatus{
					Phase: v1.PodPending,
					Conditions: []v1.PodCondition{{
						Type:    v1.PodScheduled,
						Status:  v1.ConditionFalse,
						Reason:  v1.PodReasonUnschedulable,
						Message: "0/2 nodes are available: 1 node(s) had taint {node.kubernetes.io/disk-pressure: }, that the pod didn't tolerate, 1 node(s) had taint {node.kubernetes.io/unreachable: }, that the pod didn't tolerate",
					}},
				},
			}},
			expected: []Resource{NewResource("test", "Pod", "foo", "Pending",
				fmt.Errorf("Unschedulable: 0/2 nodes available: 1 node has disk pressure, 1 node is unreachable"),
				proto.StatusCode_STATUSCHECK_NODE_DISK_PRESSURE, nil)},
		},
		{
			description: "pod is running but container terminated",
			pods: []*v1.Pod{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "test",
				},
				TypeMeta: metav1.TypeMeta{Kind: "Pod"},
				Status: v1.PodStatus{
					Phase:      v1.PodRunning,
					Conditions: []v1.PodCondition{{Type: v1.PodScheduled, Status: v1.ConditionTrue}},
					ContainerStatuses: []v1.ContainerStatus{
						{
							Name:  "foo-container",
							State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{ExitCode: 1}},
						},
					},
				},
			}},
			logOutput: mockLogOutput{
				output: []byte("main.go:57 \ngo panic"),
			},
			expected: []Resource{NewResource("test", "Pod", "foo", "Running",
				fmt.Errorf("container foo-container terminated with exit code 1"),
				proto.StatusCode_STATUSCHECK_CONTAINER_TERMINATED, []string{
					"[foo foo-container] main.go:57 ",
					"[foo foo-container] go panic"},
			)},
		},
		{
			description: "pod is running but container terminated but could not retrieve logs",
			pods: []*v1.Pod{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "test",
				},
				Status: v1.PodStatus{
					Phase:      v1.PodRunning,
					Conditions: []v1.PodCondition{{Type: v1.PodScheduled, Status: v1.ConditionTrue}},
					ContainerStatuses: []v1.ContainerStatus{
						{
							Name:  "foo-container",
							State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{ExitCode: 1}},
						},
					},
				},
			}},
			logOutput: mockLogOutput{
				err: fmt.Errorf("error"),
			},
			expected: []Resource{NewResource("test", "pod", "foo", "Running",
				fmt.Errorf("container foo-container terminated with exit code 1"),
				proto.StatusCode_STATUSCHECK_CONTAINER_TERMINATED, []string{
					"Error retrieving logs for pod foo. Try `kubectl logs foo -n test -c foo-container`"},
			)},
		},
		// Events Test cases
		{
			description: "pod condition with events",
			pods: []*v1.Pod{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "test",
				},
				TypeMeta: metav1.TypeMeta{Kind: "Pod"},
				Status: v1.PodStatus{
					Phase: v1.PodPending,
					Conditions: []v1.PodCondition{{
						Type:    v1.PodScheduled,
						Status:  v1.ConditionUnknown,
						Message: "could not determine",
					}},
				},
			}},
			events: []v1.Event{
				{
					ObjectMeta: metav1.ObjectMeta{Namespace: "test"},
					Reason:     "eventCode", Type: "Warning", Message: "dummy event",
				},
			},
			expected: []Resource{NewResource("test", "Pod", "foo", "Pending",
				fmt.Errorf("eventCode: dummy event"), proto.StatusCode_STATUSCHECK_UNKNOWN_EVENT, nil)},
		},
		{
			description: "pod condition a warning event followed up normal event",
			pods: []*v1.Pod{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "test",
				},
				TypeMeta: metav1.TypeMeta{Kind: "Pod"},
				Status: v1.PodStatus{
					Phase: v1.PodPending,
					Conditions: []v1.PodCondition{{
						Type:    v1.PodScheduled,
						Status:  v1.ConditionUnknown,
						Message: "could not determine",
					}},
				},
			}},
			events: []v1.Event{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "one", Namespace: "test"},
					Reason:     "eventCode", Type: "Warning", Message: "dummy event",
					EventTime: metav1.MicroTime{Time: before},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "two", Namespace: "test"},
					Reason:     "Created", Type: "Normal", Message: "Container Created",
					EventTime: metav1.MicroTime{Time: after},
				},
			},
			expected: []Resource{NewResource("test", "Pod", "foo", "Pending",
				fmt.Errorf("eventCode: dummy event"), proto.StatusCode_STATUSCHECK_UNKNOWN_EVENT, nil)},
		},
		{
			description: "pod condition a normal event followed by a warning event",
			pods: []*v1.Pod{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "test",
				},
				TypeMeta: metav1.TypeMeta{Kind: "Pod"},
				Status: v1.PodStatus{
					Phase: v1.PodPending,
					Conditions: []v1.PodCondition{{
						Type:    v1.PodScheduled,
						Status:  v1.ConditionUnknown,
						Message: "could not determine",
					}},
				},
			}},
			events: []v1.Event{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "two", Namespace: "test"},
					Reason:     "Created", Type: "Normal", Message: "Container Created",
					EventTime: metav1.MicroTime{Time: before},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "one", Namespace: "test"},
					Reason:     "eventCode", Type: "Warning", Message: "dummy event",
					EventTime: metav1.MicroTime{Time: after},
				},
			},
			expected: []Resource{NewResource("test", "Pod", "foo", "Pending",
				fmt.Errorf("eventCode: dummy event"), proto.StatusCode_STATUSCHECK_UNKNOWN_EVENT, nil)},
		},
		{
			description: "pod condition a warning event followed up by warning adds last warning seen",
			pods: []*v1.Pod{{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "test",
				},
				TypeMeta: metav1.TypeMeta{Kind: "Pod"},
				Status: v1.PodStatus{
					Phase: v1.PodPending,
					Conditions: []v1.PodCondition{{
						Type:    v1.PodScheduled,
						Status:  v1.ConditionUnknown,
						Message: "could not determine",
					}},
				},
			}},
			events: []v1.Event{
				{
					ObjectMeta: metav1.ObjectMeta{Name: "two", Namespace: "test"}, Reason: "FailedScheduling", Type: "Warning",
					Message:   "0/1 nodes are available: 1 node(s) had taint {key: value}, that the pod didn't tolerate",
					EventTime: metav1.MicroTime{Time: after},
				},
				{
					ObjectMeta: metav1.ObjectMeta{Name: "one", Namespace: "test"},
					Reason:     "eventCode", Type: "Warning", Message: "dummy event",
					EventTime: metav1.MicroTime{Time: before},
				},
			},
			expected: []Resource{NewResource("test", "Pod", "foo", "Pending",
				fmt.Errorf("0/1 nodes are available: 1 node(s) had taint {key: value}, that the pod didn't tolerate"), proto.StatusCode_STATUSCHECK_FAILED_SCHEDULING, nil)},
		},
	}

	for _, test := range tests {
		testutil.Run(t, test.description, func(t *testutil.T) {
			rs := make([]runtime.Object, len(test.pods))
			mRun := func(n string, args []string) ([]byte, error) {
				actualCommand := strings.Join(append([]string{n}, args...), " ")
				if expected := "kubectl logs foo -n test -c foo-container"; actualCommand != expected {
					t.Errorf("got %s, expected %s", actualCommand, expected)
				}
				return test.logOutput.output, test.logOutput.err
			}
			t.Override(&runCli, mRun)
			for i, p := range test.pods {
				rs[i] = p
			}
			rs = append(rs, &v1.EventList{Items: test.events})
			f := fakekubeclientset.NewSimpleClientset(rs...)

			actual, err := NewPodValidator(f).Validate(context.Background(), "test", metav1.ListOptions{})
			t.CheckNoError(err)
			t.CheckDeepEqual(test.expected, actual, cmp.AllowUnexported(Resource{}), cmp.Comparer(func(x, y error) bool {
				if x == nil && y == nil {
					return true
				} else if x != nil && y != nil {
					return x.Error() == y.Error()
				}
				return false
			}))
		})
	}
}
