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

package schema

import (
	"fmt"
	"testing"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/constants"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/defaults"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/util"
	"github.com/GoogleContainerTools/skaffold/testutil"
	"k8s.io/client-go/tools/clientcmd/api"
)

const (
	minimalConfig = ``

	simpleConfig = `
build:
  tagPolicy:
    gitCommit: {}
  artifacts:
  - image: example
deploy:
  kubectl: {}
`
	// This config has two tag policies set.
	invalidConfig = `
build:
  tagPolicy:
    sha256: {}
    gitCommit: {}
  artifacts:
  - image: example
deploy:
  name: example
`

	completeConfig = `
build:
  tagPolicy:
    sha256: {}
  artifacts:
  - image: image1
    context: ./examples/app1
    docker:
      dockerfile: Dockerfile.dev
  - image: image2
    context: ./examples/app2
    bazel:
      target: //:example.tar
  googleCloudBuild:
    projectId: ID
deploy:
  kubectl:
   manifests:
   - dep.yaml
   - svc.yaml
`
	minimalKanikoConfig = `
build:
  kaniko:
    buildContext:
      gcsBucket: demo
`
	completeKanikoConfig = `
build:
  kaniko:
    buildContext:
      gcsBucket: demo
    googleCloudConfig:
      path: /secret.json
      secretName: secret-name
    namespace: nskaniko
    timeout: 120m
    dockerConfig:
      secretName: config-name
      path: /kaniko/.docker
`
	badConfig = "bad config"
)

func TestParseConfig(t *testing.T) {
	cleanup := testutil.SetupFakeKubernetesContext(t, api.Config{CurrentContext: "cluster1"})
	defer cleanup()

	var tests = []struct {
		apiVersion  string
		description string
		config      string
		expected    util.VersionedConfig
		badReader   bool
		shouldErr   bool
	}{
		{
			apiVersion:  latest.Version,
			description: "Minimal config",
			config:      minimalConfig,
			expected: config(
				withLocalBuild(
					withGitTagger(),
				),
				withKubectlDeploy("k8s/*.yaml"),
			),
		},
		{
			apiVersion:  "skaffold/v1alpha1",
			description: "Old minimal config",
			config:      minimalConfig,
			expected: config(
				withLocalBuild(
					withGitTagger(),
				),
				withKubectlDeploy("k8s/*.yaml"),
			),
		},
		{
			apiVersion:  latest.Version,
			description: "Simple config",
			config:      simpleConfig,
			expected: config(
				withLocalBuild(
					withGitTagger(),
					withDockerArtifact("example", ".", "Dockerfile"),
				),
				withKubectlDeploy("k8s/*.yaml"),
			),
		},
		{
			apiVersion:  latest.Version,
			description: "Complete config",
			config:      completeConfig,
			expected: config(
				withGoogleCloudBuild("ID",
					withShaTagger(),
					withDockerArtifact("image1", "./examples/app1", "Dockerfile.dev"),
					withBazelArtifact("image2", "./examples/app2", "//:example.tar"),
				),
				withKubectlDeploy("dep.yaml", "svc.yaml"),
			),
		},
		{
			apiVersion:  latest.Version,
			description: "Minimal Kaniko config",
			config:      minimalKanikoConfig,
			expected: config(
				withKanikoBuild("demo", "kaniko-secret", "default", "", "20m",
					withGitTagger(),
				),
				withKubectlDeploy("k8s/*.yaml"),
			),
		},
		{
			apiVersion:  latest.Version,
			description: "Complete Kaniko config",
			config:      completeKanikoConfig,
			expected: config(
				withKanikoBuild("demo", "secret-name", "nskaniko", "/secret.json", "120m",
					withGitTagger(),
					withDockerConfig("config-name", "/kaniko/.docker"),
				),
				withKubectlDeploy("k8s/*.yaml"),
			),
		},
		{
			apiVersion:  latest.Version,
			description: "Bad config",
			config:      badConfig,
			shouldErr:   true,
		},
		{
			apiVersion:  latest.Version,
			description: "two taggers defined",
			config:      invalidConfig,
			shouldErr:   true,
		},
		{
			apiVersion:  "",
			description: "ApiVersion not specified",
			config:      minimalConfig,
			shouldErr:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			tmp, cleanup := testutil.NewTempDir(t)
			defer cleanup()

			yaml := fmt.Sprintf("apiVersion: %s\nkind: Config\n%s", test.apiVersion, test.config)
			tmp.Write("skaffold.yaml", yaml)

			cfg, err := ParseConfig(tmp.Path("skaffold.yaml"), true)
			if cfg != nil {
				config := cfg.(*latest.SkaffoldPipeline)
				if err := defaults.Set(config); err != nil {
					t.Fatal("unable to set default values")
				}
			}

			testutil.CheckErrorAndDeepEqual(t, test.shouldErr, err, test.expected, cfg)
		})
	}
}

func config(ops ...func(*latest.SkaffoldPipeline)) *latest.SkaffoldPipeline {
	cfg := &latest.SkaffoldPipeline{APIVersion: latest.Version, Kind: "Config"}
	for _, op := range ops {
		op(cfg)
	}
	return cfg
}

func withLocalBuild(ops ...func(*latest.BuildConfig)) func(*latest.SkaffoldPipeline) {
	return func(cfg *latest.SkaffoldPipeline) {
		b := latest.BuildConfig{BuildType: latest.BuildType{LocalBuild: &latest.LocalBuild{}}}
		for _, op := range ops {
			op(&b)
		}
		cfg.Build = b
	}
}

func withGoogleCloudBuild(id string, ops ...func(*latest.BuildConfig)) func(*latest.SkaffoldPipeline) {
	return func(cfg *latest.SkaffoldPipeline) {
		b := latest.BuildConfig{BuildType: latest.BuildType{GoogleCloudBuild: &latest.GoogleCloudBuild{
			ProjectID:   id,
			DockerImage: "gcr.io/cloud-builders/docker",
			MavenImage:  "gcr.io/cloud-builders/mvn@sha256:0ec283f2ee1ab1d2ac779dcbb24bddaa46275aec7088cc10f2926b4ea0fcac9b",
			GradleImage: "gcr.io/cloud-builders/gradle",
		}}}
		for _, op := range ops {
			op(&b)
		}
		cfg.Build = b
	}
}

func withKanikoBuild(bucket, secretName, namespace, secret string, timeout string, ops ...func(*latest.BuildConfig)) func(*latest.SkaffoldPipeline) {
	return func(cfg *latest.SkaffoldPipeline) {
		b := latest.BuildConfig{BuildType: latest.BuildType{KanikoBuild: &latest.KanikoBuild{
			BuildContext: &latest.KanikoBuildContext{
				GCSBucket: bucket,
			},
			GoogleCloudConfig: &latest.GoogleCloudConfig{
				Path:       secret,
				SecretName: secretName,
			},
			Namespace: namespace,
			Timeout:   timeout,
			Image:     constants.DefaultKanikoImage,
		}}}
		for _, op := range ops {
			op(&b)
		}
		cfg.Build = b
	}
}

func withDockerConfig(secretName string, path string) func(*latest.BuildConfig) {
	return func(cfg *latest.BuildConfig) {
		cfg.KanikoBuild.DockerConfig = &latest.DockerConfig{
			SecretName: secretName,
			Path:       path,
		}
	}
}

func withKubectlDeploy(manifests ...string) func(*latest.SkaffoldPipeline) {
	return func(cfg *latest.SkaffoldPipeline) {
		cfg.Deploy = latest.DeployConfig{
			DeployType: latest.DeployType{
				KubectlDeploy: &latest.KubectlDeploy{
					Manifests: manifests,
				},
			},
		}
	}
}

func withHelmDeploy() func(*latest.SkaffoldPipeline) {
	return func(cfg *latest.SkaffoldPipeline) {
		cfg.Deploy = latest.DeployConfig{
			DeployType: latest.DeployType{
				HelmDeploy: &latest.HelmDeploy{},
			},
		}
	}
}

func withDockerArtifact(image, workspace, dockerfile string) func(*latest.BuildConfig) {
	return func(cfg *latest.BuildConfig) {
		cfg.Artifacts = append(cfg.Artifacts, &latest.Artifact{
			ImageName: image,
			Workspace: workspace,
			ArtifactType: latest.ArtifactType{
				DockerArtifact: &latest.DockerArtifact{
					DockerfilePath: dockerfile,
				},
			},
		})
	}
}

func withBazelArtifact(image, workspace, target string) func(*latest.BuildConfig) {
	return func(cfg *latest.BuildConfig) {
		cfg.Artifacts = append(cfg.Artifacts, &latest.Artifact{
			ImageName: image,
			Workspace: workspace,
			ArtifactType: latest.ArtifactType{
				BazelArtifact: &latest.BazelArtifact{
					BuildTarget: target,
				},
			},
		})
	}
}

func withTagPolicy(tagPolicy latest.TagPolicy) func(*latest.BuildConfig) {
	return func(cfg *latest.BuildConfig) { cfg.TagPolicy = tagPolicy }
}

func withGitTagger() func(*latest.BuildConfig) {
	return withTagPolicy(latest.TagPolicy{GitTagger: &latest.GitTagger{}})
}

func withShaTagger() func(*latest.BuildConfig) {
	return withTagPolicy(latest.TagPolicy{ShaTagger: &latest.ShaTagger{}})
}

func withProfiles(profiles ...latest.Profile) func(*latest.SkaffoldPipeline) {
	return func(cfg *latest.SkaffoldPipeline) {
		cfg.Profiles = profiles
	}
}

func withTests(testCases ...*latest.TestCase) func(*latest.SkaffoldPipeline) {
	return func(cfg *latest.SkaffoldPipeline) {
		cfg.Test = testCases
	}
}

func withExecutionEnvironment(name latest.ExecEnvironment) func(*latest.BuildConfig) {
	return func(cfg *latest.BuildConfig) {
		cfg.ExecutionEnvironment = &latest.ExecutionEnvironment{
			Name: name,
		}
	}
}
func TestUpgradeToNextVersion(t *testing.T) {
	for i, schemaVersion := range SchemaVersions[0 : len(SchemaVersions)-2] {
		from := schemaVersion
		to := SchemaVersions[i+1]
		description := fmt.Sprintf("Upgrade from %s to %s", from.APIVersion, to.APIVersion)

		t.Run(description, func(t *testing.T) {
			factory, _ := SchemaVersions.Find(from.APIVersion)
			newer, err := factory().Upgrade()

			testutil.CheckErrorAndDeepEqual(t, false, err, to.APIVersion, newer.GetVersion())
		})
	}
}

func TestCantUpgradeFromLatestVersion(t *testing.T) {
	factory, present := SchemaVersions.Find(latest.Version)
	testutil.CheckDeepEqual(t, true, present)

	_, err := factory().Upgrade()
	testutil.CheckError(t, true, err)
}
