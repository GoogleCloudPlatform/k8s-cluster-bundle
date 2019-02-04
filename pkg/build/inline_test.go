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

package build

import (
	"context"
	"fmt"
	"strings"
	"testing"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/files"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/testutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type fakeLocalReader struct {
	files map[string][]byte
}

func (f *fakeLocalReader) ReadFileObj(_ context.Context, file bundle.File) ([]byte, error) {
	url := file.URL
	if strings.HasPrefix(url, "file://") {
		url = strings.TrimPrefix(url, "file://")
	}
	fi, ok := f.files[url]
	if !ok {
		return nil, fmt.Errorf("unexpected file path %q", file.URL)
	}
	return fi, nil
}

const defaultBundle = `
kind: BundleBuilder
setName: foo-bundle
version: 1.2.3
componentFiles:
- url: /path/to/apiserver-component.yaml`

var kubeApiserverComponent = []byte(`
kind: ComponentBuilder
componentName: kube-apiserver
version: 1.2.3
objectFiles:
- url: '/path/to/kube_apiserver.yaml'`)

var kubeApiserver = []byte(`
apiVersion: v1
kind: Zork
metadata:
  name: biffbam
biff: bam`)

var defaultFiles = map[string][]byte{
	"/path/to/apiserver-component.yaml": kubeApiserverComponent,
	"/path/to/kube_apiserver.yaml":      kubeApiserver,
}

type bundleRef struct {
	setName string
	version string
}

type objCheck struct {
	name     string
	annotKey string
	annotVal string

	// For if the it's imported as raw text
	cfgMap map[string]string
}

type compRef struct {
	name string
	ref  bundle.ComponentReference
	obj  []objCheck
}

func TestInlineBundleFiles(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		desc         string
		data         string
		files        map[string][]byte
		expErrSubstr string
		expBun       bundleRef
		expComps     []compRef
	}{
		{
			desc:   "success: inline basic bundle",
			data:   defaultBundle,
			files:  defaultFiles,
			expBun: bundleRef{setName: "foo-bundle", version: "1.2.3"},
			expComps: []compRef{
				{
					name: "kube-apiserver-1-2-3",
					ref: bundle.ComponentReference{
						ComponentName: "kube-apiserver",
						Version:       "1.2.3",
					},
					obj: []objCheck{
						{
							name:     "biffbam",
							annotKey: string(bundle.InlineTypeIdentifier),
							annotVal: string(bundle.KubeObjectInline),
						},
					},
				},
			},
		},

		{
			desc: "success: inline basic bundle + file prefix",
			data: `
kind: BundleBuilder
setName: foo-bundle
version: 1.2.3
componentFiles:
- url: file:///path/to/apiserver-component.yaml`,
			files:  defaultFiles,
			expBun: bundleRef{setName: "foo-bundle", version: "1.2.3"},
			expComps: []compRef{
				{
					name: "kube-apiserver-1-2-3",
					ref: bundle.ComponentReference{
						ComponentName: "kube-apiserver",
						Version:       "1.2.3",
					},
					obj: []objCheck{
						{
							name:     "biffbam",
							annotKey: string(bundle.InlineTypeIdentifier),
							annotVal: string(bundle.KubeObjectInline),
						},
					},
				},
			},
		},

		{
			desc: "success: inline bundle with raw text",
			data: `
kind: BundleBuilder
setName: foo-bundle
version: 1.2.3
componentFiles:
- url: /path/to/raw-text-component.yaml`,
			files: map[string][]byte{
				"/path/to/raw-text.yaml":   []byte("foobar"),
				"/path/to/rawer-text.yaml": []byte("boobar"),
				"/path/to/raw-text-component.yaml": []byte(`
kind: ComponentBuilder
componentName: kube-apiserver
version: 1.2.3
rawTextFiles:
- name: some-raw-text
  files:
  - url: '/path/to/raw-text.yaml'
  - url: '/path/to/rawer-text.yaml'`),
			},
			expBun: bundleRef{setName: "foo-bundle", version: "1.2.3"},
			expComps: []compRef{
				{
					name: "kube-apiserver-1-2-3",
					ref: bundle.ComponentReference{
						ComponentName: "kube-apiserver",
						Version:       "1.2.3",
					},
					obj: []objCheck{
						{
							name:     "some-raw-text",
							annotKey: string(bundle.InlineTypeIdentifier),
							annotVal: string(bundle.RawStringInline),
							cfgMap: map[string]string{
								"raw-text.yaml":   "foobar",
								"rawer-text.yaml": "boobar",
							},
						},
					},
				},
			},
		},

		{
			desc: "success: inline bundle with multi-doc",
			data: `
kind: BundleBuilder
setName: multi-bundle
version: 2.2.3
componentFiles:
- url: /path/to/multi-doc-component.yaml`,
			files: map[string][]byte{
				"/path/to/multi-doc-component.yaml": []byte(`kind: ComponentBuilder
componentName: kube-multi
version: 2.3.4
objectFiles:
- url: '/path/to/multi-doc-obj.yaml'`),
				"/path/to/multi-doc-obj.yaml": []byte(`
apiVersion: v1
kind: Zork
metadata:
  name: foobar
foo: bar
---
apiVersion: v1
kind: Zork
metadata:
  name: biffbam
biff: bam`),
			},
			expBun: bundleRef{setName: "multi-bundle", version: "2.2.3"},
			expComps: []compRef{
				{
					name: "kube-multi-2-3-4",
					ref: bundle.ComponentReference{
						ComponentName: "kube-multi",
						Version:       "2.3.4",
					},
					obj: []objCheck{
						{
							name:     "foobar",
							annotKey: string(bundle.InlineTypeIdentifier),
							annotVal: string(bundle.KubeObjectInline),
						},
						{
							name:     "biffbam",
							annotKey: string(bundle.InlineTypeIdentifier),
							annotVal: string(bundle.KubeObjectInline),
						},
					},
				},
			},
		},

		{
			desc: "success: inline bundle with component",
			data: `
kind: BundleBuilder
setName: component-bundle
version: 2.2.4
componentFiles:
- url: /path/to/foo-component.yaml`,
			files: map[string][]byte{
				"/path/to/foo-component.yaml": []byte(`kind: Component
spec:
  componentName: kube-inline
  version: 2.3.4
  objects:
  - apiVersion: v1
    kind: pod
    metadata:
      name: foo-comp`),
			},
			expBun: bundleRef{setName: "component-bundle", version: "2.2.4"},
			expComps: []compRef{
				{
					ref: bundle.ComponentReference{
						ComponentName: "kube-inline",
						Version:       "2.3.4",
					},
					obj: []objCheck{
						{
							name: "foo-comp",
						},
					},
				},
			},
		},

		// Error cases
		{
			desc: "error: bad component type",
			data: defaultBundle,
			files: map[string][]byte{
				"/path/to/apiserver-component.yaml": []byte(`
kind: Zog
componentName: kube-apiserver
version: 1.2.3
objectFiles:
- url: '/path/to/kube_apiserver.yaml'`),
			},
			expErrSubstr: "unsupported kind",
		},
		{
			desc:         "fail: can't read file",
			data:         defaultBundle,
			files:        make(map[string][]byte),
			expErrSubstr: "error reading file",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			data, err := converter.FromYAMLString(tc.data).ToBundleBuilder()
			if err != nil {
				t.Fatalf("Error converting bundle: %v", err)
			}

			inliner := NewInlinerWithScheme(files.FileScheme, &fakeLocalReader{tc.files})
			got, err := inliner.BundleFiles(ctx, data)
			cerr := testutil.CheckErrorCases(err, tc.expErrSubstr)
			if cerr != nil {
				t.Fatal(cerr)
			}
			// It's possible we expect to have an error. But, if that's the case, we
			// can't continue.
			if err != nil {
				return
			}

			if got == nil {
				t.Fatalf("Expected data to not be nil")
			}

			validateBundle(t, got, tc.expBun)
			validateComponents(t, got.Components, tc.expComps)
		})
	}
}

// Some of the Component cases are tested via TestInlineBundleFiles above
// (which calls InlineComponentFiles).

func TestInlineComponentFiles(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		desc         string
		data         []byte
		files        map[string][]byte
		expErrSubstr string
		expComp      compRef
	}{
		{
			desc:  "success: inline basic component",
			data:  kubeApiserverComponent,
			files: defaultFiles,
			expComp: compRef{
				name: "kube-apiserver-1-2-3",
				ref: bundle.ComponentReference{
					ComponentName: "kube-apiserver",
					Version:       "1.2.3",
				},
				obj: []objCheck{
					{
						name:     "biffbam",
						annotKey: string(bundle.InlineTypeIdentifier),
						annotVal: string(bundle.KubeObjectInline),
					},
				},
			},
		},

		// Error cases.
		{
			desc:         "fail: can't read file",
			data:         kubeApiserverComponent,
			files:        make(map[string][]byte),
			expErrSubstr: "error reading file",
		},
		{
			desc: "fail: can't read raw text file",
			data: []byte(`
kind: ComponentBuilder
componentName: kube-apiserver
version: 1.2.3
rawTextFiles:
- name: foo-group
  files:
  - url: '/path/to/raw-text.yaml'
  - url: '/path/to/rawer-text.yaml'`),
			files:        make(map[string][]byte),
			expErrSubstr: "error reading raw text file for",
		},
		{
			desc: "fail: can't read raw text file group: no name",
			data: []byte(`
kind: ComponentBuilder
componentName: kube-apiserver
version: 1.2.3
rawTextFiles:
- files:
  - url: '/path/to/raw-text.yaml'
  - url: '/path/to/rawer-text.yaml'`),
			files:        make(map[string][]byte),
			expErrSubstr: "error reading raw text file group",
		},
		{
			desc: "fail: can't convert object to unstructured",
			data: kubeApiserverComponent,
			files: map[string][]byte{
				"/path/to/kube_apiserver.yaml": []byte("blah"),
			},
			expErrSubstr: "error converting object to unstructured",
		},
		{
			desc: "fail: can't converting multi-doc object to unstructured",
			data: kubeApiserverComponent,
			files: map[string][]byte{
				"/path/to/kube_apiserver.yaml": []byte(`
blah
---
blar
`),
			},
			expErrSubstr: "error converting multi-doc object",
		},
		{
			desc: "error: invalid specified name",
			data: []byte(`
kind: ComponentBuilder
metadata:
  name: this-
componentName: kube-apiserver
version: 1.2.3
objectFiles:
- url: '/path/to/kube_apiserver.yaml'`),
			files:        defaultFiles,
			expErrSubstr: "DNS-1123",
		},
		{
			desc: "error: invalid generated name",
			data: []byte(`
kind: ComponentBuilder
metadata:
componentName: kube-apiserver
version: 1.2.3-
objectFiles:
- url: '/path/to/kube_apiserver.yaml'`),
			files:        defaultFiles,
			expErrSubstr: "DNS-1123",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			data, err := converter.FromYAML(tc.data).ToComponentBuilder()
			if err != nil {
				t.Fatalf("Error converting component: %v", err)
			}

			inliner := NewInlinerWithScheme(files.FileScheme, &fakeLocalReader{tc.files})
			got, err := inliner.ComponentFiles(ctx, data)
			cerr := testutil.CheckErrorCases(err, tc.expErrSubstr)
			if cerr != nil {
				t.Fatal(cerr)
			}
			if err != nil {
				return
			}

			if got == nil {
				t.Fatalf("Expected data to not be nil")
			}

			validateComponents(t, []*bundle.Component{got}, []compRef{tc.expComp})
		})
	}
}

func validateBundle(t *testing.T, got *bundle.Bundle, expBun bundleRef) {
	strbun, err := converter.FromObject(got).ToYAML()
	if err != nil {
		t.Fatalf("Error converting bundle back to YAML: %v", err)
	}
	if got.Kind != "Bundle" {
		t.Errorf("After inlining bundle, got kind %q, but expected \"Bundle\". Output\n%s", got.Kind, strbun)
	}
	if expBun.setName != "" && expBun.setName != got.SetName {
		t.Errorf("After inlining bundle, got SetName %q, expected %q. Output\n%s", got.SetName, expBun.setName, strbun)
	}
	if expBun.version != "" && expBun.version != got.Version {
		t.Errorf("After inlining bundle, got Version %q, expected %q. Output\n%s", got.Version, expBun.setName, strbun)
	}
}

func validateComponents(t *testing.T, comp []*bundle.Component, expComps []compRef) {
	gotCompRefSet := make(map[bundle.ComponentReference]*bundle.Component)
	var gotCompRefs []bundle.ComponentReference
	for _, c := range comp {
		gotCompRefSet[c.ComponentReference()] = c
		gotCompRefs = append(gotCompRefs, c.ComponentReference())
	}

	// Compare the component contents
	for _, ec := range expComps {
		ref := ec.ref

		comp, ok := gotCompRefSet[ref]
		if !ok {
			t.Errorf("got components %v, it did not contain expected component %v", gotCompRefs, ref)
		}
		if comp == nil {
			t.Fatalf("got nil component for ref: %v", ref)
		}

		if comp.GetName() != ec.name {
			t.Errorf("got component meta.name %q, but expected %q", comp.GetName(), ec.name)
		}

		// Compare the object data
		gotObjs := make(map[string]*unstructured.Unstructured)
		for _, obj := range comp.Spec.Objects {
			gotObjs[obj.GetName()] = obj
		}
		for _, o := range ec.obj {
			obj := gotObjs[o.name]
			if obj == nil {
				t.Errorf("got objs %v, but object with name name %q.", gotObjs, o.name)
			}
			an := obj.GetAnnotations()
			if an[o.annotKey] != o.annotVal {
				t.Errorf("for obj %q, got annotation %q for key %q, but expected %q", o.name, an[o.annotKey], o.annotKey, o.annotVal)
			}

			// Check the config map contents.
			for expkey, expval := range o.cfgMap {
				dataObj := obj.Object["data"]
				dataMap, ok := dataObj.(map[string]interface{})
				if !ok {
					t.Fatalf("Expected data to be a map of string to interface for comp %v in object %q", ref, obj.GetName())
				}
				val, ok := dataMap[expkey].(string)
				if !ok || val != expval {
					t.Fatalf("Could not find text object with key %q value %q for comp %v in object %q", expkey, expval, ref, obj.GetName())
				}
			}
		}
	}
}
