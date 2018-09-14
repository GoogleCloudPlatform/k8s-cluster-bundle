#!/bin/bash

set -e

cd "$(dirname "$0")"

# Change to top-level directory.
cd ../../../../

DIR=$PWD

protoc --go_out=paths=source_relative:. \
  -I ./ \
  -I ./deps/ \
  ./pkg/apis/bundle/v1alpha1/*.proto
