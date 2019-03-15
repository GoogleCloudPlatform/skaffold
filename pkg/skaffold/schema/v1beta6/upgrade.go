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

package v1beta6

import (
	next "github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/util"
	pkgutil "github.com/GoogleContainerTools/skaffold/pkg/skaffold/util"
	"github.com/pkg/errors"
)

// Upgrade upgrades a configuration to the next version.
// Config changes from v1beta6 to v1beta7
// 1. Additions:
// 2. No removals
// 3. No updates
func (config *SkaffoldPipeline) Upgrade() (util.VersionedConfig, error) {
	// convert Deploy (should be the same)
	var newDeploy next.DeployConfig
	if err := pkgutil.CloneThroughJSON(config.Deploy, &newDeploy); err != nil {
		return nil, errors.Wrap(err, "converting deploy config")
	}

	// convert Profiles (should be the same)
	var newProfiles []next.Profile
	if config.Profiles != nil {
		if err := pkgutil.CloneThroughJSON(config.Profiles, &newProfiles); err != nil {
			return nil, errors.Wrap(err, "converting new profile")
		}
	}

	// Update profile if kaniko build exists
	for i, p := range config.Profiles {
		if err := upgradeKanikoBuild(p.Build, &newProfiles[i].Build); err != nil {
			return nil, errors.Wrap(err, "upgrading kaniko build")
		}
	}

	// convert Sync and clear the input because it thrashes the cloning
	shouldSync := make([]bool, 0, len(config.Build.Artifacts))
	for _, a := range config.Build.Artifacts {
		shouldSync = append(shouldSync, len(a.Sync) > 0)
		a.Sync = nil
	}
	var newBuild next.BuildConfig
	if err := pkgutil.CloneThroughJSON(config.Build, &newBuild); err != nil {
		return nil, errors.Wrap(err, "converting new build")
	}
	// convert Kaniko if needed
	if err := upgradeKanikoBuild(config.Build, &newBuild); err != nil {
		return nil, errors.Wrap(err, "upgrading kaniko build")
	}
	// set Sync in newBuild
	for i, a := range newBuild.Artifacts {
		a.Sync = shouldSync[i]
	}

	// convert Test (should be the same)
	var newTest []*next.TestCase
	if err := pkgutil.CloneThroughJSON(config.Test, &newTest); err != nil {
		return nil, errors.Wrap(err, "converting new test")
	}

	return &next.SkaffoldPipeline{
		APIVersion: next.Version,
		Kind:       config.Kind,
		Build:      newBuild,
		Test:       newTest,
		Deploy:     newDeploy,
		Profiles:   newProfiles,
	}, nil
}

func upgradeKanikoBuild(build BuildConfig, newConfig *next.BuildConfig) error {
	if build.KanikoBuild == nil {
		return nil
	}
	kaniko := build.KanikoBuild
	// Else, transition values from old config to new config artifacts
	for _, a := range newConfig.Artifacts {
		if err := pkgutil.CloneThroughJSON(kaniko, &a.KanikoArtifact); err != nil {
			return errors.Wrap(err, "cloning kaniko artifact")
		}
	}
	// Transition values from old config to in cluster details
	if err := pkgutil.CloneThroughJSON(kaniko, &newConfig.Cluster); err != nil {
		return errors.Wrap(err, "cloning cluster details")
	}
	return nil
}
