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

package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/kubernetes/context"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

const defaultConfigLocation = ".skaffold/config"

func resolveKubectlContext() {
	if kubectx != "" {
		return
	}

	context, err := context.CurrentContext()
	if err != nil {
		logrus.Warn(errors.Wrap(err, "retrieving current kubectl context"))
		kubectx = "default"
	}
	if context == "" {
		logrus.Infof("no context currently set, falling back to default")
		kubectx = "default"
	}
	kubectx = context
}

func resolveConfigFile() error {
	var err error
	if configFile != "" {
		// we had a config provided as a flag, expand it and return
		if !filepath.IsAbs(configFile) {
			absPath, err := filepath.Abs(configFile)
			if err != nil {
				return err
			}
			configFile = absPath
		}
	} else {
		home, err := homedir.Dir()
		if err != nil {
			return errors.Wrap(err, "retrieving home directory")
		}
		configFile = filepath.Join(home, defaultConfigLocation)
	}
	_, err = os.Stat(configFile)
	// TODO(nkubala): create default config?
	if err != nil {
		return err
	}

	return nil
}

func readConfig() (*Config, error) {
	if err := resolveConfigFile(); err != nil {
		return nil, errors.Wrap(err, "resolving config file location")
	}
	contents, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, errors.Wrap(err, "reading global config")
	}
	config := Config{}
	if err := yaml.Unmarshal(contents, &config); err != nil {
		return nil, errors.Wrap(err, "unmarshalling global skaffold config")
	}
	return &config, nil
}

func getConfigsForKubectx() (*Config, error) {
	configs, err := readConfig()
	if err != nil {
		return nil, err
	}
	if kubectx == "all" {
		return configs, nil
	}
	for _, cfg := range *configs {
		if cfg.Context == kubectx {
			return &[]*ContextConfig{cfg}, nil
		}
	}
	return nil, fmt.Errorf("no config entry found for kubectx %s", kubectx)
}
