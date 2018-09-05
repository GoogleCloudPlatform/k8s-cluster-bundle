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

	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
)

// FakeReaderWriter is a fake implementation of BundleReaderWriter for unit tests.
type FakeReaderWriter struct {
	ValidFile string
}

// NewFakeReaderWriter returns a FakeReaderWriter that will fake a successful read/write when the
// given validFile is passed into the ReadBundleFile or WriteBundleFile functions.
func NewFakeReaderWriter(validFile string) *FakeReaderWriter {
	f := FakeReaderWriter{ValidFile: validFile}
	return &f
}

// ReadBundleFile fakes a bundle read by returning an empty bundle if ValidBundleFile is passed.
// Otherwise it returns an error.
func (f *FakeReaderWriter) ReadBundleFile(filepath string) (*bpb.ClusterBundle, error) {
	if filepath == f.ValidFile {
		bundle := &bpb.ClusterBundle{}
		return bundle, nil
	}
	return nil, fmt.Errorf("error reading bundle file")
}

// WriteBundleFile fakes a bundle write by returning nil if ValidBundleFile is passed as the
// filepath. Otherwise it returns an error.
func (f *FakeReaderWriter) WriteBundleFile(filepath string, bundle *bpb.ClusterBundle, permissions os.FileMode) error {
	if filepath == f.ValidFile {
		return nil
	}
	return fmt.Errorf("error writing bundle file")
}
