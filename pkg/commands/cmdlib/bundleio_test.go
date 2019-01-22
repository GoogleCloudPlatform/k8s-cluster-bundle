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
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/wrapper"
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
	bundleIn  *bundle.BundleBuilder
	bundleOut string

	componentIn  *bundle.ComponentBuilder
	componentOut string
	err          error
}

func (f *fakeInliner) BundleFiles(_ context.Context, b *bundle.BundleBuilder) (*bundle.Bundle, error) {
	f.bundleIn = b
	o, err := converter.FromYAMLString(f.bundleOut).ToBundle()
	if err != nil {
		return nil, err
	}
	return o, f.err
}

func (f *fakeInliner) ComponentFiles(_ context.Context, c *bundle.ComponentBuilder) (*bundle.Component, error) {
	f.componentIn = c
	o, err := converter.FromYAMLString(f.componentOut).ToComponent()
	if err != nil {
		return nil, err
	}
	return o, f.err
}

var bundleEx = `
apiVersion: bundle.gke.io/v1alpha1
kind: Bundle
components:
- apiVersion: bundle.gke.io/v1alpha1
  kind: Component
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

var bundleBuilderEx = `
apiVersion: bundle.gke.io/v1alpha1
kind: BundleBuilder
componentFiles:
- url: /some/inlined/component.yaml`

var componentEx = `
apiVersion: bundle.gke.io/v1alpha1
kind: Component
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

var componentBuilderEx = `
apiVersion: bundle.gke.io/v1alpha1
kind: ComponentBuilder
metadata:
  name: test-pkg
componentName: test-comp
version: 0.1.0
objectFiles:
- url: "/foo/bar/biff.yaml"`

func TestReadBundleData(t *testing.T) {
	testcases := []struct {
		desc      string
		opts      *GlobalOptions
		readFile  string
		readStdin string

		inlineBundleOut string
		inlineCompOut   string
		inlineErr       error

		expDataSubstr string
		expErrSubstr  string
	}{
		// Basic success for four base types.
		{
			desc: "Test success file read: bundle",
			opts: &GlobalOptions{
				InputFile: "/foo/bar/biff.yaml",
			},
			readFile: bundleEx,
		},
		{
			desc: "Test success file read: component",
			opts: &GlobalOptions{
				InputFile: "/foo/bar/biff.yaml",
			},
			readFile: componentEx,
		},
		{
			desc: "Test success file read: bundle builder",
			opts: &GlobalOptions{
				InputFile: "/foo/bar/biff.yaml",
			},
			readFile: bundleBuilderEx,
		},
		{
			desc: "Test success file read: component builder",
			opts: &GlobalOptions{
				InputFile: "/foo/bar/biff.yaml",
			},
			readFile: componentBuilderEx,
		},

		{
			desc: "Test success stdin read: bundle",
			opts: &GlobalOptions{
				InputFormat: "yaml",
			},
			readStdin: bundleEx,
		},
		{
			desc: "Test success stdin read: component",
			opts: &GlobalOptions{
				InputFormat: "yaml",
			},
			readStdin: componentEx,
		},
		{
			desc: "Test success file read + inline comp: bundle",
			opts: &GlobalOptions{
				InputFile: "/foo/bar/biff.yaml",
				Inline:    true,
			},
			readFile:        bundleBuilderEx,
			inlineBundleOut: bundleEx,
			expDataSubstr:   "test-pkg",
		},
		{
			desc: "Test success file read + inline obj: bundle",
			opts: &GlobalOptions{
				InputFile: "/foo/bar/biff.yaml",
				Inline:    true,
			},
			readFile:        bundleBuilderEx,
			inlineBundleOut: bundleEx,
			expDataSubstr:   "some-pod",
		},
		{
			desc: "Test success component file read + inline obj: component",
			opts: &GlobalOptions{
				InputFile: "/foo/bar/biff.yaml",
				Inline:    true,
			},
			readFile:      componentBuilderEx,
			inlineCompOut: componentEx,
			expDataSubstr: "some-pod",
		},
		{
			desc:          "Succces: default content type to yaml",
			opts:          &GlobalOptions{},
			readStdin:     bundleEx,
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
				InputFile: "foo/bar/biff.yaml",
				Inline:    true,
			},
			readFile:      componentBuilderEx,
			inlineCompOut: componentEx,
			inlineErr:     errors.New("zork"),
			expErrSubstr:  "error inlining objects",
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
						bundleOut:    tc.inlineBundleOut,
						componentOut: tc.inlineCompOut,
						err:          tc.inlineErr,
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

			obj := data.Object()
			if obj == nil {
				t.Fatalf("expected wrapped bundle object to be non-nil")
			}
			outyaml, err := converter.FromObject(obj).ToYAML()
			if err != nil {
				t.Fatalf("error converting back to yaml: %v", err)
			}
			if tc.expDataSubstr != "" && !strings.Contains(string(outyaml), tc.expDataSubstr) {
				t.Errorf("got data\n%s but expected it to contain %q", string(outyaml), tc.expDataSubstr)
			}
		})
	}
}

func TestWriteBundleData(t *testing.T) {
	testcases := []struct {
		desc           string
		opts           *GlobalOptions
		content        string
		expStdioSubstr string
		expErrSubstr   string
	}{
		{
			desc:           "Test success stdout write: bundle",
			opts:           &GlobalOptions{},
			content:        bundleEx,
			expStdioSubstr: "test-pkg",
		},
		{
			desc:           "Test success stdout write: bundle builder",
			opts:           &GlobalOptions{},
			content:        bundleBuilderEx,
			expStdioSubstr: "inlined/component",
		},
		{
			desc: "Test success stdout write: component",
			opts: &GlobalOptions{
				OutputFormat: "yaml",
			},
			content:        componentEx,
			expStdioSubstr: "some-pod",
		},
		{
			desc:           "success stdout write: component builder",
			content:        componentEx,
			opts:           &GlobalOptions{},
			expStdioSubstr: "test-comp",
		},
		{
			desc: "Test success stdout write json: component",
			opts: &GlobalOptions{
				OutputFormat: "json",
			},
			content:        componentEx,
			expStdioSubstr: "some-pod",
		},
		{
			desc:           "success: content format defaults to yaml",
			content:        componentEx,
			opts:           &GlobalOptions{},
			expStdioSubstr: "some-pod",
		},

		// Errors
		{
			desc:         "error: wrapper nil",
			opts:         &GlobalOptions{},
			expErrSubstr: "content was empty",
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

			bwrap, err := wrapper.FromRaw("yaml", []byte(tc.content))
			// error checked below.
			if err == nil {
				err = brw.WriteBundleData(ctx, bwrap, tc.opts)
			}

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
