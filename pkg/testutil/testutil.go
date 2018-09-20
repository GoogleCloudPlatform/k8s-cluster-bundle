// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package testutil provides utilities for reading testdata from children
// directories.
package testutil

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

const testDir = "pkg/testutil/testdata"

// TestPathPrefix returns the empty string or the bazel test path prefix.
func TestPathPrefix(inpath string) string {
	path := os.Getenv("TEST_SRCDIR") // For dealing with bazel.
	workspace := os.Getenv("TEST_WORKSPACE")
	if path != "" {
		return filepath.Join(path, workspace, testDir)
	}
	return inpath
}

// ReadTestBundle reads the test-Bundle from disk.
func ReadTestBundle(inpath string) ([]byte, error) {
	testpath := TestPathPrefix(inpath)
	return ioutil.ReadFile(filepath.Join(testpath, "example-bundle.yaml"))
}
