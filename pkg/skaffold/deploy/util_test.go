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

package deploy

import (
	"bufio"
	"bytes"
	"fmt"
	"testing"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/deploy/kubectl"
	"github.com/GoogleContainerTools/skaffold/testutil"
)

var (
	deploymentYaml = `apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: skaffold-helm
  labels:
    app: skaffold-helm
    chart: skaffold-helm-0.1.0
    release: skaffold-helm
    heritage: Tiller
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: skaffold-helm
        release: skaffold-helm
    spec:
      containers:
        - name: skaffold-helm
          image: gcr.io/k8s-skaffold/skaffold-helm:latest@sha256:f363d1c3d6c724dbfa9fda36207a2ac699686729bb98e00a68e8dd23f3efa94d
          imagePullPolicy:
          ports:
            - containerPort: 80
          resources:
            {}
`
	serviceYaml = `apiVersion: v1
kind: Service
metadata:
  name: skaffold-helm-skaffold-helm
  labels:
    app: skaffold-helm
    chart: skaffold-helm-0.1.0
    release: skaffold-helm
    heritage: Tiller
spec:
  type: ClusterIP
  ports:
    - port: 80
      targetPort: 80
      protocol: TCP
      name: nginx
  selector:
    app: skaffold-helm
    release: skaffold-helm
`
	podYaml = `apiVersion: v1
kind: Pod
metadata:
  name: pod
  labels:
    app: skaffold-helm
spec:
  container:
    image: some
`
)

func TestParseReleaseInfo(t *testing.T) {
	var tests = []struct {
		description  string
		releaseBytes []byte
		expected     [][]byte
	}{
		{
			description:  "no bytes",
			releaseBytes: []byte{},
			expected:     [][]byte{},
		},
		{
			description: "valid service and deployment yaml",
			releaseBytes: []byte(fmt.Sprintf(`
---
# Source: skaffold-helm/templates/service.yaml
%s---
# Source: skaffold-helm/templates/deployment.yaml
%s---`, serviceYaml, deploymentYaml)),
			expected: [][]byte{
				[]byte(serviceYaml),
				[]byte(deploymentYaml)},
		},
		{
			description: "invalid manifest ignored",
			releaseBytes: []byte(fmt.Sprintf(`
Release info
some other info
---
# Source: skaffold-helm/templates/deployment.yaml
%s---
`, serviceYaml)),
			expected: [][]byte{[]byte(serviceYaml)},
		},
		{
			description: "invalid manifest yaml",
			releaseBytes: []byte(`
apiVersi
`),
			expected: [][]byte{},
		},
		{
			description: "valid service yaml",
			releaseBytes: []byte(fmt.Sprintf(`
---
# Source: skaffold-helm/templates/service.yaml
%s---`, serviceYaml)),
			expected: [][]byte{
				[]byte(serviceYaml),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			var expected kubectl.ManifestList
			for _, b := range test.expected {
				expected.Append(b)
			}
			actual := parseReleaseInfo(bufio.NewReader(bytes.NewBuffer(test.releaseBytes)))
			if len(actual) != len(expected) {
				t.Errorf("want %d artifacts, got %d", len(expected), len(actual))
			} else {
				for i := 0; i < len(actual); i++ {
					testutil.CheckDeepEqual(t, string(expected[i]), string(actual[i]))
				}
			}
		})
	}
}
