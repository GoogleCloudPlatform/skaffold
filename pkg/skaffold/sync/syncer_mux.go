/*
Copyright 2021 The Skaffold Authors

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

package sync

import (
	"context"
	"io"
)

type SyncerMux []Syncer

func (s SyncerMux) Sync(ctx context.Context, out io.Writer, item *Item) error {
	for _, syncer := range s {
		if err := syncer.Sync(ctx, out, item); err != nil {
			return err
		}
	}
	return nil
}
