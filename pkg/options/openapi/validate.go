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

// Package openapi provides methods for using the OpenAPI schema for validation
// and defaulting.
package openapi

import (
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/options"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/validation"
)

// ValidateOptions based on an OpenAPI schema. The returned result object
// contains helpers for manipulating the data and also getting feedback on
// errors, and so may be non-nil in the case of an error when the error is the
// result of openapi validation failure.
func ValidateOptions(opts options.JSONOptions, optSchema *apiextv1beta1.JSONSchemaProps) (*validate.Result, error) {
	if optSchema == nil || opts == nil {
		return nil, nil
	}

	// We need to convert to the internal JSONSchemaProps, because that's what
	// the conversion library deals with.
	intOptSchema := &apiextensions.JSONSchemaProps{}

	// TODO(kashomon): Should I make a runtime scheme and use that to convert
	// between the types? That seems to be more standard based on looking at
	// examples. Ideally, this wouldn't need to happen here -- there would be an
	// internal bundle library which used the internal JSON schema type; and then
	// conversion would happen at the leaves.
	err := apiextv1beta1.Convert_v1beta1_JSONSchemaProps_To_apiextensions_JSONSchemaProps(optSchema, intOptSchema, nil /* conversion scope */)
	if err != nil {
		return nil, err
	}

	openapiSchema := &spec.Schema{}
	if err := validation.ConvertJSONSchemaProps(intOptSchema, openapiSchema); err != nil {
		return nil, err
	}
	validator := validate.NewSchemaValidator(openapiSchema, nil, "", strfmt.Default)

	// Convert back to map[string]interface{}, since that's what the validation expects.
	var conv map[string]interface{} = opts

	result := validator.Validate(conv)
	if result.AsError() != nil {
		return result, result.AsError()
	}
	return result, nil
}
