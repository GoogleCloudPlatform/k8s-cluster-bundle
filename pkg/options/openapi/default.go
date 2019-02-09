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

package openapi

import (
	"fmt"

	"github.com/go-openapi/validate/post"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/options"
)

// ApplyDefaults adds defaults to options, using an OpenAPI schema.
func ApplyDefaults(opts options.JSONOptions, optSchema *apiextv1beta1.JSONSchemaProps) (options.JSONOptions, error) {
	res, err := ValidateOptions(opts, optSchema)
	if err != nil {
		return nil, err
	}

	post.ApplyDefaults(res)

	data := res.Data()

	jsonData, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Could not convert option data to map[string]interface{}. was: %v", data)
	}

	return jsonData, nil
}
