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
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	testutil "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/testutil"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/transformer"
)

const (
	invalidComponent = "not-a-component"
)

// Fake implementation of exporter for unit tests.
type fakeExporter struct{}

func (f *fakeExporter) Export(_ *bpb.ClusterBundle, compName string) (*transformer.ExportedComponent, error) {
	if compName == invalidComponent {
		return nil, fmt.Errorf("could not find cluster component named \"%v\"", compName)
	}
	return &transformer.ExportedComponent{Name: compName}, nil
}

func TestRunExport(t *testing.T) {
	validBundle := "/bundle.yaml"
	invalidBundle := "/invalid.yaml"
	validDir := "/valid/dir"
	invalidDir := "/invalid/dir"

	var testcases = []struct {
		testName          string
		opts              *options
		expectErrContains string
	}{
		{
			testName: "success case",
			opts: &options{
				bundlePath: validBundle,
				outputDir:  validDir,
				components: []string{"kube-apiserver", "kube-scheduler"},
			},
		},
		{
			testName: "bundle read error",
			opts: &options{
				bundlePath: invalidBundle,
				outputDir:  validDir,
			},
			expectErrContains: "error reading",
		},
		{
			testName: "extract component error",
			opts: &options{
				bundlePath: validBundle,
				outputDir:  validDir,
				components: []string{invalidComponent},
			},
			expectErrContains: invalidComponent,
		},
		{
			testName: "component write error",
			opts: &options{
				bundlePath: validBundle,
				outputDir:  invalidDir,
				components: []string{"kube-apiserver"},
			},
			expectErrContains: "error writing",
		},
	}

	// Override the createExporterFn to return a fake exporter.
	createExporterFn = func(b *bpb.ClusterBundle) (exporter, error) {
		return &fakeExporter{}, nil
	}

	for _, tc := range testcases {
		t.Run(tc.testName, func(t *testing.T) {
			ctx := context.Background()

			// Kinda yucky setup. Probably means it's not testing the right stuff.
			var pairs []*testutil.FilePair
			if tc.opts.bundlePath == validBundle {
				pairs = append(pairs, &testutil.FilePair{tc.opts.bundlePath, testutil.FakeBundle})

				for _, c := range tc.opts.components {
					// output file depends on output dir + input file.
					if c != invalidComponent && tc.opts.outputDir == validDir {
						pairs = append(pairs, &testutil.FilePair{
							filepath.Join(tc.opts.outputDir, c) + ".yaml", "",
						})
					}
				}
			}
			rw := testutil.NewFakeReaderWriterFromPairs(pairs...)

			err := run(ctx, tc.opts, rw)
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
