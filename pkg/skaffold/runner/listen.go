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
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/color"
	"github.com/sirupsen/logrus"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/filemon"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/trigger"
)

type Listener interface {
	WatchForChanges(context.Context, io.Writer, func() (bool, func() error)) error
	LogWatchIsActive(io.Writer)
	LogWatchIsInactive(io.Writer)
}

type SkaffoldListener struct {
	Monitor    filemon.Monitor
	Trigger    trigger.Trigger
	intentChan <-chan bool
}

func (l *SkaffoldListener) LogWatchIsActive(out io.Writer) {
	l.Trigger.LogWatchToUser(out)
}

func (l *SkaffoldListener) LogWatchIsInactive(out io.Writer) {
	color.Yellow.Fprintln(out, "Not watching for changes...")
}

// WatchForChanges listens to a trigger, and when one is received, computes file changes and
// conditionally runs the dev loop. It logs to user only when the dev loop is active.
func (l *SkaffoldListener) WatchForChanges(ctx context.Context, out io.Writer, fn func() (isActive bool, devLoop func() error)) error {
	ctxTrigger, cancelTrigger := context.WithCancel(ctx)
	defer cancelTrigger()
	trigger, err := trigger.StartTrigger(ctxTrigger, l.Trigger)
	if err != nil {
		return fmt.Errorf("unable to start trigger: %w", err)
	}

	// exit if file monitor fails the first time
	if err := l.Monitor.Run(l.Trigger.Debounce()); err != nil {
		return fmt.Errorf("failed to monitor files: %w", err)
	}

	isActive, devLoop := fn()
	if isActive {
		l.LogWatchIsActive(out)
	} else {
		l.LogWatchIsInactive(out)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-l.intentChan:
			if err := l.do(devLoop); err != nil {
				return err
			}
		case <-trigger:
			if err := l.do(devLoop); err != nil {
				return err
			}
		}
	}
}

func (l *SkaffoldListener) do(devLoop func() error) error {
	if err := l.Monitor.Run(l.Trigger.Debounce()); err != nil {
		logrus.Warnf("Ignoring changes: %s", err.Error())
		return nil
	}

	if err := devLoop(); err != nil {
		// propagating this error up causes a new runner to be created
		// and a new dev loop to start
		if errors.Is(err, ErrorConfigurationChanged) {
			return err
		}
		logrus.Errorf("error running dev loop: %s", err.Error())
	}

	return nil
}
