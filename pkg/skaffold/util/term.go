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

package util

import (
	"io"
	"os/exec"
	"strconv"
	"strings"

	"golang.org/x/term"
)

var colors = tputColors

func IsTerminal(w io.Writer) (uintptr, bool) {
	type descriptor interface {
		Fd() uintptr
	}

	if f, ok := w.(descriptor); ok {
		termFd := f.Fd()
		isTerm := term.IsTerminal(int(termFd))
		return termFd, isTerm
	}

	return 0, false
}

func SupportsColor() (bool, error) {
	res, err := colors()
	if err != nil {
		return false, err
	}

	numColors, err := strconv.Atoi(strings.TrimSpace(string(res)))
	if err != nil {
		return false, err
	}

	return numColors != 0, nil
}

func tputColors() ([]byte, error) {
	return exec.Command("tput", "colors").Output()
}
