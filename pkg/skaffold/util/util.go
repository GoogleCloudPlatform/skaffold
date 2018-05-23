/*
Copyright 2018 Google LLC

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

package util

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/docker/distribution/reference"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

var Fs = afero.NewOsFs()

func RandomID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", b)
}

// These are the supported file formats for kubernetes manifests
var validSuffixes = []string{".yml", ".yaml", ".json"}

func ManifestFiles(manifests []string) ([]string, error) {
	list, err := ExpandPathsGlob(manifests)
	if err != nil {
		return nil, errors.Wrap(err, "expanding kubectl manifest paths")
	}

	var filteredManifests []string
	for _, f := range list {
		if !IsSupportedKubernetesFormat(f) {
			if !StrSliceContains(manifests, f) {
				logrus.Infof("refusing to deploy/delete non {json, yaml} file %s", f)
				logrus.Info("If you still wish to deploy this file, please specify it directly, outside a glob pattern.")
				continue
			}
		}
		filteredManifests = append(filteredManifests, f)
	}

	return filteredManifests, nil
}

// IsSupportedKubernetesFormat is for determining if a file under a glob pattern
// is deployable file format. It makes no attempt to check whether or not the file
// is actually deployable or has the correct contents.
func IsSupportedKubernetesFormat(n string) bool {
	for _, s := range validSuffixes {
		if strings.HasSuffix(n, s) {
			return true
		}
	}
	return false
}

func StrSliceContains(sl []string, s string) bool {
	for _, a := range sl {
		if a == s {
			return true
		}
	}
	return false
}

// ExpandPathsGlob expands paths according to filepath.Glob patterns
// Returns a list of unique files that match the glob patterns passed in.
func ExpandPathsGlob(paths []string) ([]string, error) {
	expandedPaths := make(map[string]bool)
	for _, p := range paths {
		if _, err := Fs.Stat(p); err == nil {
			// This is a file reference, so just add it
			expandedPaths[p] = true
			continue
		}

		files, err := afero.Glob(Fs, p)
		if err != nil {
			return nil, errors.Wrap(err, "glob")
		}
		if files == nil {
			return nil, fmt.Errorf("File pattern must match at least one file %s", p)
		}

		for _, f := range files {
			err := afero.Walk(Fs, f, func(path string, info os.FileInfo, err error) error {
				if !info.IsDir() {
					expandedPaths[path] = true
				}

				return nil
			})
			if err != nil {
				return nil, errors.Wrap(err, "filepath walk")
			}
		}
	}

	var ret []string
	for k := range expandedPaths {
		ret = append(ret, k)
	}
	sort.Strings(ret)
	return ret, nil
}

// BoolPtr returns a pointer to a bool
func BoolPtr(b bool) *bool {
	o := b
	return &o
}

func ReadConfiguration(filename string) ([]byte, error) {
	switch {
	case filename == "":
		return nil, errors.New("filename not specified")
	case filename == "-":
		return ioutil.ReadAll(os.Stdin)
	case strings.HasPrefix(filename, "http://") || strings.HasPrefix(filename, "https://"):
		return download(filename)
	default:
		return ioutil.ReadFile(filename)
	}
}

func download(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

type imageReference struct {
	BaseName       string
	FullyQualified bool
}

func ParseReference(image string) (*imageReference, error) {
	r, err := reference.Parse(image)
	if err != nil {
		return nil, err
	}

	baseName := image
	if n, ok := r.(reference.Named); ok {
		baseName = n.Name()
	}

	fullyQualified := false
	switch n := r.(type) {
	case reference.Tagged:
		fullyQualified = n.Tag() != "latest"
	case reference.Digested:
		fullyQualified = true
	}

	return &imageReference{
		BaseName:       baseName,
		FullyQualified: fullyQualified,
	}, nil
}
