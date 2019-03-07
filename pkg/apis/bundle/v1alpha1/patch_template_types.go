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

// ObjectSelector is used for identifying Objects on which to apply the patch template.
type ObjectSelector struct {
        // Kinds represent the Kinds to filter on.
        Kinds []string

        // Names represent the names to filter on. For objects, this is the
        // metadata.name field. For components, this is the ComponentName.
        Names []string

        // Annotations contain key/value pairs to filter on. An empty string value matches
        // all annotation-values for a particular key.
        Annotations map[string]string

        // Labels contain key/value pairs to filter on. An empty string value matches
        // all label-values for a particular key.
        Labels map[string]string

        // Namespaces to filter on.
        Namespaces []string

        // Invert the match
        NegativeMatch bool
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PatchTemplate contains configuration for patching objects.
type PatchTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Template is a template that creates a patch for a K8S object. In other
	// words, a templated YAML blob that's meant to be applied via
	// strategic-merge-patch. It's currently assumed to be a YAML go-template.
	Template string `json:"template,omitempty"`

	// Selector identifies the objects to which the patch should be applied
	// For each object selected, the template will have its apiVersion and
	// kind set to match the object, then be applied to the object.
	Selector ObjectSelector `json:"selector,omitempty"`

	// OptionsSchema is the schema for the parameters meant to be applied to
	// the patch template.
	OptionsSchema *apiextensions.JSONSchemaProps `json:"optionsSchema,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PatchTemplateBuilder contains configuration for creating patch templates.
type PatchTemplateBuilder struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Template is a template that creates a patch for a K8S object. In other
	// words, a templated YAML blob that's meant to be applied via
	// strategic-merge-patch. It's currently assumed to be a YAML go-template.
	Template string `json:"template,omitempty"`

	// Selector identifies the objects to which the patch should be applied
	// For each object selected, the template will have its apiVersion and
	// kind set to match the object, then be applied to the object.
	Selector ObjectSelector `json:"selector,omitempty"`

	// BuildSchema is the schema for the parameters meant to be applied to
	// the patch template.
	BuildSchema *apiextensions.JSONSchemaProps `json:"buildSchema,omitempty"`

	// TargetSchema is the schema for the parameters after the build-phase. This
	// becomes the 'OptionsSchema' field.
	TargetSchema *apiextensions.JSONSchemaProps `json:"targetSchema,omitempty"`
}
