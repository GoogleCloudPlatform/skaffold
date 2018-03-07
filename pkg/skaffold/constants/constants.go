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

package constants

import (
	"github.com/sirupsen/logrus"
)

const (
	// DefaultLogLevel is the default global verbosity
	DefaultLogLevel = logrus.ErrorLevel

	// The dockerfile path is given relative to the context directory
	DefaultDockerfilePath = "Dockerfile"

	DefaultDevTagStrategy = TagStrategySha256
	DefaultRunTagStrategy = TagStrategyGitCommit

	// TagStrategySha256 uses the checksum of the built artifact as the tag
	TagStrategySha256    = "sha256"
	TagStrategyGitCommit = "gitCommit"

	DefaultMinikubeContext         = "minikube"
	DefaultDockerForDesktopContext = "docker-for-desktop"
	GCSBucketSuffix                = "_cloudbuild"

	// TerminalBell is the sequence that triggers a beep in the terminal
	TerminalBell = "\007"
)
