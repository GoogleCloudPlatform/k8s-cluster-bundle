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

# The release-1.10 version of the tooling appears to work the best, although
# there are some mismatches since the kube-apimachinery version is set to 1.9.
K8S_VERSION="release-1.10"

command -v deepcopy-gen >/dev/null 2>&1 || {
  pt1="Error: deepcopy-gen is required, but was not found. Download $K8S_VERSION of k8s.io/code-generator.\n"
  pt2="Then, install with 'go install k8s.io/code-generator/cmd/deepcopy-gen'"
  printf >&2 "${pt1}${pt2}"; exit 1;
}

command -v register-gen >/dev/null 2>&1 || {
  pt1="Error: register-gen is required, but was not found. Download release-1.12 of k8s.io/code-generator.\n"
  pt2="Then, install with 'go install k8s.io/code-generator/cmd/register-gen'"
  printf >&2 "${pt1}${pt2}"; exit 1;
}

command -v client-gen >/dev/null 2>&1 || {
  pt1="Error: client-gen tool is required, but was not found. Download $K8S_VERSION of k8s.io/code-generator.\n"
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

# TODO(#115): Because we're using apimachinery at 1.9 and tools at 1.10, it
# means that there are some breakages that need to be fixed up.
clientsetpattern="*clientset_generated.go"
files=$(find ./pkg/clientset -type f -name *.go)
for f in $files
do
  echo "Fixing ${f}"
  # Not using sed -i because darwin / bash have different behaviors for -i =(

  # TODO(#115): Replace googlecloudplatform with GoogleCloudPlatform. This is
  # fixed with controller-tools >= release-1.11.
  sed $'s/googlecloudplatform/GoogleCloudPlatform/' $f > $f.t
  mv $f.t $f

  # TODO(#115): These issues are caused because we controller tools at 1.10 and
  # apimachinery/client-go at 1.9.
  if [[ $f == $clientsetpattern ]]; then
    # Watch is not supported in testing.ObjectTracker in 1.9
    sed '/fakePtr.AddWatchReactor/,+9d' $f > $f.t
    mv $f.t $f

    # remove the import
    sed '/"k8s\.io\/apimachinery\/pkg\/watch"/d' $f > $f.t
    mv $f.t $f
  fi
done

# Relies on ../PROJECT file
# creates CRDS in ../config/crds/
crd generate
