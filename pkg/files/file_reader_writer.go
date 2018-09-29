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

	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
)

// FileWriter is an interface for writing files. This interface is
// used by the export and patch commands in this package. There is a fake
// implementation in the testing package.
type FileWriter interface {
	// WriteFile writes a ClusterComponent to the given file path.
	WriteFile(ctx context.Context, path string, bytes []byte, permissions os.FileMode) error
}

// LocalFileSystemWriter implements the ComponentWriter interface and writes
// apps to the local filesystem.
type LocalFileSystemWriter struct{}

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

// RealFileReaderWriter combines both local file system file reading and writing
type LocalFileSystemReaderWriter struct {
	LocalFileSystemReader
	LocalFileSystemWriter
}

// FilePBReader provides a generic file-reading interface for reading file
// protos.
type FilePBReader interface {
	ReadFilePB(ctx context.Context, file *bpb.File) ([]byte, error)
}

// NewLocalFilePBReader creates a local-filesystem based file reader.
func NewLocalFilePBReader(dirOverride string) FilePBReader {
	return &LocalFilePBReader{
		WorkingDir: dirOverride,
		Rdr:        &LocalFileSystemReader{},
	}
}

// LocalFilePBReader is File proto reader that defers to another FileReader that
// reads based on paths.
type LocalFilePBReader struct {
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

// ReadFilePB reads a file proto from the local filesystem by deferring to a
// local file reader.
func (r *LocalFilePBReader) ReadFilePB(ctx context.Context, fpb *bpb.File) ([]byte, error) {
	url := fpb.GetUrl()
	if url == "" {
		return nil, fmt.Errorf("file %v was specified but no file url was provided", fpb)
	}
	if strings.HasPrefix(url, "file://") {
		url = strings.TrimPrefix(url, "file://")
	}
	return r.Rdr.ReadFile(ctx, filepath.Join(r.WorkingDir, url))
}
