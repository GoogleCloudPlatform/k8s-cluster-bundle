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
)

// FakeFileReaderWriter is a fake implementation of FileReaderWriter for unit
// tests.
type FakeFileReaderWriter struct {
	// ReadFiles contains mapping of path-to-file for files to-read. If
	// AlwaysRead is empty and the a path is supplied to read that is not in the
	// ReadFiles, an error will be returned.
	ReadFiles map[string]string

	// If AlwaysRead is present, always return the always read object.
	AlwaysRead string

	// ReadErr forces an error to occur during read.
	ReadErr error

	// WriteFiles records what files have been written
	WriteFiles map[string]string

	// WriteErr forces an error to occur during write.
	WriteErr error
}

// NewEmptyReaderWriter creates an empty FakeFileReaderWriter. It will return an
// error on all paths returned from reads and succeed on all writes.
func NewEmptyReaderWriter() *FakeFileReaderWriter {
	return &FakeFileReaderWriter{
		ReadFiles:  make(map[string]string),
		WriteFiles: make(map[string]string),
	}
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

// AddReadFile adds a file to the ReadFiles map.
func (f *FakeFileReaderWriter) AddReadFile(fp *FilePair) {
	f.ReadFiles[fp.Path] = fp.Contents
}

// ReadFile reads a file from the map
func (f *FakeFileReaderWriter) ReadFile(_ context.Context, path string) ([]byte, error) {
	if f.ReadErr != nil {
		return nil, f.ReadErr
	}
	if f.AlwaysRead != "" {
		return []byte(f.AlwaysRead), nil
	}
	if contents, ok := f.ReadFiles[path]; ok {
		return []byte(contents), nil
	}
	return nil, fmt.Errorf("error reading bundle file: path not found %q", path)
}

// ReadFileObj reads a File object by deferring to the internal map.
func (f *FakeFileReaderWriter) ReadFileObj(ctx context.Context, file bundle.File) ([]byte, error) {
	return f.ReadFile(ctx, file.URL)
}

// WriteFile checks write conditions based on path contents.
func (f *FakeFileReaderWriter) WriteFile(_ context.Context, path string, contents []byte, permissions os.FileMode) error {
	if f.WriteErr != nil {
		return f.WriteErr
	}
	f.WriteFiles[path] = string(contents)
	return nil
}
