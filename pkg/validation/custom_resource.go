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

package validation

// Eventually, this will actually have validation logic via
// github.com/go-openapi/validate to validate custom resources via an OpenAPI
// Spec; but this needs to be imported. Until that point, this is stubbed.
//
// For an example of how this works in practice, see:
// https://github.com/kubernetes/apiextensions-apiserver/blob/master/pkg/apiserver/validation/validation.go

import (
	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
)

// CustomResourceOptions contains options for custom resource validation.
type CustomResourceOptions struct {
	// Should the validation be strict? In other words, if a Custom Resource is
	// not specified in the Bundle via a CRD and Strict is enabled, it will
	// result in an error.
	Strict bool
}

// CustomResourceValidator validates custom resources for Bundle customization.
type CustomResourceValidator struct {
	// Bundle represents a Cluster Bundle
	Bundle *bpb.ClusterBundle

	// Options for validation.
	Options *CustomResourceOptions
}

// NewCustomResourceValidator creates a new CustomResourceValidator.
func NewCustomResourceValidator(bundle *bpb.ClusterBundle, options *CustomResourceOptions) (*CustomResourceValidator, error) {
	return &CustomResourceValidator{bundle, options}, nil
}

// Validate checks whether a custom resource conforms to the spec given in the CRD.
func (c *CustomResourceValidator) Validate(customResource interface{}) error {
	// TODO: This needs to be filled out once open API validation has been imported.
	return nil
}
