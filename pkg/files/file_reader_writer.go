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

package files

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
)

// FileWriter is an interface for writing files. This interface is
// used by the export and patch commands in this package. There is a fake
// implementation in the testing package.
type FileWriter interface {
	// WriteFile writes a Component or Bundle to the given file path.
	WriteFile(ctx context.Context, path string, bytes []byte, permissions os.FileMode) error
}

// LocalFileSystemWriter implements the ComponentWriter interface and writes
// apps to the local filesystem.
type LocalFileSystemWriter struct{}

// WriteFile writes a file to disk.
func (*LocalFileSystemWriter) WriteFile(_ context.Context, path string, bytes []byte, permissions os.FileMode) error {
	return ioutil.WriteFile(path, bytes, permissions)
}

// Ensure the LocalFileSystemReader fulfills the contract
var _ FileWriter = &LocalFileSystemWriter{}

// FileReader is a common command interface for reading files.
type FileReader interface {
	ReadFile(ctx context.Context, path string) ([]byte, error)
}

// LocalFileSystemReader implements the FileReader interface and reads
// files from the local filesystem.
type LocalFileSystemReader struct{}

// ReadFile reads a file from disk.
func (r *LocalFileSystemReader) ReadFile(_ context.Context, path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

// Ensure the LocalFileSystemReader fulfills the contract
var _ FileReader = &LocalFileSystemReader{}

// FileReaderWriter combines both file reading and file writing.
type FileReaderWriter interface {
	FileReader
	FileWriter
}

// LocalFileSystemReaderWriter combines both local file system file reading and
// writing.
type LocalFileSystemReaderWriter struct {
	LocalFileSystemReader
	LocalFileSystemWriter
}

// FileObjReader provides a generic file-reading interface for reading file
// objects
type FileObjReader interface {
	ReadFileObj(ctx context.Context, file bundle.File) ([]byte, error)
}

// LocalFileObjReader is File object reader that defers to another FileReader that
// reads based on paths.
type LocalFileObjReader struct {
	// WorkingDir specifies a working directory override. This is necessary
	// because paths for inlined files are specified relative to the bundle, not
	// the working directory of the user.
	//
	// TODO(kashomon): Get rid of this. Path manipulation should happen in the
	// downstream libraries.
	WorkingDir string

	// Rdr is a FileReader object.
	Rdr FileReader
}

// ReadFileObj reads a file object from the local filesystem by deferring to a
// local file reader.
func (r *LocalFileObjReader) ReadFileObj(ctx context.Context, file bundle.File) ([]byte, error) {
	url := file.URL
	if url == "" {
		return nil, fmt.Errorf("file %v was specified but no file url was provided", file)
	}
	if strings.HasPrefix(url, "file://") {
		url = strings.TrimPrefix(url, "file://")
	}
	return r.Rdr.ReadFile(ctx, filepath.Join(r.WorkingDir, url))
}
