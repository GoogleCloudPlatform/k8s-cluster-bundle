#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

command -v deepcopy-gen >/dev/null 2>&1 || {
  pt1="Error: deepcopy-gen is required, but was not found. Download k8s.io/code-generator.\n"
  pt2="Then, install with 'go install k8s.io/code-generator/cmd/deepcopy-gen'"
  printf >&2 "${pt1}${pt2}"; exit 1;
}

command -v register-gen >/dev/null 2>&1 || {
  pt1="Error: register-gen is required, but was not found. Download k8s.io/code-generator.\n"
  pt2="Then, install with 'go install k8s.io/code-generator/cmd/register-gen'"
  printf >&2 "${pt1}${pt2}"; exit 1;
}

command -v client-gen >/dev/null 2>&1 || {
  pt1="Error: client-gen tool is required, but was not found. Download k8s.io/code-generator.\n"
  pt2="Then, install with 'go install k8s.io/code-generator/cmd/openapi-gen'"
  printf >&2 "${pt1}${pt2}"; exit 1;
}

command -v crd >/dev/null 2>&1 || {
  pt1="Error: crd (generator) is required, but was not found. Download sigs.k8s.io/controller-tools.\n"
  pt2="Then, install with 'go install sigs.k8s.io/controller-tools/cmd/crd'"
  printf >&2 "${pt1}${pt2}"; exit 1;
}

cd "$(dirname "$0")"

# **Note**: For all commands, the working directory is the parent directory
cd ..

# We can't use deepcopy-gen until issues/128 is fixed (or we write our own copy logic).
# https://github.com/kubernetes/gengo/issues/128
deepcopy-gen \
  -h hack/boilerplate.go.txt \
  -O zz_generated.deepcopy \
  --input-dirs=github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/k8sbundle/v1alpha1 \
  --output-package=github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/k8sbundle/v1alpha1

register-gen \
  -h hack/boilerplate.go.txt \
  --input-dirs=github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/k8sbundle/v1alpha1 \
  --output-package=github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/k8sbundle/v1alpha1

client-gen --clientset-name=versioned \
  -h hack/boilerplate.go.txt \
  --input-base=github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/k8sbundle/v1alpha1 \
  --output-package=github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/clientset

# Relies on ../PROJECT file
# creates CRDS in ../config/crds/
crd generate --skip-map-validation

# Some hackery to rename k8sbundle to just bundle and remove k8sbundle
# references
for f in config/crds/*.yaml; do
  mv "$f" "$(echo "$f" | sed s/k8sbundle/bundle/)"
done

for f in config/crds/*.yaml; do
  sed 's/k8sbundle/bundle/g' $f > $f.t
  mv $f.t $f
done

  # --input-dirs=./pkg/apis/k8sbundle/v1alpha1 \
  # --included-types-overrides=gke.io/k8s-cluster-bundle/v1alpha1 \
