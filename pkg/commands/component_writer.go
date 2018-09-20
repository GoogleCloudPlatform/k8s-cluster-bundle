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

package commands

import (
	"io/ioutil"
	"os"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"

	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
)

// compWriter is an interface for writing applications.
// This interface is used by the export and patch commands in this package.
// There is a fake implementation in the testing package.
type compWriter interface {
	// WriteComponentToFile writes a ClusterComponent to the given file path.
	WriteComponentToFile(comp *bpb.ClusterComponent, path string, permissions os.FileMode) error
}

// localFileSystemWriter implements the appWriter interface and writes apps to the local filesystem.
type localFileSystemWriter struct{}

func (*localFileSystemWriter) WriteComponentToFile(comp *bpb.ClusterComponent, path string, permissions os.FileMode) error {
	yaml, err := converter.Struct.ProtoToYAML(comp)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, yaml, permissions)
}
