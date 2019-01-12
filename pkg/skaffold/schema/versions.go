/*
Copyright 2018 The Skaffold Authors

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
	"path/filepath"

	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"

	apiversion "github.com/GoogleContainerTools/skaffold/pkg/skaffold/apiversion"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/util"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/v1alpha1"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/v1alpha2"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/v1alpha3"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/v1alpha4"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/v1alpha5"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/v1beta1"
	misc "github.com/GoogleContainerTools/skaffold/pkg/skaffold/util"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/yamltags"
)

type APIVersion struct {
	Version string `yaml:"apiVersion"`
}

var schemaVersions = versions{
	{v1alpha1.Version, v1alpha1.NewSkaffoldPipeline},
	{v1alpha2.Version, v1alpha2.NewSkaffoldPipeline},
	{v1alpha3.Version, v1alpha3.NewSkaffoldPipeline},
	{v1alpha4.Version, v1alpha4.NewSkaffoldPipeline},
	{v1alpha5.Version, v1alpha5.NewSkaffoldPipeline},
	{v1beta1.Version, v1beta1.NewSkaffoldPipeline},
	{latest.Version, latest.NewSkaffoldPipeline},
}

type version struct {
	apiVersion string
	factory    func() util.VersionedConfig
}

type versions []version

// Find search the constructor for a given api version.
func (v *versions) Find(apiVersion string) (func() util.VersionedConfig, bool) {
	for _, version := range *v {
		if version.apiVersion == apiVersion {
			return version.factory, true
		}
	}

	return nil, false
}

func ParseSingleConfigFile(filename string, upgrade bool) (util.VersionedConfig, error) {
	buf, err := misc.ReadConfiguration(filename)
	if err != nil {
		return nil, errors.Wrap(err, "read skaffold config")
	}

	apiVersion := &APIVersion{}
	if err := yaml.Unmarshal(buf, apiVersion); err != nil {
		return nil, errors.Wrap(err, "parsing api version")
	}

	factory, present := schemaVersions.Find(apiVersion.Version)
	if !present {
		return nil, errors.Errorf("unknown api version: '%s'", apiVersion.Version)
	}

	cfg := factory()
	if err := yaml.UnmarshalStrict(buf, cfg); err != nil {
		return nil, errors.Wrap(err, "unable to parse config")
	}

	if err := yamltags.ProcessStruct(cfg); err != nil {
		return nil, errors.Wrap(err, "invalid config")
	}

	if upgrade && cfg.GetVersion() != latest.Version {
		cfg, err = upgradeToLatest(cfg)
		if err != nil {
			return nil, err
		}
	}
	return cfg, nil

}
func ReadAdditionalConfigurationFile(originalConfigFile *latest.SkaffoldPipeline, filename string, upgrade bool) {
	println(filename)
	if misc.FileExists(filename) {
		profileConfiguration, err := ParseSingleConfigFile(filename, upgrade)
		if err != nil {
			logrus.Warnf("unable to %s %s", filename, err)
			return
		}
		logrus.Debugf("%-v", profileConfiguration)
		if err := mergo.Merge(originalConfigFile, profileConfiguration, mergo.WithOverride); err != nil {
			logrus.Warnf("unable to merge configurations from %s %s", filename, err)
			return
		}
	}

}

// ParseConfig reads a configuration file.
func ParseConfig(filename string, upgrade bool, activeProfiles []string) (util.VersionedConfig, error) {
	cfg, err := ParseSingleConfigFile(filename, upgrade)
	if err != nil {
		return nil, err
	}
	materializedConfig := cfg.(*latest.SkaffoldPipeline)
	for _, profile := range activeProfiles {
		directory := filepath.Dir(filename)
		for _, extension := range []string{"yaml", "yml"} {
			profileSkaffoldFile := filepath.Join(directory, fmt.Sprintf("skaffold_%s.%s", profile, extension))
			logrus.Debugf("Testing if profile %s has configuration file:%s", profile, profileSkaffoldFile)
			ReadAdditionalConfigurationFile(materializedConfig, profileSkaffoldFile, upgrade)
		}
	}
	return cfg, nil
}

// upgradeToLatest upgrades a configuration to the latest version.
func upgradeToLatest(vc util.VersionedConfig) (util.VersionedConfig, error) {
	var err error

	// first, check to make sure config version isn't too new
	version, err := apiversion.Parse(vc.GetVersion())
	if err != nil {
		return nil, errors.Wrap(err, "parsing api version")
	}

	semver := apiversion.MustParse(latest.Version)
	if version.EQ(semver) {
		return vc, nil
	}
	if version.GT(semver) {
		return nil, fmt.Errorf("config version %s is too new for this version of skaffold: upgrade skaffold", vc.GetVersion())
	}

	logrus.Warnf("config version (%s) out of date: upgrading to latest (%s)", vc.GetVersion(), latest.Version)

	for vc.GetVersion() != latest.Version {
		vc, err = vc.Upgrade()
		if err != nil {
			return nil, errors.Wrapf(err, "transforming skaffold config")
		}
	}

	return vc, nil
}
