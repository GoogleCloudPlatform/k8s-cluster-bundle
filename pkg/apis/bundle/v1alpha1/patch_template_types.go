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
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PatchTemplate contains configuration for patching objects.
type PatchTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Template is a template that creates a patch for a K8S object. In other
	// words, a templated YAML blob that's meant to be applied via
	// strategic-merge-patch. It's currently assumed to be a YAML go-template.
	Template string

	// OptionsSchema is the schema for the parameters meant to be applied to
	// the patch template.
	OptionsSchema *apiextensions.JSONSchemaProps
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PatchTemplateBuilder contains configuration for creating patch templates.
type PatchTemplateBuilder struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Template is a template that creates a patch for a K8S object. In other
	// words, a templated YAML blob that's meant to be applied via
	// strategic-merge-patch. It's currently assumed to be a YAML go-template.
	Template string

	// BuildSchema is the schema for the parameters meant to be applied to
	// the patch template.
	BuildSchema *apiextensions.JSONSchemaProps

	// TargetSchema is the schema for the parameters after the build-phase. This
	// becomes the 'OptionsSchema' field.
	TargetSchema *apiextensions.JSONSchemaProps
}
