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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Bundle encapsulates component data, which can be used for build systems,
// release automation, libraries, and tooling. It is not intended for building
// controllers or even applying directly to clusters, and as such, is
// intentionally designed with and spec/status fields and without a generated
// client library.
type Bundle struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// SetName is the human-readable string for this set of components. It
	// must only contain lower case alphanumerics, periods, and dashes. See more
	// details at k8s.io/docs/concepts/overview/working-with-objects/names/
	SetName string `json:"setName,omitempty"`

	// Version is the required version string for this component set and should
	// have the form X.Y.Z (Major.Minor.Patch). Generally speaking, major-version
	// changes should indicate breaking changes, minor-versions should indicate
	// backwards compatible features, and patch changes should indicate backwords
	// compatible. If there are any changes to the bundle, then the version
	// string must be incremented.
	Version string `json:"version,omitempty"`

	// ComponentFiles reference ComponentPackage files. The component files can
	// be inlined as directly included components (see `Components` below).
	ComponentFiles []File `json:"componentFiles,omitempty"`

	// Components are ComponentPackages are files that are inlined. The
	// components must be unique based on the combination of ComponentName +
	// Version.
	Components []*ComponentPackage `json:"components,omitempty"`
}
