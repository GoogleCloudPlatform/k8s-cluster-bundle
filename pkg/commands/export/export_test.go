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

package export

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/cmdlib"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/testutil"
)

const (
	invalidComponent = "not-a-component"
)

// Fake implementation of compfinder for unit tests.
type fakeFinder struct{}

func (f *fakeFinder) ClusterComponent(compName string) *bpb.ClusterComponent {
	if compName == invalidComponent {
		return nil
	}
	return &bpb.ClusterComponent{
		Metadata: &bpb.ObjectMeta{
			Name: compName,
		},
	}
}

func TestRunExport(t *testing.T) {
	validBundle := "/bundle.yaml"
	invalidBundle := "/invalid.yaml"
	validDir := "/valid/dir"
	invalidDir := "/invalid/dir"

	var testcases = []struct {
		testName          string
		in                string
		opts              *options
		expectErrContains string
	}{
		{
			testName: "success case",
			in:       validBundle,
			opts: &options{
				outputDir:  validDir,
				components: []string{"kube-apiserver", "kube-scheduler"},
			},
		},
		{
			testName: "bundle read error",
			in:       invalidBundle,
			opts: &options{
				outputDir: validDir,
			},
			expectErrContains: "error reading",
		},
		{
			testName: "extract component error",
			in:       validBundle,
			opts: &options{
				outputDir:  validDir,
				components: []string{invalidComponent},
			},
			expectErrContains: invalidComponent,
		},
		{
			testName: "component write error",
			in:       validBundle,
			opts: &options{
				outputDir:  invalidDir,
				components: []string{"kube-apiserver"},
			},
			expectErrContains: "error writing",
		},
	}

	// Override the createExporterFn to return a fake compfinder.
	createFinderFn = func(b *bpb.ClusterBundle) (compfinder, error) {
		return &fakeFinder{}, nil
	}

	for _, tc := range testcases {
		t.Run(tc.testName, func(t *testing.T) {
			ctx := context.Background()

			// Kinda yucky setup. Probably means it's not testing the right stuff.
			var pairs []*testutil.FilePair
			if tc.in == validBundle {
				pairs = append(pairs, &testutil.FilePair{tc.in, testutil.FakeBundle})

				for _, c := range tc.opts.components {
					// output file depends on output dir + input file.
					if c != invalidComponent && tc.opts.outputDir == validDir {
						pairs = append(pairs, &testutil.FilePair{
							filepath.Join(tc.opts.outputDir, c) + ".yaml", "",
						})
					}
				}
			}

			globalOpts := &cmdlib.GlobalOptions{
				BundleFile:   tc.in,
				InputFormat:  "yaml",
				OutputFormat: "yaml",
			}
			rw := testutil.NewFakeReaderWriterFromPairs(pairs...)

			err := run(ctx, tc.opts, rw, globalOpts)
			if (tc.expectErrContains != "" && err == nil) || (tc.expectErrContains == "" && err != nil) {
				t.Errorf("run(opts: %+v) returned err: %v, Want Err: %v", tc.opts, err, tc.expectErrContains)
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
