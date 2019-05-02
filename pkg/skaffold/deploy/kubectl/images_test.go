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

package kubectl

import (
	"testing"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/build"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/warnings"
	"github.com/GoogleContainerTools/skaffold/testutil"
)

func TestReplaceImages(t *testing.T) {
	manifests := ManifestList{[]byte(`
apiVersion: v1
kind: Pod
metadata:
  name: getting-started
spec:
  containers:
  - image: gcr.io/k8s-skaffold/example
    name: not-tagged
  - image: gcr.io/k8s-skaffold/example:latest
    name: latest
  - image: gcr.io/k8s-skaffold/example:v1
    name: fully-qualified
  - image: skaffold/other
    name: other
  - image: gcr.io/k8s-skaffold/example@sha256:81daf011d63b68cfa514ddab7741a1adddd59d3264118dfb0fd9266328bb8883
    name: digest
  - image: skaffold/usedbyfqn:TAG
  - image: skaffold/usedwrongfqn:OTHER
  - image: in valid
`)}

	builds := []build.Artifact{{
		ImageName: "gcr.io/k8s-skaffold/example",
		Tag:       "gcr.io/k8s-skaffold/example:TAG",
	}, {
		ImageName: "skaffold/other",
		Tag:       "skaffold/other:OTHER_TAG",
	}, {
		ImageName: "skaffold/unused",
		Tag:       "skaffold/unused:TAG",
	}, {
		ImageName: "skaffold/usedbyfqn",
		Tag:       "skaffold/usedbyfqn:TAG",
	}, {
		ImageName: "skaffold/usedwrongfqn",
		Tag:       "skaffold/usedwrongfqn:TAG",
	}}

	expected := ManifestList{[]byte(`
apiVersion: v1
kind: Pod
metadata:
  name: getting-started
spec:
  containers:
  - image: gcr.io/k8s-skaffold/example:TAG
    name: not-tagged
  - image: gcr.io/k8s-skaffold/example:TAG
    name: latest
  - image: gcr.io/k8s-skaffold/example:v1
    name: fully-qualified
  - image: skaffold/other:OTHER_TAG
    name: other
  - image: gcr.io/k8s-skaffold/example@sha256:81daf011d63b68cfa514ddab7741a1adddd59d3264118dfb0fd9266328bb8883
    name: digest
  - image: skaffold/usedbyfqn:TAG
  - image: skaffold/usedwrongfqn:OTHER
  - image: in valid
`)}

	testutil.Run(t, "", func(t *testutil.T) {
		fakeWarner := &warnings.Collect{}
		t.Override(&warnings.Printf, fakeWarner.Warnf)

		resultManifest, err := manifests.ReplaceImages(builds, "")

		t.CheckNoError(err)
		t.CheckDeepEqual(expected.String(), resultManifest.String())
		t.CheckDeepEqual([]string{
			"Couldn't parse image: in valid",
			"image [skaffold/unused] is not used by the deployment",
			"image [skaffold/usedwrongfqn] is not used by the deployment",
		}, fakeWarner.Warnings)
	})
}

func TestReplaceEmptyManifest(t *testing.T) {
	manifests := ManifestList{[]byte(""), []byte("  ")}
	expected := ManifestList{}

	resultManifest, err := manifests.ReplaceImages(nil, "")

	testutil.CheckErrorAndDeepEqual(t, false, err, expected.String(), resultManifest.String())
}

func TestReplaceInvalidManifest(t *testing.T) {
	manifests := ManifestList{[]byte("INVALID")}

	_, err := manifests.ReplaceImages(nil, "")

	testutil.CheckError(t, true, err)
}

func TestReplaceNonStringImageField(t *testing.T) {
	manifests := ManifestList{[]byte(`
apiVersion: v1
image:
- value1
- value2
`)}

	output, err := manifests.ReplaceImages(nil, "")

	testutil.CheckErrorAndDeepEqual(t, false, err, manifests.String(), output.String())
}

func TestReplaceImagerShouldReplaceForKind(t *testing.T) {
	tests := []struct {
		description string
		kind        string
		expect bool
	}{
		{
			description: "shd replace for pod",
			kind:        "pod",
			expect: true,
		},
		{
			description: "shd replace for POD (case ignore equal)",
			kind:        "POD",
			expect:      true,
		},
		{
			description: "shd replace for any other kind",
			kind:        "random",
			expect: true,
		},
		{
			description: "shd not replace if kind not set",
			kind:        "",
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			l := newImageReplacer(nil, "")
			l.SetKind(test.kind)
			if l.ShouldReplaceForKind() != test.expect {
				t.Errorf("expected to see %t for %s. found %t", test.expect, test.kind, !test.expect)
			}
		})
	}
}