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

package v1alpha1

import (
	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Configuration for node images. This is a resource that provides information
// about which images are available for node creation and how to initialize the
// node images.
type NodeConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Note: only one of InitFile or ExternalInitFile should be specified.

	// The file specified inline as a UTF-8 encoded byte string.
	InitFile []byte `json:"initFile,omitempty"`

	// An externally specified init file.
	ExternalInitFile bundle.File `json:"externalInitFile,omitempty"`

	// Envirnoment variables to set before startup to configure the init script.
	EnvVars []EnvVar `json:"envVars,omitempty"`

	// The OS image to use for VM creation.
	OsImage bundle.File `json:"osImage,omitempty"`
}

// An environment variable specified for node startup.
type EnvVar struct {
	// Name of this environment variable. E.g., FOO_VAR. The name of the
	// environment variable should be unique within a node bootstrap
	// configuration.
	Name string `json:"name,omitempty"`

	// The value to set for this environment variable.
	Value string `json:"value,omitempty"`
}
