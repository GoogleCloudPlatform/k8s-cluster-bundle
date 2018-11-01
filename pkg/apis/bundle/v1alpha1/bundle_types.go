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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ComponentSetSpec represents a versioned selection of Kubernetes components.
type ComponentSetSpec struct {
	// Version is the required version string for this component set and should
	// have the form X.Y.Z (Major.Minor.Patch). Generally speaking, major-version
	// changes should indicate breaking changes, minor-versions should indicate
	// backwards compatible features, and patch changes should indicate backwords
	// compatible. If there are any changes to the bundle, then the version
	// string must be incremented.
	Version string `json:"version,omitempty"`

	// Components are references to component objects that make up the component
	// set. For the referenced components, APIVersion must be the same as the
	// ComponentSet and the Kind must be 'ComponentPackage'.
	Components []corev1.LocalObjectReference
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
	// CanonicalName is the canonical name of this component. For example, 'etcd' or
	// 'kube-proxy'.
	CanonicalName string `json:"canonicalName,omitempty"`

	// Version is the required version for this component. The version
	// should be a SemVer 2 string (see https://semver.org/) of the form X.Y.Z
	// (Major.Minor.Patch).  A major-version changes should indicate breaking
	// changes, minor-versions should indicate backwards compatible features, and
	// patch changes should indicate backwards compatible. If there are any
	// changes to the component, then the version string must be incremented.
	Version string `json:"version,omitempty"`

	// Structured Kubenetes objects that run as part of this app, whether on the
	// master, on the nodes, or in some other fashio.  These Kubernetes objects
	// are inlined and must be YAML/JSON compatible. Each must have `apiVersion`,
	// `kind`, and `metadata`.
	//
	// This is essentially equivalent to the Kubernetes `Unstructured` type.
	Objects []*unstructured.Unstructured `json:"objects,omitempty"`

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
	// being added to the objects. A ConfigMap is generated per-file,
	// with the metadata.name and the data-key both being set to the base-file
	// name. Thus, if the url is something like 'file://foo/bar/biff.txt', the
	// metadata.name and data-key will be 'biff.txt'.
	RawTextFiles []File `json:"rawTextFiles,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// The ComponentSet references a precise set of component packages.
type ComponentSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// The specification object for the ComponentSet
	Spec ComponentSetSpec `json:"spec,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ComponentPackage represents Kubernetes objects grouped into
// components and versioned together. These could be applications or they
// could be some sort of supporting collection of objects.
type ComponentPackage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// The specification object for the ComponentPackage.
	Spec ComponentPackageSpec `json:"spec,omitempty"`
}
