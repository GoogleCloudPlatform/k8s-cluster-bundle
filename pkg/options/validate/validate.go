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
	_ "github.com/go-openapi/spec"
	_ "github.com/go-openapi/strfmt"
	_ "github.com/go-openapi/validate"
	_ "github.com/go-openapi/validate/post"
	_ "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/options"
)

// ValidateOptions validates options based on a schema.
func ValidateOptions(opts options.JSONOptions, optSchema *apiext.JSONSchemaProps) error {
	if optSchema == nil {
		return nil
	}
	return nil
}
