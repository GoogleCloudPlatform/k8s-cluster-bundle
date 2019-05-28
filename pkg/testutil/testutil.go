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
	"fmt"
	"os"
	"path/filepath"
)

// ChangeToBazelDir changes the CWD to a bazel directory if necessary.
// pathToDir specifies the path from the root bazel-directory to the relevant
// directory.  If changing the directory fails, the method panics.
func ChangeToBazelDir(curDir string) {
	bazelTestPath := os.Getenv("TEST_SRCDIR")
	if bazelTestPath != "" {
		workspace := os.Getenv("TEST_WORKSPACE")
		dir := filepath.Join(bazelTestPath, workspace, curDir)
		if err := os.Chdir(dir); err != nil {
			panic(fmt.Sprintf("os.Chdir(%q): %v", dir, err))
		}
	}
}

// ChangeToBazelDirWithoutWorkspace, like ChangeToBazelDir, changes the CWD to
// a bazel directory if necessary, but doesn't use the workspace path as part
// of the path building. If changing the directory fails, the method panics.
func ChangeToBazelDirWithoutWorkspace(curDir string) {
	bazelTestPath := os.Getenv("TEST_SRCDIR")
	if bazelTestPath != "" {
		dir := filepath.Join(bazelTestPath, curDir)
		if err := os.Chdir(dir); err != nil {
			panic(fmt.Sprintf("os.Chdir(%q): %v", dir, err))
		}
	}
}
