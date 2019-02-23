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

// BundleBuilder builds component data. Practically speaking, the BundleBuilder
// builds Bundles, ComponentSets, and Components.
type BundleBuilder struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// SetName for the resulting Bundle and ComponentSet. The combination of
	// SetName and Version should provide a unique identifier for the generate.
	SetName string `json:"setName,omitempty"`

	// Version for the Bundle and ComponentSet. See Bundle.Version for more
	// details. The version is optional for the ComponentBuilder
	Version string `json:"version,omitempty"`

	// ComponentNamePolicy defines how to generate the metadata.name
	// for a Component or ComponentBuilder that does not already have one.
	//  - SetAndComponent generates a name from the set name and version and
	//    component name and version.
	//  - Component (default) generates a name from the component name and version
	ComponentNamePolicy string `json:"componentNamePolicy,omitempty"`

	// ComponentFiles represent ComponentBuilder or Component types that are
	// referenced via file urls.
	ComponentFiles []File `json:"componentFiles,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ComponentBuilder builds Components.
type ComponentBuilder struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// ComponentName is the canonical name of this component. See
	// ComponentSpec.ComponentName for more details.
	ComponentName string `json:"componentName,omitempty"`

	// Version is the version for this component. See ComponentSpec.Version for
	// more details. The version is optional for the ComponentBuilder.
	Version string `json:"version,omitempty"`

	// AppVersion specifies the application version that the component provides
	// and should have the form X.Y or X.Y.Z (Major.Minor.Patch). The AppVersion
	// will frequently be related to the version of the container image used by
	// the application and need not be updated when a component Version field is
	// updated, unless the application contract changes.
	//
	// For example, for an Etcd component, the version field might be something
	// like 10.9.8, but the app version would probalby be something like 3.3.10,
	// representing the version of Etcd application.
	//
	// In order for component A to depend on component B, component B must
	// specify a Requirements object with an AppVersion. Eliding the AppVersion
	// prevents other components from depending on your component.
	AppVersion string `json:"appVersion,omitempty"`

	// Objects that are specified via a File-URL. The process of inlining a
	// component turns object files into objects.  During the inline process, if
	// the file is YAML-formatted and contains multiple objects in the YAML-doc,
	// the objects will be split into separate inline objects. In other words,
	// one object file may result in multiple objects.
	//
	// Each object file must be parsable into a Struct: In other words,
	// it should be representable as either YAML or JSON.
	ObjectFiles []File `json:"objectFiles,omitempty"`

	// Raw files represent arbitrary string data. Unlike object files,
	// these files don't need to be parsable as YAML or JSON. So, during the
	// inline process, the data is inserted into a generated config map before
	// being added to the objects. A ConfigMap is generated per-filegroup.
	RawTextFiles []FileGroup `json:"rawTextFiles,omitempty"`
}
