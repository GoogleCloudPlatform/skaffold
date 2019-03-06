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

package debugging

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/GoogleContainerTools/skaffold/testutil"
	"github.com/google/go-cmp/cmp"
)

func TestExtractInspectArg(t *testing.T) {
	tests := []struct {
		in     string
		result *inspectSpec
	}{
		{"", nil},
		{"foo", nil},
		{"--foo", nil},
		{"-inspect", nil},
		{"-inspect=9329", nil},
		{"--inspect", &inspectSpec{port: 9229, brk: false}},
		{"--inspect=9329", &inspectSpec{port: 9329, brk: false}},
		{"--inspect=:9329", &inspectSpec{port: 9329, brk: false}},
		{"--inspect=foo:9329", &inspectSpec{host: "foo", port: 9329, brk: false}},
		{"--inspect-brk", &inspectSpec{port: 9229, brk: true}},
		{"--inspect-brk=9329", &inspectSpec{port: 9329, brk: true}},
		{"--inspect-brk=:9329", &inspectSpec{port: 9329, brk: true}},
		{"--inspect-brk=foo:9329", &inspectSpec{host: "foo", port: 9329, brk: true}},
	}
	for _, test := range tests {
		t.Run(test.in, func(t *testing.T) {
			if test.result == nil {
				testutil.CheckEqual(t, nil, test.result, extractInspectArg(test.in))
			} else {
				testutil.CheckEqual(t, cmp.Options{cmp.AllowUnexported(inspectSpec{})}, *test.result, *extractInspectArg(test.in))
			}
		})
	}
}

func TestNodeTransformer_IsApplicable(t *testing.T) {
	tests := []struct {
		description string
		source      imageConfiguration
		result      bool
	}{
		{
			description: "NODE_VERSION",
			source:      imageConfiguration{env: map[string]string{"NODE_VERSION": "10"}},
			result:      true,
		},
		{
			description: "entrypoint node",
			source:      imageConfiguration{entrypoint: []string{"node", "init.js"}},
			result:      true,
		},
		{
			description: "entrypoint /usr/bin/node",
			source:      imageConfiguration{entrypoint: []string{"/usr/bin/node", "init.js"}},
			result:      true,
		},
		{
			description: "no entrypoint, args node",
			source:      imageConfiguration{arguments: []string{"node", "init.js"}},
			result:      true,
		},
		{
			description: "no entrypoint, arguments /usr/bin/node",
			source:      imageConfiguration{arguments: []string{"/usr/bin/node", "init.js"}},
			result:      true,
		},
		{
			description: "entrypoint nodemon",
			source:      imageConfiguration{entrypoint: []string{"nodemon", "init.js"}},
			result:      true,
		},
		{
			description: "entrypoint /usr/bin/nodemon",
			source:      imageConfiguration{entrypoint: []string{"/usr/bin/nodemon", "init.js"}},
			result:      true,
		},
		{
			description: "no entrypoint, args nodemon",
			source:      imageConfiguration{arguments: []string{"nodemon", "init.js"}},
			result:      true,
		},
		{
			description: "no entrypoint, arguments /usr/bin/nodemon",
			source:      imageConfiguration{arguments: []string{"/usr/bin/nodemon", "init.js"}},
			result:      true,
		},
		{
			description: "entrypoint /bin/sh",
			source:      imageConfiguration{entrypoint: []string{"/bin/sh"}},
			result:      false,
		},
		{
			description: "nothing",
			source:      imageConfiguration{},
			result:      false,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result := nodeTransformer{}.IsApplicable(test.source)
			testutil.CheckDeepEqual(t, test.result, result)
		})
	}
}

func TestNodeTransformerApply(t *testing.T) {
	tests := []struct {
		description   string
		containerSpec v1.Container
		configuration imageConfiguration
		result        v1.Container
	}{
		{
			description:   "empty",
			containerSpec: v1.Container{},
			configuration: imageConfiguration{},
			result:        v1.Container{},
		},
		{
			description:   "basic",
			containerSpec: v1.Container{},
			configuration: imageConfiguration{entrypoint: []string{"node"}},
			result: v1.Container{
				Command: []string{"node", "--inspect=9229"},
				Ports:   []v1.ContainerPort{v1.ContainerPort{Name: "devtools", ContainerPort: 9229}},
			},
		},
		{
			description: "existing port",
			containerSpec: v1.Container{
				Ports: []v1.ContainerPort{v1.ContainerPort{Name: "http-server", ContainerPort: 8080}},
			},
			configuration: imageConfiguration{entrypoint: []string{"node"}},
			result: v1.Container{
				Command: []string{"node", "--inspect=9229"},
				Ports:   []v1.ContainerPort{v1.ContainerPort{Name: "http-server", ContainerPort: 8080}, v1.ContainerPort{Name: "devtools", ContainerPort: 9229}},
			},
		},
		{
			description:   "command not entrypoint",
			containerSpec: v1.Container{},
			configuration: imageConfiguration{arguments: []string{"node"}},
			result: v1.Container{
				Args:  []string{"node", "--inspect=9229"},
				Ports: []v1.ContainerPort{v1.ContainerPort{Name: "devtools", ContainerPort: 9229}},
			},
		},
	}
	var identity portAllocator = func(port int32) int32 {
		return port
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			nodeTransformer{}.Apply(&test.containerSpec, test.configuration, identity)
			testutil.CheckDeepEqual(t, test.result, test.containerSpec)
		})
	}
}

func TestTransformManifestNodeJS(t *testing.T) {
	int32p := func(x int32) *int32 { return &x }
	tests := []struct {
		description string
		in          runtime.Object
		transformed bool
		out         runtime.Object
	}{
		{
			"Pod with no transformable container",
			&v1.Pod{
				Spec: v1.PodSpec{Containers: []v1.Container{
					v1.Container{
						Name:    "test",
						Command: []string{"echo", "Hello World"},
					},
				}}},
			false,
			&v1.Pod{
				Spec: v1.PodSpec{Containers: []v1.Container{
					v1.Container{
						Name:    "test",
						Command: []string{"echo", "Hello World"},
					},
				}}},
		},
		{
			"Pod with NodeJS container",
			&v1.Pod{
				Spec: v1.PodSpec{Containers: []v1.Container{
					v1.Container{
						Name:    "test",
						Command: []string{"node", "foo.js"},
					},
				}}},
			true,
			&v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{"debug.cloud.google.com/config": `{"test":{"devtools":9229,"runtime":"nodejs"}}`},
				},
				Spec: v1.PodSpec{Containers: []v1.Container{
					v1.Container{
						Name:    "test",
						Command: []string{"node", "--inspect=9229", "foo.js"},
						Ports:   []v1.ContainerPort{v1.ContainerPort{Name: "devtools", ContainerPort: 9229}},
					},
				}}},
		},
		{
			"Deployment with NodeJS container",
			&appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Replicas: int32p(2),
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{Containers: []v1.Container{
							v1.Container{
								Name:    "test",
								Command: []string{"node", "foo.js"},
							},
						}}}}},
			true,
			&appsv1.Deployment{
				//ObjectMeta: metav1.ObjectMeta{
				//	Labels: map[string]string{"debug.cloud.google.com/enabled": `yes`},
				//},
				Spec: appsv1.DeploymentSpec{
					Replicas: int32p(1),
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{"debug.cloud.google.com/config": `{"test":{"devtools":9229,"runtime":"nodejs"}}`},
						},
						Spec: v1.PodSpec{Containers: []v1.Container{
							v1.Container{
								Name:    "test",
								Command: []string{"node", "--inspect=9229", "foo.js"},
								Ports:   []v1.ContainerPort{v1.ContainerPort{Name: "devtools", ContainerPort: 9229}},
							},
						}}}}},
		},
		{
			"ReplicaSet with NodeJS container",
			&appsv1.ReplicaSet{
				Spec: appsv1.ReplicaSetSpec{
					Replicas: int32p(2),
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{Containers: []v1.Container{
							v1.Container{
								Name:    "test",
								Command: []string{"node", "foo.js"},
							},
						}}}}},
			true,
			&appsv1.ReplicaSet{
				//ObjectMeta: metav1.ObjectMeta{
				//	Labels: map[string]string{"debug.cloud.google.com/enabled": `yes`},
				//},
				Spec: appsv1.ReplicaSetSpec{
					Replicas: int32p(1),
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{"debug.cloud.google.com/config": `{"test":{"devtools":9229,"runtime":"nodejs"}}`},
						},
						Spec: v1.PodSpec{Containers: []v1.Container{
							v1.Container{
								Name:    "test",
								Command: []string{"node", "--inspect=9229", "foo.js"},
								Ports:   []v1.ContainerPort{v1.ContainerPort{Name: "devtools", ContainerPort: 9229}},
							},
						}}}}},
		},
		{
			"StatefulSet with NodeJS container",
			&appsv1.StatefulSet{
				Spec: appsv1.StatefulSetSpec{
					Replicas: int32p(2),
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{Containers: []v1.Container{
							v1.Container{
								Name:    "test",
								Command: []string{"node", "foo.js"},
							},
						}}}}},
			true,
			&appsv1.StatefulSet{
				//ObjectMeta: metav1.ObjectMeta{
				//	Labels: map[string]string{"debug.cloud.google.com/enabled": `yes`},
				//},
				Spec: appsv1.StatefulSetSpec{
					Replicas: int32p(1),
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{"debug.cloud.google.com/config": `{"test":{"devtools":9229,"runtime":"nodejs"}}`},
						},
						Spec: v1.PodSpec{Containers: []v1.Container{
							v1.Container{
								Name:    "test",
								Command: []string{"node", "--inspect=9229", "foo.js"},
								Ports:   []v1.ContainerPort{v1.ContainerPort{Name: "devtools", ContainerPort: 9229}},
							},
						}}}}},
		},
		{
			"DaemonSet with NodeJS container",
			&appsv1.DaemonSet{
				Spec: appsv1.DaemonSetSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{Containers: []v1.Container{
							v1.Container{
								Name:    "test",
								Command: []string{"node", "foo.js"},
							},
						}}}}},
			true,
			&appsv1.DaemonSet{
				//ObjectMeta: metav1.ObjectMeta{
				//	Labels: map[string]string{"debug.cloud.google.com/enabled": `yes`},
				//},
				Spec: appsv1.DaemonSetSpec{
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{"debug.cloud.google.com/config": `{"test":{"devtools":9229,"runtime":"nodejs"}}`},
						},
						Spec: v1.PodSpec{Containers: []v1.Container{
							v1.Container{
								Name:    "test",
								Command: []string{"node", "--inspect=9229", "foo.js"},
								Ports:   []v1.ContainerPort{v1.ContainerPort{Name: "devtools", ContainerPort: 9229}},
							},
						}}}}},
		},
		{
			"Job with NodeJS container",
			&batchv1.Job{
				Spec: batchv1.JobSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{Containers: []v1.Container{
							v1.Container{
								Name:    "test",
								Command: []string{"node", "foo.js"},
							},
						}}}}},
			true,
			&batchv1.Job{
				//ObjectMeta: metav1.ObjectMeta{
				//	Labels: map[string]string{"debug.cloud.google.com/enabled": `yes`},
				//},
				Spec: batchv1.JobSpec{
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{"debug.cloud.google.com/config": `{"test":{"devtools":9229,"runtime":"nodejs"}}`},
						},
						Spec: v1.PodSpec{Containers: []v1.Container{
							v1.Container{
								Name:    "test",
								Command: []string{"node", "--inspect=9229", "foo.js"},
								Ports:   []v1.ContainerPort{v1.ContainerPort{Name: "devtools", ContainerPort: 9229}},
							},
						}}}}},
		},
		{
			"ReplicationController with NodeJS container",
			&v1.ReplicationController{
				Spec: v1.ReplicationControllerSpec{
					Replicas: int32p(2),
					Template: &v1.PodTemplateSpec{
						Spec: v1.PodSpec{Containers: []v1.Container{
							v1.Container{
								Name:    "test",
								Command: []string{"node", "foo.js"},
							},
						}}}}},
			true,
			&v1.ReplicationController{
				//ObjectMeta: metav1.ObjectMeta{
				//	Labels: map[string]string{"debug.cloud.google.com/enabled": `yes`},
				//},
				Spec: v1.ReplicationControllerSpec{
					Replicas: int32p(1),
					Template: &v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{"debug.cloud.google.com/config": `{"test":{"devtools":9229,"runtime":"nodejs"}}`},
						},
						Spec: v1.PodSpec{Containers: []v1.Container{
							v1.Container{
								Name:    "test",
								Command: []string{"node", "--inspect=9229", "foo.js"},
								Ports:   []v1.ContainerPort{v1.ContainerPort{Name: "devtools", ContainerPort: 9229}},
							},
						}}}}},
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			value := test.in.DeepCopyObject()

			retriever := func(image string) (imageConfiguration, error) {
				return imageConfiguration{}, nil
			}
			result := transformManifest(value, retriever)
			testutil.CheckDeepEqual(t, test.transformed, result)
			testutil.CheckDeepEqual(t, test.out, value)
		})
	}
}
