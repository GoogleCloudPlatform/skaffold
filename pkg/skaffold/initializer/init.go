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

package initializer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/GoogleContainerTools/skaffold/cmd/skaffold/app/tips"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/config"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/docker"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/initializer/kubectl"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/jib"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/defaults"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	survey "gopkg.in/AlecAivazis/survey.v1"
	yaml "gopkg.in/yaml.v2"
)

// NoBuilder allows users to specify they don't want to build
// an image we parse out from a kubernetes manifest
const NoBuilder = "None (image not built from these sources)"

// Initializer is the Init API of skaffold and responsible for generating
// skaffold configuration file.
type Initializer interface {
	// GenerateDeployConfig generates Deploy Config for skaffold configuration.
	GenerateDeployConfig() latest.DeployConfig
	// GetImages fetches all the images defined in the manifest files.
	GetImages() []string
}

// InitBuilder represents a builder that can be chosen by skaffold init.
type InitBuilder interface {
	// getPrompt returns the initBuilder's string representation, used when prompting the user to choose a builder.
	GetPrompt() string
	// getArtifact returns the Artifact used to generate the Build Config.
	GetArtifact(image string) *latest.Artifact
	// getConfiguredImage returns the target image configured by the builder
	GetConfiguredImage() string
}

// Config defines the Initializer Config for Init API of skaffold.
type Config struct {
	ComposeFile  string
	CliArtifacts []string
	SkipBuild    bool
	Force        bool
	Analyze      bool
	Opts         *config.SkaffoldOptions
}

// DoInit executes the `skaffold init` flow.
func DoInit(out io.Writer, c Config) error {
	rootDir := "."

	if c.ComposeFile != "" {
		// run kompose first to generate k8s manifests, then run skaffold init
		logrus.Infof("running 'kompose convert' for file %s", c.ComposeFile)
		komposeCmd := exec.Command("kompose", "convert", "-f", c.ComposeFile)
		if err := util.RunCmd(komposeCmd); err != nil {
			return errors.Wrap(err, "running kompose")
		}
	}

	potentialConfigs, buildConfigs, err := walk(rootDir, c.Force, detectBuildFile)
	if err != nil {
		return err
	}

	k, err := kubectl.New(potentialConfigs)
	if err != nil {
		return err
	}
	images := k.GetImages()
	if c.Analyze {
		return printAnalyzeJSON(out, c.SkipBuild, buildConfigs, images)
	}
	// conditionally generate build artifacts
	var pairs, newPairs []buildConfigPair
	filteredImages := []string{}
	if !c.SkipBuild {
		if len(buildConfigs) == 0 {
			return errors.New("one or more valid builder configuration (Dockerfile or Jib configuration) must be present to build images with skaffold; please provide at least one build config and try again or run `skaffold init --skip-build`")
		}

		// Auto-select builders that have a definite target image
		for _, image := range images {
			matchingConfigIndex := -1
			for i, config := range buildConfigs {
				if image != config.GetConfiguredImage() {
					continue
				}

				// Found more than one match; can't auto-select.
				if matchingConfigIndex != -1 {
					matchingConfigIndex = -1
					break
				}
				matchingConfigIndex = i
			}

			if matchingConfigIndex != -1 {
				// Exactly one pair found
				pairs = append(pairs, buildConfigPair{ImageName: image, BuildConfig: buildConfigs[matchingConfigIndex]})
				buildConfigs = append(buildConfigs[:matchingConfigIndex], buildConfigs[matchingConfigIndex+1:]...)
			} else {
				// No definite pair found, add to images list
				filteredImages = append(filteredImages, image)
			}
		}

		if c.CliArtifacts != nil {
			newPairs, err = processCliArtifacts(c.CliArtifacts)
			if err != nil {
				return errors.Wrap(err, "processing cli artifacts")
			}
		} else {
			newPairs = resolveDockerfileImages(buildConfigs, filteredImages)
		}
	}

	pairs = append(pairs, newPairs...)
	pipeline, err := generateSkaffoldConfig(k, pairs)
	if err != nil {
		return err
	}

	if c.Opts.ConfigurationFile == "-" {
		out.Write(pipeline)
		return nil
	}

	if !c.Force {
		fmt.Fprintln(out, string(pipeline))

		reader := bufio.NewReader(os.Stdin)
	confirmLoop:
		for {
			fmt.Fprintf(out, "Do you want to write this configuration to %s? [y/n]: ", c.Opts.ConfigurationFile)

			response, err := reader.ReadString('\n')
			if err != nil {
				return errors.Wrap(err, "reading user confirmation")
			}

			response = strings.ToLower(strings.TrimSpace(response))
			switch response {
			case "y", "yes":
				break confirmLoop
			case "n", "no":
				return nil
			}
		}
	}

	if err := ioutil.WriteFile(c.Opts.ConfigurationFile, pipeline, 0644); err != nil {
		return errors.Wrap(err, "writing config to file")
	}

	fmt.Fprintf(out, "Configuration %s was written\n", c.Opts.ConfigurationFile)
	tips.PrintForInit(out, c.Opts)

	return nil
}

func detectBuildFile(path string) ([]InitBuilder, error) {
	// Check for jib
	if builders := jib.CheckForJib(path); builders != nil {
		return builders, filepath.SkipDir
	}

	// Check for Dockerfile
	if docker.ValidateDockerfile(path) {
		results := []InitBuilder{docker.Dockerfile(path)}
		return results, nil
	}
	return nil, nil
}

func processCliArtifacts(artifacts []string) ([]buildConfigPair, error) {
	var pairs []buildConfigPair
	for _, artifact := range artifacts {
		parts := strings.Split(artifact, "=")
		if len(parts) != 2 {
			return nil, fmt.Errorf("malformed artifact provided: %s", artifact)
		}

		// TODO: Allow passing Jib config via CLI
		pairs = append(pairs, buildConfigPair{
			BuildConfig: docker.Dockerfile(parts[0]),
			ImageName:   parts[1],
		})
	}
	return pairs, nil
}

// For each image parsed from all k8s manifests, prompt the user for
// the dockerfile that builds the referenced image
func resolveDockerfileImages(buildConfigs []InitBuilder, images []string) []buildConfigPair {
	// if we only have 1 image and 1 dockerfile, don't bother prompting
	if len(images) == 1 && len(buildConfigs) == 1 {
		return []buildConfigPair{{
			BuildConfig: buildConfigs[0],
			ImageName:   images[0],
		}}
	}

	choices := []string{}
	choiceMap := map[string]InitBuilder{}
	for _, b := range buildConfigs {
		choice := b.GetPrompt()
		choices = append(choices, choice)
		choiceMap[choice] = b
	}

	pairs := []buildConfigPair{}
	for {
		if len(images) == 0 {
			break
		}
		image := images[0]
		choice := promptUserForBuildConfig(image, choices)
		if choice != NoBuilder {
			pairs = append(pairs, buildConfigPair{BuildConfig: choiceMap[choice], ImageName: image})
			choices = util.RemoveFromSlice(choices, choice)
		}
		images = util.RemoveFromSlice(images, image)
	}
	if len(buildConfigs) > 0 {
		logrus.Warnf("unused dockerfiles found in repository: %v", buildConfigs)
	}
	return pairs
}

func promptUserForBuildConfig(image string, choices []string) string {
	var selectedBuildConfig string
	options := append(choices, NoBuilder)
	prompt := &survey.Select{
		Message:  fmt.Sprintf("Choose the builder to build image %s", image),
		Options:  options,
		PageSize: 15,
	}
	survey.AskOne(prompt, &selectedBuildConfig, nil)
	return selectedBuildConfig
}

func processBuildArtifacts(pairs []buildConfigPair) latest.BuildConfig {
	var config latest.BuildConfig
	if len(pairs) > 0 {
		var artifacts []*latest.Artifact
		for _, pair := range pairs {
			artifacts = append(artifacts, pair.BuildConfig.GetArtifact(pair.ImageName))
		}
		config.Artifacts = artifacts
	}
	return config
}

func generateSkaffoldConfig(k Initializer, buildConfigPairs []buildConfigPair) ([]byte, error) {
	// if we're here, the user has no skaffold yaml so we need to generate one
	// if the user doesn't have any k8s yamls, generate one for each dockerfile
	logrus.Info("generating skaffold config")

	cfg := &latest.SkaffoldConfig{
		APIVersion: latest.Version,
		Kind:       "Config",
	}
	if err := defaults.Set(cfg); err != nil {
		return nil, errors.Wrap(err, "generating default pipeline")
	}

	cfg.Build = processBuildArtifacts(buildConfigPairs)
	cfg.Deploy = k.GenerateDeployConfig()

	pipelineStr, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "marshaling generated pipeline")
	}

	return pipelineStr, nil
}

func printAnalyzeJSON(out io.Writer, skipBuild bool, buildConfigs []InitBuilder, images []string) error {
	if !skipBuild && len(buildConfigs) == 0 {
		return errors.New("one or more valid build configuration must be present to build images with skaffold; please provide at least one Dockerfile or Jib configuration and try again, or run `skaffold init --skip-build`")
	}
	a := struct {
		BuilderConfigs []InitBuilder `json:"builderconfigs,omitempty"`
		Images         []string      `json:"images,omitempty"`
	}{
		BuilderConfigs: buildConfigs,
		Images:         images,
	}
	contents, err := json.Marshal(a)
	if err != nil {
		return errors.Wrap(err, "marshalling contents")
	}
	_, err = out.Write(contents)
	return err
}

type buildConfigPair struct {
	BuildConfig InitBuilder
	ImageName   string
}

func walk(dir string, force bool, validateBuildFile func(string) ([]InitBuilder, error)) ([]string, []InitBuilder, error) {
	var potentialConfigs []string
	var buildFiles []InitBuilder
	err := filepath.Walk(dir, func(path string, f os.FileInfo, e error) error {
		if f.IsDir() && util.IsHiddenDir(f.Name()) {
			logrus.Debugf("skip walking hidden dir %s", f.Name())
			return filepath.SkipDir
		}
		if f.IsDir() || util.IsHiddenFile(f.Name()) {
			return nil
		}
		if IsSkaffoldConfig(path) {
			if !force {
				return fmt.Errorf("pre-existing %s found", path)
			}
			logrus.Debugf("%s is a valid skaffold configuration: continuing since --force=true", path)
			return nil
		}
		if IsSupportedKubernetesFileExtension(path) {
			potentialConfigs = append(potentialConfigs, path)
			return nil
		}
		// try and parse build file
		if builderConfigs, err := validateBuildFile(path); builderConfigs != nil {
			for _, b := range builderConfigs {
				logrus.Infof("existing builder found: %s", b.GetPrompt())
				buildFiles = append(buildFiles, b)
			}
			return err
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return potentialConfigs, buildFiles, nil
}
