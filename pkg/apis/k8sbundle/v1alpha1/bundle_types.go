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

// UnstructuredJSON represents JSON not bound to any particular schema
type UnstructuredJSON map[string]interface{}

// ClusterBundleSpec is the the specification for the cluster bundle.
type ClusterBundleSpec struct {
	// Version-string for this bundle. The version should be a SemVer string (see
	// https://semver.org/) of the form X.Y.Z (Major.Minor.Patch).  Generally
	// speaking, a major-version (changes should indicate breaking changes,
	// minor-versions should indicate backwards compatible features, and patch
	// changes should indicate backwords compatible. If there are any changes to
	// the bundle, then the version string must be incremented.
	//
	// If a bundle is versioned, then all its components must be versioned.
	Version string `json:"version,omitempty"`

	// Kubernetes objects grouped into component packages and versioned together.
	// These could be applications or they could be some sort of supporting
	// object collection.
	Components []ComponentPackage `json:"componentPackage,omitempty"`

	// Cluster components that are specified externally as Files. The process of inlining
	// for a bundle reads component files into components, and so after
	// inlining, this list will be empty.
	ComponentFiles []File `json:"componentFiles,omitempty"`
}

// File represents some sort of file that's specified external to the bundle,
// which could be on either a local or remote file system.
type File struct {
	// URL to find this file.
	URL string `json:"url,omitempty"`

	// Optional Sha256 hash of the binary to ensure we are pulling the correct
	// binary/file.
	Hash string `json:"hash,omitempty"`
}

// ComponentPackageSpec represents the spec for the component.
type ComponentPackageSpec struct {
	// Version-string for this component. The version should be a SemVer 2 string
	// (see https://semver.org/) of the form X.Y.Z (Major.Minor.Patch).  A
	// major-version changes should indicate breaking changes, minor-versions
	// should indicate backwards compatible features, and patch changes should
	// indicate backwords compatible. If there are any changes to the component,
	// then the version string must be incremented.
	Version string `json:"version,omitempty"`

	// Optional. A version-string representing the version of the API this
	// component offers to other components (for the purposes of requirement
	// satisfaction). If no API is offered to other components, then this field
	// may be blank, but then other components may not depend on this component.
	ComponentAPIVersion string `json:"componentApiVersion,omitempty"`

	// A list of components that must be packaged in a bundle in with this
	// component.
	Requirements []MinRequirement `json:"requirements,omitempty"`

	// Structured Kubenetes objects that run as part of this app, whether on the
	// master, on the nodes, or in some other fashio.  These Kubernetes objects
	// are inlined and must be YAML/JSON compatible. Each must have `apiVersion`,
	// `kind`, and `metadata`.
	//
	// This is essentially equivalent to the Kubernetes `Unstructured` type.
	ClusterObjects []UnstructuredJSON `json:"clusterObjects,omitempty"`

	// Cluster objects that are specified via a File-URL. The process of inlining
	// the a component turns cluster object files into cluster objects.
	// During the inline process, if the file is YAML-formatted and contains multiple
	// objects, the objects will be split into separate inline objects. In other
	// words, one cluster object file may result in multiple cluster objects.
	//
	// Each cluster object file must be parseable into a Struct: In other words,
	// it should be representable as either YAML or JSON.
	ClusterObjectFiles []File `json:"clusterObjectFiles,omitempty"`

	// Raw files represent arbitrary string data. Unlike cluster object files,
	// these files don't need to be parseable as YAML or JSON. So, during the
	// inline process, the data is inserted into a generated config map before
	// being added to the cluster objects. A ConfigMap is generated per-file,
	// with the metadata.name and the data-key both being set to the base-file
	// name. Thus, if the url is something like 'file://foo/bar/biff.txt', the
	// metadata.name and data-key will be 'biff.txt'.
	RawTextFiles []File `json:"rawTextFiles,omitempty"`
}

// MinRequirement represents a component that this component must be packaged
// with in a ClusterBundle (or must be satisfied in some other fashion). This is
// roughly based on Semantic Import Versioning in the Go Language:
// https://research.swtch.com/vgo-import.
type MinRequirement struct {
	// Name of a component. The component name specified here must match exactly a
	// component name in the `metadata.name` field of another component.
	Component string `json:"component,omitempty"`

	// The sem-ver apiVersion of the component. The API Version is only a minimum
	// requirement. The assumption any newer component with only backwards
	// compatable changes is acceptable.
	ComponentAPIVersion string `json:"componentApiVersion,omitempty"`
}

// +genclient

// The ClusterBundle is a packaging format for Kubernetes Components.
type ClusterBundle struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// The specification object for the ClusterBundle.
	Spec ClusterBundleSpec `json:"spec,omitempty"`
}

// +genclient

// ComponentPackage represents Kubernetes objects grouped into cluster
// components and versioned together. These could be applications or they
// could be some sort of supporting collection of objects.
type ComponentPackage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// The specification object for the ComponentPackage.
	Spec ComponentPackageSpec `json:"spec,omitempty"`
}
