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

package validate

import (
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	internalapiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/options"
)

// ValidateOptions validates options based on a schema.
func ValidateOptions(opts options.JSONOptions, optSchema *apiext.JSONSchemaProps) error {
	if optSchema == nil {
		return nil
	}

	// We need to convert to the internal JSONSchemaProps, because that's what
	// the library deals with.
	intOptSchema := &internalapiext.JSONSchemaProps{}

	// TODO(kashomon): Should I make a runtime scheme and use that to convert
	// between the types? That seems to be more standard based on looking at
	// examples. Ideally, this wouldn't need to happen here -- there would be an
	// internal bundle library which used the internal JSON schema type; and then
	// conversion would happen at the leaves..
	err := apiext.Convert_v1beta1_JSONSchemaProps_To_apiextensions_JSONSchemaProps(optSchema, intOptSchema, nil /* conversion scope */)
	if err != nil {
		return err
	}

	openapiSchema := &spec.Schema{}
	if err := ConvertJSONSchemaProps(intOptSchema, openapiSchema); err != nil {
		return err
	}
	validator := validate.NewSchemaValidator(openapiSchema, nil, "", strfmt.Default)

	result := validator.Validate(opts)
	if result.AsError() != nil {
		return result.AsError()
	}
	return nil
}
