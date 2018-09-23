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

package cmdlib

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"

	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
)

// BundleReaderWriter is a bundle file i/o interface for bundler commands to use.
type BundleReaderWriter interface {
	// ReadBundleFile reads in a ClusterBundle from a yaml file found at the
	// given filepath.
	ReadBundleFile(filepath string) (*bpb.ClusterBundle, error)

	// WriteBundleFile writes a given ClusterBundle proto to a yaml file at the
	// specified filepath with given permissions.
	WriteBundleFile(filepath string, bundle *bpb.ClusterBundle, permissions os.FileMode) error
}

// RealReaderWriter implements the BundleReaderWriter interface and
// reads/writes a bundle using the local filesystem.
type RealReaderWriter struct{}

func (r *RealReaderWriter) ReadBundleFile(filepath string) (*bpb.ClusterBundle, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	return parseBundleYAML(file)
}

func parseBundleYAML(reader io.Reader) (*bpb.ClusterBundle, error) {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	b, err := converter.Bundle.YAMLToProto(bytes)
	if err != nil {
		return nil, err
	}

	return converter.ToBundle(b), nil
}

func (r *RealReaderWriter) WriteBundleFile(filepath string, bundle *bpb.ClusterBundle, permissions os.FileMode) error {
	yaml, err := converter.Bundle.ProtoToYAML(bundle)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath, yaml, os.FileMode(permissions))
}
