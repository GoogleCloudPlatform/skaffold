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

import "sync"

type intents struct {
	build      bool
	sync       bool
	test       bool
	deploy     bool
	autoBuild  bool
	autoSync   bool
	autoTest   bool
	autoDeploy bool

	lock sync.Mutex
}

func newIntents(autoBuild, autoSync, autoTest, autoDeploy bool) *intents {
	i := &intents{
		autoBuild:  autoBuild,
		autoSync:   autoSync,
		autoTest:   autoTest,
		autoDeploy: autoDeploy,
	}

	return i
}

func (i *intents) resetBuild() {
	i.lock.Lock()
	i.build = i.autoBuild
	i.lock.Unlock()
}

func (i *intents) resetSync() {
	i.lock.Lock()
	i.sync = i.autoSync
	i.lock.Unlock()
}

func (i *intents) resetTest() {
	i.lock.Lock()
	i.test = i.autoTest
	i.lock.Unlock()
}

func (i *intents) resetDeploy() {
	i.lock.Lock()
	i.deploy = i.autoDeploy
	i.lock.Unlock()
}

func (i *intents) setBuild(val bool) {
	i.lock.Lock()
	i.build = val
	i.lock.Unlock()
}

func (i *intents) setSync(val bool) {
	i.lock.Lock()
	i.sync = val
	i.lock.Unlock()
}

func (i *intents) setTest(val bool) {
	i.lock.Lock()
	i.test = val
	i.lock.Unlock()
}

func (i *intents) setDeploy(val bool) {
	i.lock.Lock()
	i.deploy = val
	i.lock.Unlock()
}

func (i *intents) getAutoBuild() bool {
	i.lock.Lock()
	defer i.lock.Unlock()
	return i.autoBuild
}

func (i *intents) getAutoSync() bool {
	i.lock.Lock()
	defer i.lock.Unlock()
	return i.autoSync
}

func (i *intents) getAutoTest() bool {
	i.lock.Lock()
	defer i.lock.Unlock()
	return i.autoTest
}

func (i *intents) getAutoDeploy() bool {
	i.lock.Lock()
	defer i.lock.Unlock()
	return i.autoDeploy
}

func (i *intents) setAutoBuild(val bool) {
	i.lock.Lock()
	i.autoBuild = val
	i.lock.Unlock()
}

func (i *intents) setAutoSync(val bool) {
	i.lock.Lock()
	i.autoSync = val
	i.lock.Unlock()
}

func (i *intents) setAutoTest(val bool) {
	i.lock.Lock()
	i.autoTest = val
	i.lock.Unlock()
}

func (i *intents) setAutoDeploy(val bool) {
	i.lock.Lock()
	i.autoDeploy = val
	i.lock.Unlock()
}

// returns build, sync, test, and deploy intents (in that order)
func (i *intents) GetIntents() (bool, bool, bool, bool) {
	i.lock.Lock()
	defer i.lock.Unlock()
	return i.build, i.sync, i.test, i.deploy
}

func (i *intents) IsAnyAutoEnabled() bool {
	i.lock.Lock()
	defer i.lock.Unlock()
	return i.autoBuild || i.autoSync || i.autoTest || i.autoDeploy
}
