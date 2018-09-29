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

package converter

import (
	"context"
	"os"

	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/files"
)

// BundleReaderWriter uses a file reader writer to read, write. and convert bundles.
type BundleReaderWriter struct {
	// RW is a FileReaderWriter instance.
	RW files.FileReaderWriter
}

func NewFileSystemBundleReaderWriter() *BundleReaderWriter {
	return &BundleReaderWriter{
		&files.LocalFileSystemReaderWriter{},
	}
}

// ReadBundleFile reads in a ClusterBundle from a yaml file found at the
// given path.
func (r *BundleReaderWriter) ReadBundleFile(ctx context.Context, path string) (*bpb.ClusterBundle, error) {
	bytes, err := r.RW.ReadFile(ctx, path)
	if err != nil {
		return nil, err
	}
	b, err := Bundle.YAMLToProto(bytes)
	if err != nil {
		return nil, err
	}
	return ToBundle(b), nil
}

// WriteBundleFile writes a given ClusterBundle proto to a yaml file at the
// specified path with given permissions.
func (r *BundleReaderWriter) WriteBundleFile(ctx context.Context, path string, bundle *bpb.ClusterBundle, permissions os.FileMode) error {
	bytes, err := Bundle.ProtoToYAML(bundle)
	if err != nil {
		return err
	}
	return r.RW.WriteFile(ctx, path, bytes, os.FileMode(permissions))
}
