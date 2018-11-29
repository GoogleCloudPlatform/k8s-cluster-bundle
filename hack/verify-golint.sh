#!/usr/bin/env bash
# Copyright 2018 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset

REPO_ROOT=$(git rev-parse --show-toplevel)
cd "${REPO_ROOT}"

command -v golint >/dev/null 2>&1 || {
  errmsg="Error: golint is required. run 'go install github.com/golang/golint'"
  printf >&2 "${errmsg}"; exit 1;
}

golint -set_exit_status ./pkg/...
