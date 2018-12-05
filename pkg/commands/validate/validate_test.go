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
	"context"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/util/validation/field"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/cmdlib"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/testutil"
)

type fakeBundleValidator struct {
	errs field.ErrorList
}

func (f *fakeBundleValidator) AllComponents(_ []*bundle.ComponentPackage) field.ErrorList {
	return f.errs
}

func TestRunValidate(t *testing.T) {
	dummyFile := "/bundle.yaml"

	var testcases = []struct {
		testName          string
		errors            field.ErrorList
		expectErrContains string
	}{
		{
			testName: "success case",
		},
		{
			testName:          "bundle validation errors",
			errors:            field.ErrorList{field.Invalid(field.NewPath("f"), "z", "yar")},
			expectErrContains: "one or more errors",
		},
	}

	rw := testutil.NewFakeReaderWriterFromPairs(
		&testutil.FilePair{dummyFile, testutil.FakeComponentData})
	opts := &options{}
	for _, tc := range testcases {
		t.Run(tc.testName, func(t *testing.T) {
			ctx := context.Background()
			globalOpts := &cmdlib.GlobalOptions{
				BundleFile:  dummyFile,
				InputFormat: "yaml",
			}

			// Override the createValidatorFn to return a fake bundleValidator
			createValidatorFn = func() bundleValidator {
				return &fakeBundleValidator{errs: tc.errors}
			}

			err := runValidate(ctx, opts, rw, globalOpts)
			if (tc.expectErrContains != "" && err == nil) || (tc.expectErrContains == "" && err != nil) {
				t.Errorf("runInline() returned err: %v, Want Err: %v", err, tc.expectErrContains)
			}
			if err == nil {
				return
			}
			if !strings.Contains(err.Error(), tc.expectErrContains) {
				t.Errorf("runInline() returned unexpected error message: %v, Should contain: %v", err, tc.expectErrContains)
			}
		})
	}
}
