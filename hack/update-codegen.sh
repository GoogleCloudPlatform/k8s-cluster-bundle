#!/usr/bin/env bash
# Copyright 2020 Google LLC
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
set -o pipefail

K8S_VERSION="release-1.18"

command -v deepcopy-gen >/dev/null 2>&1 || {
  pt1="Error: deepcopy-gen is required, but was not found. Download $K8S_VERSION of k8s.io/code-generator.\n"
  pt2="Then, install with 'go install k8s.io/code-generator/cmd/deepcopy-gen'"
  printf >&2 "${pt1}${pt2}"; exit 1;
}

command -v register-gen >/dev/null 2>&1 || {
  pt1="Error: register-gen is required, but was not found. Download $K&S_VERSION of k8s.io/code-generator.\n"
  pt2="Then, install with 'go install k8s.io/code-generator/cmd/register-gen'"
  printf >&2 "${pt1}${pt2}"; exit 1;
}

command -v client-gen >/dev/null 2>&1 || {
  pt1="Error: client-gen tool is required, but was not found. Download $K8S_VERSION of k8s.io/code-generator.\n"
  pt2="Then, install with 'go install k8s.io/code-generator/cmd/client-gen'"
  printf >&2 "${pt1}${pt2}"; exit 1;
}

command -v controller-gen >/dev/null 2>&1 || {
  pt1="Error: controller-gen (generator) is required, but was not found. Download sigs.k8s.io/controller-tools.\n"
  pt2="Then, install with 'go install sigs.k8s.io/controller-tools/cmd/controller-gen'"
  printf >&2 "${pt1}${pt2}"; exit 1;
}

# **Note**: For all commands, the working directory is the parent directory
REPO_ROOT=$(git rev-parse --show-toplevel)
cd "${REPO_ROOT}"

echo "Generating with deepcopy-gen"
deepcopy-gen \
  -h hack/boilerplate.go.txt \
  -O zz_generated.deepcopy \
  --input-dirs=github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1 \
  --output-package=github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1

echo "Generating with register-gen"
register-gen \
  -h hack/boilerplate.go.txt \
  --input-dirs=github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1 \
  --output-package=github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1

echo "Generating with client-gen"
client-gen --clientset-name=versioned \
  -h hack/boilerplate.go.txt \
  --input-base "" \
  --input=github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1 \
  --output-package=github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/clientset

# controller-gen doesn't currently work with apiserver extensions. =/
# controller-gen crd paths=./pkg/apis/...
