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

package config

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/imdario/mergo"
	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/cluster"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/constants"
	kubectx "github.com/GoogleContainerTools/skaffold/pkg/skaffold/kubernetes/context"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/util"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/yaml"
)

const (
	defaultConfigDir  = ".skaffold"
	defaultConfigFile = "config"
)

var (
	// config-file content
	configFile     *GlobalConfig
	configFileErr  error
	configFileOnce sync.Once

	// config for a kubeContext
	config     *ContextConfig
	configErr  error
	configOnce sync.Once

	ReadConfigFile             = readConfigFileCached
	GetConfigForCurrentKubectx = getConfigForCurrentKubectx
	current                    = time.Now

	// update global config with the time the survey was last taken
	updateLastTaken = "skaffold config set --survey --global last-taken %s"
	// update global config with the time the survey was last prompted
	updateLastPrompted = "skaffold config set --survey --global last-prompted %s"
)

// readConfigFileCached reads the specified file and returns the contents
// parsed into a GlobalConfig struct.
// This function will always return the identical data from the first read.
func readConfigFileCached(filename string) (*GlobalConfig, error) {
	configFileOnce.Do(func() {
		filenameOrDefault, err := ResolveConfigFile(filename)
		if err != nil {
			configFileErr = err
			return
		}
		configFile, configFileErr = ReadConfigFileNoCache(filenameOrDefault)
		if configFileErr == nil {
			logrus.Infof("Loaded Skaffold defaults from %q", filenameOrDefault)
		}
	})
	return configFile, configFileErr
}

// ResolveConfigFile determines the default config location, if the configFile argument is empty.
func ResolveConfigFile(configFile string) (string, error) {
	if configFile == "" {
		home, err := homedir.Dir()
		if err != nil {
			return "", fmt.Errorf("retrieving home directory: %w", err)
		}
		configFile = filepath.Join(home, defaultConfigDir, defaultConfigFile)
	}
	return configFile, util.VerifyOrCreateFile(configFile)
}

// ReadConfigFileNoCache reads the given config yaml file and unmarshals the contents.
// Only visible for testing, use ReadConfigFile instead.
func ReadConfigFileNoCache(configFile string) (*GlobalConfig, error) {
	contents, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("reading global config: %w", err)
	}
	config := GlobalConfig{}
	if err := yaml.Unmarshal(contents, &config); err != nil {
		return nil, fmt.Errorf("unmarshalling global skaffold config: %w", err)
	}
	return &config, nil
}

// GetConfigForCurrentKubectx returns the specific config to be modified based on the kubeContext.
// Either returns the config corresponding to the provided or current context,
// or the global config.
func getConfigForCurrentKubectx(configFile string) (*ContextConfig, error) {
	configOnce.Do(func() {
		cfg, err := ReadConfigFile(configFile)
		if err != nil {
			configErr = err
			return
		}
		kubeconfig, err := kubectx.CurrentConfig()
		if err != nil {
			configErr = err
			return
		}
		config, configErr = getConfigForKubeContextWithGlobalDefaults(cfg, kubeconfig.CurrentContext)
	})

	return config, configErr
}

func getConfigForKubeContextWithGlobalDefaults(cfg *GlobalConfig, kubeContext string) (*ContextConfig, error) {
	if kubeContext == "" {
		if cfg.Global == nil {
			return &ContextConfig{}, nil
		}
		return cfg.Global, nil
	}

	var mergedConfig ContextConfig
	for _, contextCfg := range cfg.ContextConfigs {
		if util.RegexEqual(contextCfg.Kubecontext, kubeContext) {
			logrus.Debugf("found config for context %q", kubeContext)
			mergedConfig = *contextCfg
		}
	}
	// in case there was no config for this kubeContext in cfg.ContextConfigs
	mergedConfig.Kubecontext = kubeContext

	if cfg.Global != nil {
		// if values are unset for the current context, retrieve
		// the global config and use its values as a fallback.
		if err := mergo.Merge(&mergedConfig, cfg.Global, mergo.WithAppendSlice); err != nil {
			return nil, fmt.Errorf("merging context-specific and global config: %w", err)
		}
	}
	return &mergedConfig, nil
}

func GetDefaultRepo(configFile string, cliValue *string) (string, error) {
	// CLI flag takes precedence.
	if cliValue != nil {
		return *cliValue, nil
	}
	cfg, err := GetConfigForCurrentKubectx(configFile)
	if err != nil {
		return "", err
	}
	if cfg.DefaultRepo != "" {
		logrus.Infof("Using default-repo=%s from config", cfg.DefaultRepo)
	}
	return cfg.DefaultRepo, nil
}

func GetInsecureRegistries(configFile string) ([]string, error) {
	cfg, err := GetConfigForCurrentKubectx(configFile)
	if err != nil {
		return nil, err
	}
	if len(cfg.InsecureRegistries) > 0 {
		logrus.Infof("Using insecure-registries=%v from config", cfg.InsecureRegistries)
	}
	return cfg.InsecureRegistries, nil
}

func GetDebugHelpersRegistry(configFile string) (string, error) {
	cfg, err := GetConfigForCurrentKubectx(configFile)
	if err != nil {
		return "", err
	}

	if cfg.DebugHelpersRegistry != "" {
		logrus.Infof("Using debug-helpers-registry=%s from config", cfg.DebugHelpersRegistry)
		return cfg.DebugHelpersRegistry, nil
	}
	return constants.DefaultDebugHelpersRegistry, nil
}

func GetCluster(configFile string, minikubeProfile string, detectMinikube bool) (Cluster, error) {
	cfg, err := GetConfigForCurrentKubectx(configFile)
	if err != nil {
		return Cluster{}, err
	}

	kubeContext := cfg.Kubecontext
	isKindCluster, isK3dCluster, isMicrok8sCluster := IsKindCluster(kubeContext), IsK3dCluster(kubeContext), IsMicrok8sCluster(kubeContext)

	var local bool
	switch {
	case minikubeProfile != "":
		local = true

	case cfg.LocalCluster != nil:
		logrus.Infof("Using local-cluster=%t from config", *cfg.LocalCluster)
		local = *cfg.LocalCluster

	case kubeContext == constants.DefaultMinikubeContext ||
		kubeContext == constants.DefaultDockerForDesktopContext ||
		kubeContext == constants.DefaultDockerDesktopContext ||
		isKindCluster || isK3dCluster || isMicrok8sCluster:
		local = true

	case detectMinikube:
		local = cluster.GetClient().IsMinikube(kubeContext)

	default:
		local = false
	}

	kindDisableLoad := cfg.KindDisableLoad != nil && *cfg.KindDisableLoad
	k3dDisableLoad := cfg.K3dDisableLoad != nil && *cfg.K3dDisableLoad
	microk8sDisableLoad := cfg.Microk8sDisableLoad != nil && *cfg.Microk8sDisableLoad

	// load images for local kind/k3d/microk8s cluster unless explicitly disabled
	loadImages := local && ((isKindCluster && !kindDisableLoad) || (isK3dCluster && !k3dDisableLoad) || (isMicrok8sCluster && !microk8sDisableLoad))

	// push images for remote cluster or local kind/k3d/microk8s cluster with image loading disabled
	pushImages := !local || (isKindCluster && kindDisableLoad) || (isK3dCluster && k3dDisableLoad) || (isMicrok8sCluster && microk8sDisableLoad)

	return Cluster{
		Local:      local,
		LoadImages: loadImages,
		PushImages: pushImages,
	}, nil
}

// IsKindCluster checks that the given `kubeContext` is talking to `kind`.
func IsKindCluster(kubeContext string) bool {
	switch {
	// With kind version < 0.6.0, the k8s context
	// is `[CLUSTER NAME]@kind`.
	// For eg: `cluster@kind`
	// the default name is `kind@kind`
	case strings.HasSuffix(kubeContext, "@kind"):
		return true

	// With kind version >= 0.6.0, the k8s context
	// is `kind-[CLUSTER NAME]`.
	// For eg: `kind-cluster`
	// the default name is `kind-kind`
	case strings.HasPrefix(kubeContext, "kind-"):
		return true

	default:
		return false
	}
}

// KindClusterName returns the internal kind name of a kubernetes cluster.
func KindClusterName(clusterName string) string {
	switch {
	// With kind version < 0.6.0, the k8s context
	// is `[CLUSTER NAME]@kind`.
	// For eg: `cluster@kind`
	// the default name is `kind@kind`
	case strings.HasSuffix(clusterName, "@kind"):
		return strings.TrimSuffix(clusterName, "@kind")

	// With kind version >= 0.6.0, the k8s context
	// is `kind-[CLUSTER NAME]`.
	// For eg: `kind-cluster`
	// the default name is `kind-kind`
	case strings.HasPrefix(clusterName, "kind-"):
		return strings.TrimPrefix(clusterName, "kind-")
	}

	return clusterName
}

// IsK3dCluster checks that the given `kubeContext` is talking to `k3d`.
func IsK3dCluster(kubeContext string) bool {
	return strings.HasPrefix(kubeContext, "k3d-")
}

// K3dClusterName returns the internal name of a k3d cluster.
func K3dClusterName(clusterName string) string {
	if strings.HasPrefix(clusterName, "k3d-") {
		return strings.TrimPrefix(clusterName, "k3d-")
	}
	return clusterName
}

// IsMicrok8sCluster checks that the given `kubeContext` is talking to `microk8s`.
func IsMicrok8sCluster(kubeContext string) bool {
	return strings.HasPrefix(kubeContext, "microk8s-")
}

// Microk8sClusterName returns the internal name of a microk8s cluster.
func Microk8sClusterName(clusterName string) string {
	if strings.HasPrefix(clusterName, "microk8s-") {
		return strings.TrimPrefix(clusterName, "microk8s-")
	}
	return clusterName
}

func IsUpdateCheckEnabled(configfile string) bool {
	cfg, err := GetConfigForCurrentKubectx(configfile)
	if err != nil {
		return true
	}
	return cfg == nil || cfg.UpdateCheck == nil || *cfg.UpdateCheck
}

func ShouldDisplayPrompt(configfile string) bool {
	cfg, disabled := isSurveyPromptDisabled(configfile)
	return !disabled && !recentlyPromptedOrTaken(cfg)
}

func isSurveyPromptDisabled(configfile string) (*ContextConfig, bool) {
	cfg, err := GetConfigForCurrentKubectx(configfile)
	if err != nil {
		return nil, false
	}
	return cfg, cfg != nil && cfg.Survey != nil && cfg.Survey.DisablePrompt != nil && *cfg.Survey.DisablePrompt
}

func recentlyPromptedOrTaken(cfg *ContextConfig) bool {
	if cfg == nil || cfg.Survey == nil {
		return false
	}
	return lessThan(cfg.Survey.LastTaken, 365*24*time.Hour) || lessThan(cfg.Survey.LastPrompted, 60*24*time.Hour)
}

func lessThan(date string, duration time.Duration) bool {
	t, err := time.Parse(time.RFC3339, date)
	if err != nil {
		logrus.Debugf("could not parse date %q", date)
		return false
	}
	return current().Sub(t) < duration
}

func UpdateGlobalSurveyTaken(configFile string) error {
	// Today's date
	today := current().Format(time.RFC3339)
	ai := fmt.Sprintf(updateLastTaken, today)
	aiErr := fmt.Errorf("could not automatically update the survey timestamp - please run `%s`", ai)

	configFile, err := ResolveConfigFile(configFile)
	if err != nil {
		return aiErr
	}
	fullConfig, err := ReadConfigFile(configFile)
	if err != nil {
		return aiErr
	}
	if fullConfig.Global == nil {
		fullConfig.Global = &ContextConfig{}
	}
	if fullConfig.Global.Survey == nil {
		fullConfig.Global.Survey = &SurveyConfig{}
	}
	fullConfig.Global.Survey.LastTaken = today
	err = WriteFullConfig(configFile, fullConfig)
	if err != nil {
		return aiErr
	}
	return err
}

func UpdateGlobalSurveyPrompted(configFile string) error {
	// Today's date
	today := current().Format(time.RFC3339)
	ai := fmt.Sprintf(updateLastPrompted, today)
	aiErr := fmt.Errorf("could not automatically update the survey prompted timestamp - please run `%s`", ai)

	configFile, err := ResolveConfigFile(configFile)
	if err != nil {
		return aiErr
	}
	fullConfig, err := ReadConfigFile(configFile)
	if err != nil {
		return aiErr
	}
	if fullConfig.Global == nil {
		fullConfig.Global = &ContextConfig{}
	}
	if fullConfig.Global.Survey == nil {
		fullConfig.Global.Survey = &SurveyConfig{}
	}
	fullConfig.Global.Survey.LastPrompted = today
	err = WriteFullConfig(configFile, fullConfig)
	if err != nil {
		return aiErr
	}
	return err
}

func UpdateGlobalCollectMetrics(configFile string, collectMetrics bool) error {
	configFile, err := ResolveConfigFile(configFile)
	if err != nil {
		return err
	}
	fullConfig, err := ReadConfigFile(configFile)
	if err != nil {
		return err
	}
	if fullConfig.Global == nil {
		fullConfig.Global = &ContextConfig{}
	}
	fullConfig.Global.CollectMetrics = &collectMetrics
	err = WriteFullConfig(configFile, fullConfig)
	if err != nil {
		return err
	}
	return err
}

func WriteFullConfig(configFile string, cfg *GlobalConfig) error {
	contents, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	configFileOrDefault, err := ResolveConfigFile(configFile)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(configFileOrDefault, contents, 0644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}
	return nil
}
