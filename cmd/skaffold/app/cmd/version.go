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

package cmd

import (
	"io"

	"github.com/GoogleContainerTools/skaffold/cmd/skaffold/app/cmd/helper"
	"github.com/GoogleContainerTools/skaffold/cmd/skaffold/app/flags"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/version"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var versionFlag = flags.NewTemplateFlag("{{.Version}}\n", version.Info{})

func NewCmdVersion(out io.Writer) *cobra.Command {
	cmd := helper.NoArgCommand(out, "version", "Print the version information", RunVersion)
	cmd.Flags().VarP(versionFlag, "output", "o", versionFlag.Usage())
	return cmd
}

func RunVersion(out io.Writer) error {
	if err := versionFlag.Template().Execute(out, version.Get()); err != nil {
		return errors.Wrap(err, "executing template")
	}
	return nil
}
