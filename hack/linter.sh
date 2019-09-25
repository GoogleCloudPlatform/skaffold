#!/bin/bash

# Copyright 2019 The Skaffold Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e -o pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

if ! [ -x "$(command -v golangci-lint)" ]; then
	echo "Installing GolangCI-Lint"
	${DIR}/install_golint.sh -b $GOPATH/bin v1.19.1
fi

VERBOSE=""
if [[ "${TRAVIS}" == "true" ]]; then
    # Use less memory on Travis
    # See https://github.com/golangci/golangci-lint#memory-usage-of-golangci-lint
    export GOGC=${GOLINT_GOGC:-8}
    VERBOSE="-v --print-resources-usage"
fi

# Limit number of default jobs, to avoid the CI builds running out of memory
GOLINT_JOBS=${GOLINT_JOBS:-4}

golangci-lint run ${VERBOSE} -c ${DIR}/golangci.yml --concurrency $GOLINT_JOBS \
    | awk '/out of memory/ || /Deadline exceeded/ {failed = 1}; {print}; END {exit failed}'
