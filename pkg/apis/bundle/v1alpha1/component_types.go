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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ComponentSetSpec represents a versioned selection of Kubernetes components.
type ComponentSetSpec struct {
	// SetName is the human-readable string for this group of components. It
	// must only contain lower case alphanumerics, periods, and dashes. See more
	// details at k8s.io/docs/concepts/overview/working-with-objects/names/
	SetName string `json:"setName,omitempty"`

	// Version is the required version string for this component set and should
	// have the form X.Y.Z (Major.Minor.Patch). Generally speaking, major-version
	// changes should indicate breaking changes, minor-versions should indicate
	// backwards compatible features, and patch changes should indicate backwords
	// compatible. If there are any changes to the bundle, then the version
	// string must be incremented. As such, the version should not be tied to the
	// version of the container images.
	Version string `json:"version,omitempty"`

	// Components are references to component objects that make up the component
	// set. To get the Metadata.Name for the component, GetLocalObjectRef()
	// should be called on the component reference.
	Components []ComponentReference `json:"components,omitempty"`
}

// ComponentSetReference provides a reference to a component set
type ComponentSetReference struct {
	// SetName is the readable name of a component set.
	SetName string `json:"setName,omitempty"`

	// Version is the version string for a component set.
	Version string `json:"version,omitempty"`
}

// ComponentReference provides a reference
type ComponentReference struct {
	// ComponentName is the readable name of a component.
	ComponentName string `json:"componentName,omitempty"`

	// Version is the version string for a component.
	Version string `json:"version,omitempty"`
}

// ComponentSpec represents the spec for the component.
type ComponentSpec struct {
	// ComponentName is the canonical name of this component. For example, 'etcd'
	// or 'kube-proxy'. It must have the same naming properties as the
	// Metadata.Name to allow for constructing the name.
	// See more at k8s.io/docs/concepts/overview/working-with-objects/names/
	ComponentName string `json:"componentName,omitempty"`

	// Version is the required version for this component. The version
	// should be a SemVer 2 string (see https://semver.org/) of the form X.Y.Z
	// (Major.Minor.Patch).  A major-version changes should indicate breaking
	// changes, minor-versions should indicate backwards compatible features, and
	// patch changes should indicate backwards compatible. If there are any
	// changes to the component, then the version string must be incremented.
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

	// Structured Kubenetes objects that run as part of this app, whether on the
	// master, on the nodes, or in some other fashio.  These Kubernetes objects
	// are inlined and must be YAML/JSON compatible. Each must have `apiVersion`,
	// `kind`, and `metadata`.
	//
	// This is essentially equivalent to the Kubernetes `Unstructured` type.
	Objects []*unstructured.Unstructured `json:"objects,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ComponentSetList contains a list of ComponentSets.
type ComponentSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ComponentSet `json:"items,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ComponentList contains a list of Components.
type ComponentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Component `json:"items,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ComponentSet references a precise set of component packages.
type ComponentSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// The specification object for the ComponentSet
	Spec ComponentSetSpec `json:"spec,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Component represents Kubernetes objects grouped into
// components and versioned together. These could be applications or they
// could be some sort of supporting collection of objects.
type Component struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// The specification object for the Component.
	Spec ComponentSpec `json:"spec,omitempty"`
}
