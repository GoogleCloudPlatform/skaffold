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
	"github.com/GoogleContainerTools/skaffold/proto"
)

type changeSet struct {
	needsRebuild   []*latest.Artifact
	rebuildTracker map[string]*latest.Artifact
	needsResync    []*sync.Item
	resyncTracker  map[string]*sync.Item
	changeType     proto.ChangeType
	needsRedeploy  bool
	needsReload    bool
}

func (c *changeSet) AddRebuild(a *latest.Artifact) {
	if _, ok := c.rebuildTracker[a.ImageName]; ok {
		return
	}
	c.rebuildTracker[a.ImageName] = a
	c.needsRebuild = append(c.needsRebuild, a)
	c.needsRedeploy = true
	c.changeType = proto.ChangeType_BUILD
}

func (c *changeSet) AddResync(s *sync.Item) {
	if _, ok := c.resyncTracker[s.Image]; ok {
		return
	}
	c.resyncTracker[s.Image] = s
	c.needsResync = append(c.needsResync, s)
	c.changeType = proto.ChangeType_SYNC
}

func (c *changeSet) AddRedeploy(ct proto.ChangeType) {
	c.needsRedeploy = true
	c.changeType = ct
}

func (c *changeSet) NeedsReload() {
	c.needsReload = true
	c.changeType = proto.ChangeType_CONFIG
}

func (c *changeSet) resetBuild() {
	c.rebuildTracker = make(map[string]*latest.Artifact)
	c.needsRebuild = nil
}

func (c *changeSet) resetSync() {
	c.resyncTracker = make(map[string]*sync.Item)
	c.needsResync = nil
}

func (c *changeSet) resetDeploy() {
	c.needsRedeploy = false
}
