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
	"errors"
	"strings"
	"testing"

	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	testutil "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/testutil"
)

type fakeBundleValidator struct {
	errs []error
}

func (f *fakeBundleValidator) Validate() []error {
	return f.errs
}

func TestRunValidate(t *testing.T) {
	validFile := "/bundle.yaml"

	var testcases = []struct {
		testName          string
		opts              *options
		errors            []error
		expectErrContains string
	}{
		{
			testName: "success case",
			opts: &options{
				bundle: validFile,
			},
		},
		{
			testName: "bundle validation errors",
			opts: &options{
				bundle: validFile,
			},
			errors:            []error{errors.New("yarr")},
			expectErrContains: "one or more errors",
		},
	}

	brw := &converter.BundleReaderWriter{
		testutil.NewFakeReaderWriterFromPairs(&testutil.FilePair{validFile, testutil.FakeBundle}),
	}
	for _, tc := range testcases {
		t.Run(tc.testName, func(t *testing.T) {
			ctx := context.Background()

			// Override the createValidatorFn to return a fake bundleValidator
			createValidatorFn = func(b *bpb.ClusterBundle) bundleValidator {
				return &fakeBundleValidator{errs: tc.errors}
			}

			err := runValidate(ctx, tc.opts, brw)
			if (tc.expectErrContains != "" && err == nil) || (tc.expectErrContains == "" && err != nil) {
				t.Errorf("runInline(opts: %+v) returned err: %v, Want Err: %v", tc.opts, err, tc.expectErrContains)
			}
			if err == nil {
				return
			}
			if !strings.Contains(err.Error(), tc.expectErrContains) {
				t.Errorf("runInline(opts: %+v) returned unexpected error message: %v, Should contain: %v", tc.opts, err, tc.expectErrContains)
			}
		})
	}
}
