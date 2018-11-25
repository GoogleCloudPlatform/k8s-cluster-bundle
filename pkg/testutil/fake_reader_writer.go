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

package testutil

import (
	"context"
	"fmt"
	"os"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/files"
)

// FakeComponentData is simple fake component data string that should always
// parse.
var FakeComponentData = `
components:
- apiVersion: bundle.gke.io/v1alpha1
  kind: ComponentPackage
  metadata:
    name: test-pkg
  spec:
    componentName: test-comp
    version: 0.1.0
`

// FakeReaderWriter is a fake implementation of FileReaderWriter for unit
// tests. If a path is not in the PathToFiles, both reads and writes will fail.
type FakeReaderWriter struct {
	PathToFiles map[string]string
}

// NewEmptyReaderWriter creates an empty FakeReaderWriter. It will return an
// error on all paths returned from reads and writes.
func NewEmptyReaderWriter() *FakeReaderWriter {
	return &FakeReaderWriter{make(map[string]string)}
}

// NewFakeReaderWriter returns a FakeReaderWriter that will fake a successful read/write when the
// given validFile is passed into the ReadBundleFile or WriteBundleFile functions.
func NewFakeReaderWriter(files map[string]string) *FakeReaderWriter {
	return &FakeReaderWriter{files}
}

// FilePair is a helper for constructing a fake file reader
type FilePair struct {
	// Path is some fake file path
	Path string

	// Contents is some expected contents
	Contents string
}

func (f *FilePair) String() string {
	return fmt.Sprintf("{Path: %q  Contents:%s}", f.Path, f.Contents)
}

// NewFakeReaderWriterFromPairs creates a map based on pairs of string inputs
func NewFakeReaderWriterFromPairs(pairs ...*FilePair) *FakeReaderWriter {
	m := make(map[string]string)
	for _, v := range pairs {
		m[v.Path] = v.Contents
	}
	return NewFakeReaderWriter(m)
}

// ReadFile reads a file from the map
func (f *FakeReaderWriter) ReadFile(_ context.Context, path string) ([]byte, error) {
	if contents, ok := f.PathToFiles[path]; ok {
		return []byte(contents), nil
	}
	return nil, fmt.Errorf("error reading bundle file: path not found %q", path)
}

// ReadFilePB reads a File proto object by deferring to the internal map.
func (f *FakeReaderWriter) ReadFileObj(ctx context.Context, file bundle.File) ([]byte, error) {
	return f.ReadFile(ctx, file.URL)
}

// WriteFile checks write conditions based on path contents.
func (f *FakeReaderWriter) WriteFile(_ context.Context, path string, bytes []byte, permissions os.FileMode) error {
	_, ok := f.PathToFiles[path]
	if !ok {
		return fmt.Errorf("error writing file: path not found %q ", path)
	}
	return nil
}

var _ files.FileReaderWriter = &FakeReaderWriter{}
