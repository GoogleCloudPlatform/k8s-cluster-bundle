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
	// UndefinedTemplateType represents an undefined template type.
	UndefinedTemplateType TemplateType = ""

	// GoTemplate represents a go-template type.
	GoTemplate TemplateType = "go-template"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ObjectTemplateBuilder contains configuration for creating ObjectTemplates.
type ObjectTemplateBuilder struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// TemplateFile is references a template file
	TemplateFile File `json:"templateFile,omitempty"`

	// TemplateType indicates how the template should be detemplatized. By
	// default, it defaults to Go-Templates during build if left unspecified.
	TemplateType TemplateType `json:"templateType,omitempty"`

	// OptionsSchema is the schema for the parameters meant to be applied to
	// the object template, which includes both defaulting and validation.
	OptionsSchema *apiextensions.JSONSchemaProps `json:"optionsSchema,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ObjectTemplate contains configuration for creating objects.
type ObjectTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Template is a template-string that creates a K8S object.
	Template string `json:"template,omitempty"`

	// TemplateType is required and indicates how the template should be
	// detemplatized.
	TemplateType TemplateType `json:"templateType,omitempty"`

	// OptionsSchema is the schema for the parameters meant to be applied to
	// the object template, which includes both defaulting and validation.
	OptionsSchema *apiextensions.JSONSchemaProps `json:"optionsSchema,omitempty"`
}
