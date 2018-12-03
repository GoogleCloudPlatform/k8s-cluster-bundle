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
set -o pipefail

# The release-1.10 version appears to work the best.
K8S_VERSION="release-1.10"

command -v deepcopy-gen >/dev/null 2>&1 || {
  pt1="Error: deepcopy-gen is required, but was not found. Download $K8S_VERSION of k8s.io/code-generator.\n"
  pt2="Then, install with 'go install k8s.io/code-generator/cmd/deepcopy-gen'"
  printf >&2 "${pt1}${pt2}"; exit 1;
}

command -v register-gen >/dev/null 2>&1 || {
  pt1="Error: register-gen is required, but was not found. Download $K8S_VERSION of k8s.io/code-generator.\n"
  pt2="Then, install with 'go install k8s.io/code-generator/cmd/register-gen'"
  printf >&2 "${pt1}${pt2}"; exit 1;
}

command -v client-gen >/dev/null 2>&1 || {
  pt1="Error: client-gen tool is required, but was not found. Download $K8S_VERSION of k8s.io/code-generator.\n"
  pt2="Then, install with 'go install k8s.io/code-generator/cmd/openapi-gen'"
  printf >&2 "${pt1}${pt2}"; exit 1;
}

command -v lister-gen >/dev/null 2>&1 || {
  pt1="Error: lister-gen tool is required, but was not found. Download $K8S_VERSION of k8s.io/code-generator.\n"
  pt2="Then, install with 'go install k8s.io/code-generator/cmd/openapi-gen'"
  printf >&2 "${pt1}${pt2}"; exit 1;
}

command -v crd >/dev/null 2>&1 || {
  pt1="Error: crd (generator) is required, but was not found. Download sigs.k8s.io/controller-tools.\n"
  pt2="Then, install with 'go install sigs.k8s.io/controller-tools/cmd/crd'"
  printf >&2 "${pt1}${pt2}"; exit 1;
}

# **Note**: For all commands, the working directory is the parent directory
REPO_ROOT=$(git rev-parse --show-toplevel)
cd "${REPO_ROOT}"

deepcopy-gen \
  -h hack/boilerplate.go.txt \
  -O zz_generated.deepcopy \
  --input-dirs=github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1 \
  --output-package=github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1

register-gen \
  -h hack/boilerplate.go.txt \
  --input-dirs=github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1 \
  --output-package=github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1

client-gen --clientset-name=versioned \
  -h hack/boilerplate.go.txt \
  --input-base "" \
  --input=github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1 \
  --output-package=github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/clientset

# TODO(kashomon): Should we add this?
# lister-gen \
  # -h hack/boilerplate.go.txt \
  # --input-dirs=github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1 \
  # --output-package=github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/listers

# TODO(kashomon): Because we're using 1.9 still, it means that there are some
# breakages that need to be manually cleaned up when this is run, if client-gen
# is run at head. For example, 'Must' from utilruntime must be removed.
files=$(find ./pkg/clientset -type f -name *.go)
for f in $files
do
  echo "Fixing ${f}"
  # Not using sed -i because darwin / bash have different behaviors for -i =(

  # Replace googlecloudplatform with GoogleCloudPlatform
  sed $'s/googlecloudplatform/GoogleCloudPlatform/' $f > $f.t
  mv $f.t $f

done

# Relies on ../PROJECT file
# creates CRDS in ../config/crds/
crd generate --skip-map-validation
