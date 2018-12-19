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

package cmdlib

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/files"
)

type fakeStdioRW struct {
	rdBytes []byte
	rdErr   error

	wrBytes []byte
	wrErr   error
}

func (f *fakeStdioRW) ReadAll() ([]byte, error) {
	return f.rdBytes, f.rdErr
}

func (f *fakeStdioRW) Write(b []byte) (int, error) {
	f.wrBytes = b
	return len(f.wrBytes), f.wrErr
}

type fakeFileRW struct {
	rdPath  string
	rdBytes []byte
	rdErr   error

	wrPath  string
	wrBytes []byte
	wrError error
}

func (rw *fakeFileRW) ReadFile(_ context.Context, path string) ([]byte, error) {
	rw.rdPath = path
	return rw.rdBytes, rw.rdErr
}

func (rw *fakeFileRW) WriteFile(_ context.Context, path string, bytes []byte, permissions os.FileMode) error {
	rw.wrPath = path
	rw.wrBytes = bytes
	return rw.wrError
}

type fakeInliner struct {
	bundleCompIn  *bundle.Bundle
	bundleCompOut string

	bundleObjIn  *bundle.Bundle
	bundleObjOut string

	componentIn  *bundle.ComponentPackage
	componentOut string
	err          error
}

func (f *fakeInliner) InlineBundleFiles(_ context.Context, b *bundle.Bundle) (*bundle.Bundle, error) {
	f.bundleCompIn = b
	o, err := converter.FromYAMLString(f.bundleCompOut).ToBundle()
	if err != nil {
		return nil, err
	}
	return o, f.err
}

func (f *fakeInliner) InlineComponentsInBundle(_ context.Context, b *bundle.Bundle) (*bundle.Bundle, error) {
	f.bundleObjIn = b
	o, err := converter.FromYAMLString(f.bundleObjOut).ToBundle()
	if err != nil {
		return nil, err
	}
	return o, f.err
}

func (f *fakeInliner) InlineComponent(_ context.Context, c *bundle.ComponentPackage) (*bundle.ComponentPackage, error) {
	f.componentIn = c
	o, err := converter.FromYAMLString(f.componentOut).ToComponentPackage()
	if err != nil {
		return nil, err
	}
	return o, f.err
}

var successBundle = `
apiVersion: bundle.gke.io/v1alpha1
kind: Bundle
components:
- apiVersion: bundle.gke.io/v1alpha1
  kind: ComponentPackage
  metadata:
    name: test-pkg
  spec:
    componentName: test-comp
    version: 0.1.0
    objectFiles:
    - url: zip/bar/biff.yaml`

var notInlineBundle = `
apiVersion: bundle.gke.io/v1alpha1
kind: Bundle
componentFiles:
- url: some/inline/component.yaml`

var successComponent = `
apiVersion: bundle.gke.io/v1alpha1
kind: ComponentPackage
metadata:
  name: test-pkg
spec:
  componentName: test-comp
  version: 0.1.0
  objectFiles:
  - url: zip/bar/biff.yaml`

var inlinedComponent = `
apiVersion: bundle.gke.io/v1alpha1
kind: ComponentPackage
metadata:
  name: test-pkg
spec:
  componentName: test-comp
  version: 0.1.0
  objects:
  - apiVersion: v1
    kind: Pod
    metadata:
      name: some-pod`

var fullyInlinedBundle = `
apiVersion: bundle.gke.io/v1alpha1
kind: Bundle
components:
- apiVersion: bundle.gke.io/v1alpha1
  kind: ComponentPackage
  metadata:
    name: test-pkg
  spec:
    componentName: test-comp
    version: 0.1.0
    objects:
    - apiVersion: v1
      kind: Pod
      metadata:
        name: some-pod`

func TestReadBundleData(t *testing.T) {
	testcases := []struct {
		desc      string
		opts      *GlobalOptions
		readFile  string
		readStdin string

		inlineBundleCompOut string
		inlineBundleObjOut  string
		inlineCompObjOut    string
		inlineErr           error

		expDataSubstr string
		expErrSubstr  string
	}{
		{
			desc: "Test success file read: bundle",
			opts: &GlobalOptions{
				InputFile: "foo/bar/biff.yaml",
			},
			readFile: successBundle,
		},
		{
			desc: "Test success file read: component",
			opts: &GlobalOptions{
				InputFile: "foo/bar/biff.yaml",
			},
			readFile: successComponent,
		},
		{
			desc: "Test success stdin read: bundle",
			opts: &GlobalOptions{
				InputFormat: "yaml",
			},
			readStdin: successBundle,
		},
		{
			desc: "Test success stdin read: component",
			opts: &GlobalOptions{
				InputFormat: "yaml",
			},
			readStdin: successComponent,
		},
		{
			desc: "Test success file read + inline comp: bundle",
			opts: &GlobalOptions{
				InputFile:        "foo/bar/biff.yaml",
				InlineComponents: true,
			},
			readFile:            notInlineBundle,
			inlineBundleCompOut: successBundle,
			expDataSubstr:       "test-pkg",
		},
		{
			desc: "Test success file read + inline obj: bundle",
			opts: &GlobalOptions{
				InputFile:        "foo/bar/biff.yaml",
				InlineComponents: true,
				InlineObjects:    true,
			},
			readFile:            notInlineBundle,
			inlineBundleCompOut: successBundle,
			inlineBundleObjOut:  fullyInlinedBundle,
			expDataSubstr:       "some-pod",
		},
		{
			desc: "Test success component file read + inline obj: component",
			opts: &GlobalOptions{
				InputFile:     "foo/bar/biff.yaml",
				InlineObjects: true,
			},
			readFile:         successComponent,
			inlineCompObjOut: inlinedComponent,
			expDataSubstr:    "some-pod",
		},
		{
			desc:          "Succces: default content type to yaml",
			opts:          &GlobalOptions{},
			readStdin:     successBundle,
			expDataSubstr: "test-pkg",
		},

		// Some error cases
		{
			desc: "Error: bad kind",
			opts: &GlobalOptions{
				InputFormat: "yaml",
			},
			readStdin: `
apiVersion: bundle.gke.io/v1alpha1
kind: Foobar
spec:
  componentName: test-comp
  version: 0.1.0`,
			expErrSubstr: "unrecognized bundle-kind",
		},
		{
			desc: "Error: inlining",
			opts: &GlobalOptions{
				InputFile:     "foo/bar/biff.yaml",
				InlineObjects: true,
			},
			readFile:         successComponent,
			inlineCompObjOut: inlinedComponent,
			inlineErr:        errors.New("zork!"),
			expErrSubstr:     "error inlining objects",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			ctx := context.Background()

			brw := &realBundleReaderWriter{
				rw: &fakeFileRW{
					rdBytes: []byte(tc.readFile),
				},
				stdio: &fakeStdioRW{
					rdBytes: []byte(tc.readStdin),
				},
				makeInlinerFn: func(rw files.FileReaderWriter, inputFile string) fileInliner {
					return &fakeInliner{
						bundleCompOut: tc.inlineBundleCompOut,
						bundleObjOut:  tc.inlineBundleObjOut,
						componentOut:  tc.inlineCompObjOut,
						err:           tc.inlineErr,
					}
				},
			}

			data, err := brw.ReadBundleData(ctx, tc.opts)
			if err != nil && tc.expErrSubstr == "" {
				t.Fatalf("Got error %q but expected no error", err.Error())
			} else if err == nil && tc.expErrSubstr != "" {
				t.Fatalf("Got no error but expected error containing %q", tc.expErrSubstr)
			} else if err != nil && !strings.Contains(err.Error(), tc.expErrSubstr) {
				t.Fatalf("Got error %q but expected it to contain %q", err.Error(), tc.expErrSubstr)
			} else if err != nil {
				// Even though this is a success case, we need to return because we
				// can't validate any properties of the data
				return
			}

			if data == nil {
				t.Fatalf("Got nil data, but expected some data")
			}
			if data.Bundle == nil && data.Component == nil {
				t.Fatalf("expected one of bundle or component to be non-nil")
			}
			outyaml, err := converter.FromObject(data).ToYAML()
			if err != nil {
				t.Fatalf("error converting back to yaml: %v", err)
			}
			if tc.expDataSubstr != "" && !strings.Contains(string(outyaml), tc.expDataSubstr) {
				t.Errorf("got data %s but expected it to contain %q", string(outyaml), tc.expDataSubstr)
			}
		})
	}
}

func TestWriteBundleData(t *testing.T) {
	testcases := []struct {
		desc      string
		opts      *GlobalOptions
		bundle    string
		component string

		expStdioSubstr string

		expErrSubstr string
	}{
		{
			desc:           "Test success stdout write: bundle",
			opts:           &GlobalOptions{},
			bundle:         successBundle,
			expStdioSubstr: "test-pkg",
		},
		{
			desc: "Test success stdout write: component",
			opts: &GlobalOptions{
				OutputFormat: "yaml",
			},
			component:      inlinedComponent,
			expStdioSubstr: "some-pod",
		},
		{
			desc: "Test success stdout write json: component",
			opts: &GlobalOptions{
				OutputFormat: "json",
			},
			component:      inlinedComponent,
			expStdioSubstr: "some-pod",
		},
		{
			desc:           "success: content format defaults to yaml",
			component:      inlinedComponent,
			opts:           &GlobalOptions{},
			expStdioSubstr: "some-pod",
		},

		// Errors
		{
			desc:         "error: both nil",
			opts:         &GlobalOptions{},
			expErrSubstr: "both the bundle and the component fields were nil",
		},
		{
			desc:         "error: both nil",
			opts:         &GlobalOptions{},
			bundle:       successBundle,
			component:    inlinedComponent,
			expErrSubstr: "both the bundle and the component fields were non-nil",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			ctx := context.Background()

			fileRW := &fakeFileRW{}
			stdioRW := &fakeStdioRW{}

			brw := &realBundleReaderWriter{
				rw:    fileRW,
				stdio: stdioRW,
				makeInlinerFn: func(rw files.FileReaderWriter, inputFile string) fileInliner {
					return &fakeInliner{}
				},
			}

			bwrap := &BundleWrapper{}
			if tc.bundle != "" {
				b, err := converter.FromYAMLString(tc.bundle).ToBundle()
				if err != nil {
					t.Fatalf("Error converting bundle to YAML: %v", err)
				}
				bwrap.Bundle = b
			}
			if tc.component != "" {
				c, err := converter.FromYAMLString(tc.component).ToComponentPackage()
				if err != nil {
					t.Fatalf("Error converting bundle to YAML: %v", err)
				}
				bwrap.Component = c
			}

			err := brw.WriteBundleData(ctx, bwrap, tc.opts)
			if err != nil && tc.expErrSubstr == "" {
				t.Fatalf("Got error %q but expected no error", err.Error())
			} else if err == nil && tc.expErrSubstr != "" {
				t.Fatalf("Got no error but expected error containing %q", tc.expErrSubstr)
			} else if err != nil && !strings.Contains(err.Error(), tc.expErrSubstr) {
				t.Fatalf("Got error %q but expected it to contain %q", err.Error(), tc.expErrSubstr)
			} else if err != nil {
				// Even though this is a success case, we need to return because we
				// can't validate any properties of the data
				return
			}

			writtenStdio := string(stdioRW.wrBytes)

			if tc.expStdioSubstr != "" && !strings.Contains(writtenStdio, tc.expStdioSubstr) {
				t.Errorf("got stdout content %s, but expected it to contain %q", writtenStdio, tc.expStdioSubstr)
			}
		})
	}
}
