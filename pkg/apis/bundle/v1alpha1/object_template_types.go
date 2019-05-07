// Copyright 2019 Google LLC
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

// TemplateType indicates the type of the template.
type TemplateType string

const (
	// TemplateTypeUndefined represents an undefined template type.
	TemplateTypeUndefined TemplateType = ""

	// TemplateTypeGo represents a go-template, which is assumed to be YAML.
	TemplateTypeGo TemplateType = "go-template"

	// TemplateTypeJsonnet represents a jsonnet type.
	TemplateTypeJsonnet TemplateType = "jsonnet"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ObjectTemplateBuilder contains configuration for creating ObjectTemplates.
type ObjectTemplateBuilder struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// File is references a template file. Files can contain one or more objects.
	// If the object contains more.
	File File `json:"file,omitempty"`

	// Type indicates how the template should be detemplatized. It defaults to Go
	// Templates during build if left unspecified.
	Type TemplateType `json:"type,omitempty"`

	// OptionsSchema is the schema for the parameters meant to be applied to
	// the object template, which includes both defaulting and validation.
	OptionsSchema *apiextensions.JSONSchemaProps `json:"optionsSchema,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ObjectTemplate contains configuration for creating objects.
type ObjectTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Template is a template-string that creates one or more Kubernetes object.
	Template string `json:"template,omitempty"`

	// Type indicates how the template should be detemplatized and is required.
	Type TemplateType `json:"type,omitempty"`

	// OptionsSchema is the schema for the parameters meant to be applied to
	// the object template, which includes both defaulting and validation.
	OptionsSchema *apiextensions.JSONSchemaProps `json:"optionsSchema,omitempty"`
}
