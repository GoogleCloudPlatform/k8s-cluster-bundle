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

package testing

import (
	"fmt"
	"os"
	"strings"

	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
)

// FakeAppWriter is a fake implementation of appWriter for unit tests.
type FakeAppWriter struct {
	ValidPath string
	ValidDir  string
}

// NewFakeAppWriterForPath returns a FakeAppWriter that will fake a successful write when the
// given ValidPath is passed into the write functions.
func NewFakeAppWriterForPath(s string) *FakeAppWriter {
	return &FakeAppWriter{ValidPath: s}
}

// NewFakeAppWriterForDir returns a FakeAppWriter that will fake a successful write when the
// path passed to the write functions is in the given ValidDir.
func NewFakeAppWriterForDir(s string) *FakeAppWriter {
	return &FakeAppWriter{ValidDir: s}
}

// WriteAppToFile fakes an app write by returning nil if ValidString is passed as the path.
// Otherwise it returns an error.
func (f *FakeAppWriter) WriteAppToFile(_ *bpb.ClusterApplication, path string, _ os.FileMode) error {
	if f.ValidPath != "" && path == f.ValidPath {
		return nil
	}
	if f.ValidDir != "" && strings.Contains(path, f.ValidDir) {
		return nil
	}
	return fmt.Errorf("error writing application")
}
