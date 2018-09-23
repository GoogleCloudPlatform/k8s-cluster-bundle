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

package inline

import (
	"context"
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/testutil"
)

func TestRunInline(t *testing.T) {
	validFile := "/bundle.yaml"
	invalidFile := "/invalid.yaml"

	var testcases = []struct {
		testName          string
		opts              *options
		expectErrContains string
	}{
		{
			testName: "success case",
			opts: &options{
				bundle: "in" + validFile,
				output: "out" + validFile,
			},
		},
		{
			testName: "bundle read error",
			opts: &options{
				bundle: "in" + invalidFile,
				output: "out" + validFile,
			},
			expectErrContains: "error reading",
		},
		{
			testName: "bundle write error",
			opts: &options{
				bundle: "in" + validFile,
				output: "out" + invalidFile,
			},
			expectErrContains: "error writing",
		},
	}

	ctx := context.Background()
	for _, tc := range testcases {
		t.Run(tc.testName, func(t *testing.T) {
			var pairs []*testutil.FilePair
			if strings.Contains(tc.opts.bundle, validFile) {
				pairs = append(pairs, &testutil.FilePair{tc.opts.bundle, testutil.FakeBundle})
			}
			if strings.Contains(tc.opts.output, validFile) {
				pairs = append(pairs, &testutil.FilePair{tc.opts.output, testutil.FakeBundle})
			}

			frw := testutil.NewFakeReaderWriterFromPairs(pairs...)
			err := run(ctx, tc.opts, &converter.BundleReaderWriter{frw}, frw)
			if (tc.expectErrContains != "" && err == nil) || (tc.expectErrContains == "" && err != nil) {
				t.Errorf("runInline(opts: %+v) returned err: %v, Want Err: %q", tc.opts, err, tc.expectErrContains)
			}
			if err == nil {
				return
			}
			if !strings.Contains(err.Error(), tc.expectErrContains) {
				t.Errorf("run(opts: %+v) returned unexpected error message: %v, Should contain: %v", tc.opts, err, tc.expectErrContains)
			}
		})
	}
}
