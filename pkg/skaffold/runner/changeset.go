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

package runner

import (
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/sync"
)

type changeSet struct {
	needsRebuild  []*latest.Artifact
	needsResync   []*sync.Item
	needsRedeploy bool
	needsReload   bool
}

func (c *changeSet) AddRebuild(a *latest.Artifact) {
	c.needsRebuild = append(c.needsRebuild, a)
}

func (c *changeSet) AddResync(s *sync.Item) {
	c.needsResync = append(c.needsResync, s)
}

func (c *changeSet) reset() {
	c.needsRebuild = nil
	c.needsResync = nil

	c.needsRedeploy = false
	c.needsReload = false
}

func (c *changeSet) needsAction() bool {
	return c.needsReload || len(c.needsRebuild) > 0 || c.needsRedeploy
}
