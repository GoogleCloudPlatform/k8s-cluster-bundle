#!/bin/bash
set -o xtrace
set -o errexit
set -o nounset
set -o pipefail

cd "$(dirname ${BASH_SOURCE[0]})"

export GO111MODULE=on
go mod tidy
go mod vendor

find vendor/ -type f \( \
    -name 'BUILD' \
    -o -name 'BUILD.bazel' \
    -o -name '*.bzl' \
  \) -exec rm {} \;
bazel run //:gazelle
