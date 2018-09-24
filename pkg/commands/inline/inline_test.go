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

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/cmdlib"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/testutil"
)

func TestRunInline(t *testing.T) {
	validInBundle := "/in-bundle.yaml"
	validOutBundle := "/out-bundle.yaml"
	invalidFile := "/invalid.yaml"

	var testcases = []struct {
		testName          string
		opts              *options
		in                string
		out               string
		expectErrContains string
	}{
		{
			testName: "success case",
			in:       validInBundle,
			out:      validOutBundle,
		},
		{
			testName:          "bundle read error",
			in:                invalidFile,
			out:               validOutBundle,
			expectErrContains: "error reading",
		},
		{
			testName:          "bundle write error",
			in:                validInBundle,
			out:               invalidFile,
			expectErrContains: "error writing",
		},
	}

	ctx := context.Background()
	for _, tc := range testcases {
		t.Run(tc.testName, func(t *testing.T) {
			var pairs []*testutil.FilePair
			if tc.in == validInBundle {
				pairs = append(pairs, &testutil.FilePair{tc.in, testutil.FakeBundle})
			}
			if tc.out == validOutBundle {
				pairs = append(pairs, &testutil.FilePair{tc.out, testutil.FakeBundle})
			}
			globalOpts := &cmdlib.GlobalOptions{
				BundleFile:   tc.in,
				OutputFile:   tc.out,
				InputFormat:  "yaml",
				OutputFormat: "yaml",
			}

			frw := testutil.NewFakeReaderWriterFromPairs(pairs...)
			err := run(ctx, tc.opts, &converter.BundleReaderWriter{frw}, globalOpts)
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
