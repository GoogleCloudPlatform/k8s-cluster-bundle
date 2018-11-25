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

// Identifier is a key used for a metadata label or annotation to experess
// extra information about components and objects.
type Identifier string

const (
	// InlineTypeIdentifier is an identifier used to identify how an object in a
	// component was inlined.
	InlineTypeIdentifier Identifier = "bundle.gke.io/inline-type"
)

// InlineType is a value that the InlineTypeIdentifier can take.
type InlineType string

const (
	// KubeObjectInline indicates the object was inlined as an untructured object.
	KubeObjectInline InlineType = "kube-object"

	// RawStringInline indicates the object was inlined via raw-strings into a
	// ConfigMap.
	RawStringInline InlineType = "raw-string"
)
