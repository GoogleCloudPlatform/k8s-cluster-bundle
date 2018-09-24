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

package patch

import (
	"context"
	"errors"
	"testing"

	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/cmdlib"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/testutil"
)

const optionsCR = `
apiVersion: bundles/v1alpha1
kind: BundleOptions
foo: Bar
`

// Fake implementation of componentFinder for unit tests.
type fakeFinder struct {
	validComp string
}

func (f *fakeFinder) ClusterComponent(name string) *bpb.ClusterComponent {
	if name == f.validComp {
		return &bpb.ClusterComponent{}
	}
	return nil
}

// Fake implementation of patcher for unit tests.
type fakePatcher struct {
	throwsErr bool
}

func (f *fakePatcher) PatchBundle([]map[string]interface{}) (*bpb.ClusterBundle, error) {
	if f.throwsErr {
		return nil, errors.New("error patching bundle")
	}
	return &bpb.ClusterBundle{}, nil
}

func (f *fakePatcher) PatchComponent(*bpb.ClusterComponent, []map[string]interface{}) (*bpb.ClusterComponent, error) {
	if f.throwsErr {
		return nil, errors.New("error patching component")
	}
	return &bpb.ClusterComponent{}, nil
}

func TestRunPatchBundle(t *testing.T) {
	validInBundleFile := "in/path/to/bundle.yaml"
	validOutBundleFile := "out/path/to/bundle.yaml"
	validOptionsFile := "/path/to/options.yaml"
	invalidFile := "/invalid.yaml"

	testCases := []struct {
		testName string
		in       string
		out      string
		opts     *options
		patcher  *fakePatcher
		wantErr  bool
	}{
		{
			testName: "success case",
			in:       validInBundleFile,
			out:      validOutBundleFile,
			opts: &options{
				optionsCRs: []string{validOptionsFile},
			},
			patcher: &fakePatcher{throwsErr: false},
			wantErr: false,
		},
		{
			testName: "bundle read error",
			in:       invalidFile,
			out:      validOutBundleFile,
			opts: &options{
				optionsCRs: []string{validOptionsFile},
			},
			patcher: &fakePatcher{throwsErr: false},
			wantErr: true,
		},
		{
			testName: "options read error",
			in:       validInBundleFile,
			out:      validOutBundleFile,
			opts: &options{
				optionsCRs: []string{validOptionsFile, invalidFile},
			},
			patcher: &fakePatcher{throwsErr: false},
			wantErr: true,
		},
		{
			testName: "bundle patching error",
			in:       validInBundleFile,
			out:      validOutBundleFile,
			opts: &options{
				optionsCRs: []string{validOptionsFile},
			},
			patcher: &fakePatcher{throwsErr: true},
			wantErr: true,
		},
		{
			testName: "bundle write error",
			in:       validInBundleFile,
			out:      invalidFile,
			opts: &options{
				optionsCRs: []string{validOptionsFile},
			},
			patcher: &fakePatcher{throwsErr: false},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			ctx := context.Background()
			var pairs []*testutil.FilePair
			if tc.in == validInBundleFile {
				pairs = append(pairs, &testutil.FilePair{tc.in, testutil.FakeBundle})
			}
			if tc.out == validOutBundleFile {
				pairs = append(pairs, &testutil.FilePair{tc.out, testutil.FakeBundle})
			}
			for _, o := range tc.opts.optionsCRs {
				if o == validOptionsFile {
					pairs = append(pairs, &testutil.FilePair{o, testutil.FakeBundle})
				}
			}
			globalOpts := &cmdlib.GlobalOptions{
				BundleFile:   tc.in,
				OutputFile:   tc.out,
				InputFormat:  "yaml",
				OutputFormat: "yaml",
			}

			rw := testutil.NewFakeReaderWriterFromPairs(pairs...)
			// Override the createPatcherFn to return a fakePatcher.
			createPatcherFn = func(b *bpb.ClusterBundle) (patcher, error) {
				return tc.patcher, nil
			}
			err := runPatchBundle(ctx, tc.opts, rw, globalOpts)
			if !tc.wantErr && err != nil {
				t.Errorf("runPatchBundle(opts: %+v) = error %v, want no error", tc.opts, err)
			}
			if tc.wantErr && err == nil {
				t.Errorf("runPatchBundle(opts: %+v) = no error, want error", tc.opts)
			}
		})
	}
}

func TestRunPatchComponent(t *testing.T) {
	validBundleFile := "/path/to/bundle.yaml"
	validOptionsFile := "/path/to/options.yaml"
	validComponent := "valid-component"
	validOutFile := "/path/to/patched.yaml"
	invalidFile := "/invalid.yaml"

	testCases := []struct {
		testName string
		in       string
		out      string
		opts     *options
		patcher  *fakePatcher
		wantErr  bool
	}{
		{
			testName: "success case",
			in:       validBundleFile,
			out:      validOutFile,
			opts: &options{
				optionsCRs: []string{validOptionsFile},
				component:  validComponent,
			},
			patcher: &fakePatcher{throwsErr: false},
			wantErr: false,
		},
		{
			testName: "bundle read error",
			in:       invalidFile,
			out:      validOutFile,
			opts: &options{
				optionsCRs: []string{validOptionsFile},
				component:  validComponent,
			},
			patcher: &fakePatcher{throwsErr: false},
			wantErr: true,
		},
		{
			testName: "options read error",
			in:       validBundleFile,
			out:      validOutFile,
			opts: &options{
				optionsCRs: []string{validOptionsFile, invalidFile},
				component:  validComponent,
			},
			patcher: &fakePatcher{throwsErr: false},
			wantErr: true,
		},
		{
			testName: "component not found error",
			in:       validBundleFile,
			out:      validOutFile,
			opts: &options{
				optionsCRs: []string{validOptionsFile},
				component:  "invalid-component",
			},
			patcher: &fakePatcher{throwsErr: false},
			wantErr: true,
		},
		{
			testName: "component patching error",
			in:       validBundleFile,
			out:      validOutFile,
			opts: &options{
				optionsCRs: []string{validOptionsFile},
				component:  validComponent,
			},
			patcher: &fakePatcher{throwsErr: true},
			wantErr: true,
		},
		{
			testName: "component write error",
			in:       validBundleFile,
			out:      invalidFile,
			opts: &options{
				optionsCRs: []string{validOptionsFile},
				component:  validComponent,
			},
			patcher: &fakePatcher{throwsErr: false},
			wantErr: true,
		},
	}

	// Override the createFinderFn to return a fakeFinder.
	createFinderFn = func(*bpb.ClusterBundle) (componentFinder, error) {
		return &fakeFinder{validComp: validComponent}, nil
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			ctx := context.Background()
			var pairs []*testutil.FilePair
			if tc.in == validBundleFile {
				pairs = append(pairs, &testutil.FilePair{tc.in, testutil.FakeBundle})
			}
			if tc.out == validOutFile {
				pairs = append(pairs, &testutil.FilePair{tc.out, testutil.FakeBundle})
			}
			for _, o := range tc.opts.optionsCRs {
				if o == validOptionsFile {
					pairs = append(pairs, &testutil.FilePair{o, testutil.FakeBundle})
				}
			}
			globalOpts := &cmdlib.GlobalOptions{
				BundleFile:   tc.in,
				OutputFile:   tc.out,
				InputFormat:  "yaml",
				OutputFormat: "yaml",
			}

			rw := testutil.NewFakeReaderWriterFromPairs(pairs...)

			// Override the createPatcherFn to return a fakePatcher.
			createPatcherFn = func(*bpb.ClusterBundle) (patcher, error) {
				return tc.patcher, nil
			}
			err := runPatchComponent(ctx, tc.opts, rw, globalOpts)
			if !tc.wantErr && err != nil {
				t.Errorf("runPatchComponent(opts: %+v) = error %v, want no error", tc.opts, err)
			}
			if tc.wantErr && err == nil {
				t.Errorf("runPatchComponent(opts: %+v) = no error, want error", tc.opts)
			}
		})
	}
}
