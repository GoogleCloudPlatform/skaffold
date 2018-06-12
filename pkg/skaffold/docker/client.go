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

package docker

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/constants"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/kubernetes"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/util"
	"github.com/docker/docker/api"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/tlsconfig"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type APIClient interface {
	client.CommonAPIClient
}

var (
	dockerAPIClientOnce sync.Once
	dockerAPIClient     APIClient
	dockerAPIClientErr  error
)

// NewAPIClient guesses the docker client to use based on current kubernetes context.
func NewAPIClient() (APIClient, error) {
	dockerAPIClientOnce.Do(func() {
		kubeContext, err := kubernetes.CurrentContext()
		if err != nil {
			dockerAPIClientErr = errors.Wrap(err, "getting current cluster context")
			return
		}

		dockerAPIClient, dockerAPIClientErr = newAPIClient(kubeContext)
	})

	return dockerAPIClient, dockerAPIClientErr
}

// newAPIClient guesses the docker client to use based on current kubernetes context.
func newAPIClient(kubeContext string) (APIClient, error) {
	if kubeContext == constants.DefaultMinikubeContext {
		return newMinikubeAPIClient()
	}
	return newEnvAPIClient()
}

// newEnvAPIClient returns a docker client based on the environment variables set.
// It will "negotiate" the highest possible API version supported by both the client
// and the server if there is a mismatch.
func newEnvAPIClient() (APIClient, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, fmt.Errorf("Error getting docker client: %s", err)
	}
	cli.NegotiateAPIVersion(context.Background())

	return cli, nil
}

// newMinikubeAPIClient returns a docker client using the environment variables
// provided by minikube.
func newMinikubeAPIClient() (APIClient, error) {
	env, err := getMinikubeDockerEnv()
	if err != nil {
		logrus.Warnf("Could not get minikube docker env, falling back to local docker daemon")
		return newEnvAPIClient()
	}

	var httpclient *http.Client
	if dockerCertPath := env["DOCKER_CERT_PATH"]; dockerCertPath != "" {
		options := tlsconfig.Options{
			CAFile:             filepath.Join(dockerCertPath, "ca.pem"),
			CertFile:           filepath.Join(dockerCertPath, "cert.pem"),
			KeyFile:            filepath.Join(dockerCertPath, "key.pem"),
			InsecureSkipVerify: env["DOCKER_TLS_VERIFY"] == "",
		}
		tlsc, err := tlsconfig.Client(options)
		if err != nil {
			return nil, err
		}

		httpclient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsc,
			},
			CheckRedirect: client.CheckRedirect,
		}
	}

	host := env["DOCKER_HOST"]
	if host == "" {
		host = client.DefaultDockerHost
	}
	version := env["DOCKER_API_VERSION"]
	if version == "" {
		version = api.DefaultVersion
	}

	return client.NewClient(host, version, httpclient, nil)
}

func getMinikubeDockerEnv() (map[string]string, error) {
	cmd := exec.Command("minikube", "docker-env", "--shell", "none")
	out, err := util.RunCmdOut(cmd)
	if err != nil {
		return nil, errors.Wrap(err, "getting minikube env")
	}

	env := map[string]string{}
	for _, line := range strings.Split(string(out), "\n") {
		if line == "" {
			continue
		}
		kv := strings.Split(line, "=")
		if len(kv) != 2 {
			return nil, fmt.Errorf("Unable to parse minikube docker-env keyvalue: %s, line: %s, output: %s", kv, line, string(out))
		}
		env[kv[0]] = kv[1]
	}
	return env, nil
}
