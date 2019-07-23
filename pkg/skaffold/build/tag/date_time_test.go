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

package tag

import (
	"testing"
	"time"

	"github.com/GoogleContainerTools/skaffold/testutil"
)

func TestDateTime_GenerateFullyQualifiedImageName(t *testing.T) {
	aLocalTimeStamp := time.Date(2015, 03, 07, 11, 06, 39, 123456789, time.Local)
	localZone, _ := aLocalTimeStamp.Zone()

	tests := []struct {
		description string
		format      string
		buildTime   time.Time
		timezone    string
		imageName   string
		image       string
		want        string
	}{
		{
			description: "default formatter",
			buildTime:   aLocalTimeStamp,
			imageName:   "test_image",
			want:        "test_image:2015-03-07_11-06-39.123_" + localZone,
		},
		{
			description: "user provided timezone",
			buildTime:   time.Unix(1234, 123456789),
			timezone:    "UTC",
			imageName:   "test_image",
			want:        "test_image:1970-01-01_00-20-34.123_UTC",
		},
		{
			description: "user provided format",
			buildTime:   aLocalTimeStamp,
			format:      "2006-01-02",
			imageName:   "test_image",
			want:        "test_image:2015-03-07",
		},
	}
	for _, test := range tests {
		testutil.Run(t, test.description, func(t *testutil.T) {
			c := &dateTimeTagger{
				Format:   test.format,
				TimeZone: test.timezone,
				timeFn:   func() time.Time { return test.buildTime },
			}

			tag, err := c.GenerateFullyQualifiedImageName(".", test.imageName)

			t.CheckNoError(err)
			t.CheckDeepEqual(test.want, tag)
		})
	}
}
