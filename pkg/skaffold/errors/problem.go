/*
Copyright 2020 The Skaffold Authors

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

package errors

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/GoogleContainerTools/skaffold/proto/v1"
)

type descriptionFunc func(error) string
type suggestionFunc func(cfg Config) []*proto.Suggestion

// Problem defines a problem which can list suggestions and error codes
// evaluated when showing Actionable error messages
type Problem struct {
	Regexp      *regexp.Regexp
	Description func(error) string
	ErrCode     proto.StatusCode
	Suggestion  func(cfg Config) []*proto.Suggestion
	Err         error
}

func NewProblem(d descriptionFunc, sc proto.StatusCode, s suggestionFunc, err error) Problem {
	return Problem{
		Description: d,
		ErrCode:     sc,
		Suggestion:  s,
		Err:         err,
	}
}

func (p Problem) Error() string {
	description := fmt.Sprintf("%s\n", p.Err)
	if p.Description != nil {
		description = p.Description(p.Err)
	}
	return description
}

func (p Problem) AIError(cfg Config) string{
	if suggestions := p.Suggestion(cfg); suggestions != nil {
		return fmt.Sprintf("%s. %s", strings.Trim(p.Error(), "."), concatSuggestions(suggestions))
	}
	return p.Error()
}

func (p Problem) withConfigAndErr(err error) Problem {
	p.Err = err
	return p
}

func isProblem(err error) (Problem, bool) {
	if p, ok := err.(Problem); ok {
		return p, true
	}
	return Problem{}, false
}
