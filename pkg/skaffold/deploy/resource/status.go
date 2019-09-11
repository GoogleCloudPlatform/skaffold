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

package resource

type Status struct {
	err     error
	details string
}

func (rs Status) Error() error {
	return rs.err
}

func (rs Status) String() string {
	if rs.err != nil {
		return rs.err.Error()
	}
	return rs.details
}

func newStatus(msg string, err error) Status {
	return Status{
		details: msg,
		err:     err,
	}
}
