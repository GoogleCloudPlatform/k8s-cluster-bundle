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
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/files"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/testutil"
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

var kubeApiserverComponent = `
kind: ComponentBuilder
componentName: kube-apiserver
version: 1.2.3
objectFiles:
- url: '/path/to/kube_apiserver.yaml'`

var kubeApiserver = `
apiVersion: v1
kind: Zork
metadata:
  name: biffbam
biff: bam`

var podTemplate = `
kind: pod
metadata:
  name: {{.foo}}`

var defaultFiles = map[string][]byte{
	"/path/to/apiserver-component.yaml": []byte(kubeApiserverComponent),
	"/path/to/kube_apiserver.yaml":      []byte(kubeApiserver),
}

type bundleRef struct {
	setName string
	version string
}

type objCheck struct {
	name       string
	subStrings []string

	// For if it's imported as raw text
	cfgMap map[string]string

	// For if it's imported as binary raw text. Still parsed in the
	// unstructured.Unstructured as a string, though.
	binaryCfgMap map[string]string
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
		pathToBundle string
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
					name: "kube-apiserver-1.2.3",
					ref: bundle.ComponentReference{
						ComponentName: "kube-apiserver",
						Version:       "1.2.3",
					},
					obj: []objCheck{
						{name: "biffbam"},
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
					name: "kube-apiserver-1.2.3",
					ref: bundle.ComponentReference{
						ComponentName: "kube-apiserver",
						Version:       "1.2.3",
					},
					obj: []objCheck{
						{name: "biffbam"},
					},
				},
			},
		},
		{
			desc: "success: inline basic bundle + path rewriting",
			data: `
kind: BundleBuilder
setName: foo-bundle
version: 1.2.3
componentFiles:
- url: path/to/apiserver-component.yaml`,
			files: map[string][]byte{
				"/derp/path/to/apiserver-component.yaml": []byte(kubeApiserverComponent),
				"/path/to/kube_apiserver.yaml":           []byte(kubeApiserver),
			},
			pathToBundle: "/derp/derp.yaml",
			expBun:       bundleRef{setName: "foo-bundle", version: "1.2.3"},
			expComps: []compRef{
				{
					name: "kube-apiserver-1.2.3",
					ref: bundle.ComponentReference{
						ComponentName: "kube-apiserver",
						Version:       "1.2.3",
					},
					obj: []objCheck{
						{name: "biffbam"},
					},
				},
			},
		},

		{
			desc: "success: inline basic bundle + file prefix + SetAndComponent policy",
			data: `
kind: BundleBuilder
setName: foo-bundle
version: 1.2.3
componentNamePolicy: SetAndComponent
componentFiles:
- url: file:///path/to/apiserver-component.yaml`,
			files:  defaultFiles,
			expBun: bundleRef{setName: "foo-bundle", version: "1.2.3"},
			expComps: []compRef{
				{
					name: "foo-bundle-1.2.3-kube-apiserver-1.2.3",
					ref: bundle.ComponentReference{
						ComponentName: "kube-apiserver",
						Version:       "1.2.3",
					},
					obj: []objCheck{
						{name: "biffbam"},
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
					name: "kube-apiserver-1.2.3",
					ref: bundle.ComponentReference{
						ComponentName: "kube-apiserver",
						Version:       "1.2.3",
					},
					obj: []objCheck{
						{
							name: "some-raw-text",
							subStrings: []string{
								"foobar",
								"boobar",
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
---
---
apiVersion: v1
kind: Zork
metadata:
  name: biffbam
biff: bam
---`),
			},
			expBun: bundleRef{setName: "multi-bundle", version: "2.2.3"},
			expComps: []compRef{
				{
					name: "kube-multi-2.3.4",
					ref: bundle.ComponentReference{
						ComponentName: "kube-multi",
						Version:       "2.3.4",
					},
					obj: []objCheck{
						{name: "foobar"},
						{name: "biffbam"},
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
						{name: "foo-comp"},
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
				t.Fatal(err)
			}

			inliner := NewInlinerWithScheme(files.FileScheme, &fakeLocalReader{tc.files})
			got, err := inliner.BundleFiles(ctx, data, tc.pathToBundle)
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
		desc            string
		data            string
		files           map[string][]byte
		pathToComponent string
		expErrSubstr    string
		expComp         compRef
	}{
		{
			desc: "success: inline basic component",
			data: kubeApiserverComponent,
			files: map[string][]byte{
				"/path/to/apiserver-component.yaml": []byte(kubeApiserverComponent),
				"/path/to/kube_apiserver.yaml":      []byte(kubeApiserver),
			},
			expComp: compRef{
				name: "kube-apiserver-1.2.3",
				ref: bundle.ComponentReference{
					ComponentName: "kube-apiserver",
					Version:       "1.2.3",
				},
				obj: []objCheck{
					{name: "biffbam"},
				},
			},
		},
		{
			desc: "success: inline basic component, absolute path, no path rewriting",
			data: kubeApiserverComponent,
			files: map[string][]byte{
				"/path/to/apiserver-component.yaml": []byte(kubeApiserverComponent),
				"/path/to/kube_apiserver.yaml":      []byte(kubeApiserver),
			},
			pathToComponent: "/my/component.yaml",
			expComp: compRef{
				name: "kube-apiserver-1.2.3",
				ref: bundle.ComponentReference{
					ComponentName: "kube-apiserver",
					Version:       "1.2.3",
				},
				obj: []objCheck{
					{name: "biffbam"},
				},
			},
		},
		{
			desc: "success: inline basic component, path rewriting",
			data: `kind: ComponentBuilder
componentName: kube-apiserver
version: 1.2.3
objectFiles:
- url: 'path/to/kube_apiserver.yaml'`,
			files: map[string][]byte{
				"/my/path/to/apiserver-component.yaml": []byte(kubeApiserverComponent),
				"/my/path/to/kube_apiserver.yaml":      []byte(kubeApiserver),
			},
			pathToComponent: "/my/component.yaml",
			expComp: compRef{
				name: "kube-apiserver-1.2.3",
				ref: bundle.ComponentReference{
					ComponentName: "kube-apiserver",
					Version:       "1.2.3",
				},
				obj: []objCheck{
					{name: "biffbam"},
				},
			},
		},

		{
			desc: "success: component, raw text, annotated",
			data: `
kind: ComponentBuilder
componentName: kube-apiserver-blob
version: 1.2.3
rawTextFiles:
- name: data-blob
  annotations:
    foo: bar
    biff: bam
  labels:
    zip: zap
  files:
  - url: '/path/to/kube_apiserver.yaml'`,
			files: defaultFiles,
			expComp: compRef{
				name: "kube-apiserver-blob-1.2.3",
				ref: bundle.ComponentReference{
					ComponentName: "kube-apiserver-blob",
					Version:       "1.2.3",
				},
				obj: []objCheck{
					{name: "data-blob"},
					{
						name: "data-blob",
						subStrings: []string{
							"foo: bar",
							"zip: zap",
						},
					},
					{
						name: "data-blob",
						subStrings: []string{
							"biff: bam",
						},
					},
				},
			},
		},

		{
			desc: "success: component, raw text, binary",
			data: `
kind: ComponentBuilder
componentName: binary-blob
version: 1.2.3
rawTextFiles:
- name: data-blob
  asBinary: true
  files:
  - url: '/path/to/blobby.yaml'`,
			files: map[string][]byte{
				"/path/to/blobby.yaml": []byte("blar"),
			},
			expComp: compRef{
				name: "binary-blob-1.2.3",
				ref: bundle.ComponentReference{
					ComponentName: "binary-blob",
					Version:       "1.2.3",
				},
				obj: []objCheck{
					{
						name: "data-blob",
						subStrings: []string{
							base64.StdEncoding.EncodeToString([]byte("blar")),
						},
					},
				},
			},
		},

		{
			desc: "success: component, raw text, binary: relative path",
			data: `
kind: ComponentBuilder
componentName: binary-blob
version: 1.2.3
rawTextFiles:
- name: data-blob
  asBinary: true
  files:
  - url: 'path/to/blobby.yaml'`,
			files: map[string][]byte{
				"/foo/bar/path/to/blobby.yaml": []byte("blar"),
			},
			pathToComponent: "/foo/bar/component.yaml",
			expComp: compRef{
				name: "binary-blob-1.2.3",
				ref: bundle.ComponentReference{
					ComponentName: "binary-blob",
					Version:       "1.2.3",
				},
				obj: []objCheck{
					{
						name: "data-blob",
						subStrings: []string{
							base64.StdEncoding.EncodeToString([]byte("blar")),
						},
					},
				},
			},
		},

		{
			desc: "success: component, object template builder",
			data: `
kind: ComponentBuilder
componentName: binary-blob
version: 1.2.3
objectFiles:
- url: '/path/to/tmpl-builder.yaml'`,
			files: map[string][]byte{
				"/path/to/tmpl-builder.yaml": []byte(`
kind: ObjectTemplateBuilder
metadata:
  name: obj-tmpl
file:
  url: '/path/to/tmpl.yaml'
optionsSchema:
  properties:
    foo:
      type: string
`),
				"/path/to/tmpl.yaml": []byte(podTemplate),
			},
			expComp: compRef{
				name: "binary-blob-1.2.3",
				ref: bundle.ComponentReference{
					ComponentName: "binary-blob",
					Version:       "1.2.3",
				},
				obj: []objCheck{
					{
						name: "obj-tmpl",
						subStrings: []string{
							"name: {{.foo}}",
							"type: string",
							"type: go-template",
						},
					},
				},
			},
		},

		{
			desc: "success: component, object template builder: relative path",
			data: `
kind: ComponentBuilder
componentName: binary-blob
version: 1.2.3
objectFiles:
- url: '/path/to/tmpl-builder.yaml'`,
			files: map[string][]byte{
				"/path/to/tmpl-builder.yaml": []byte(`
kind: ObjectTemplateBuilder
metadata:
  name: obj-tmpl
file:
  url: 'manifest/tmpl.yaml'
optionsSchema:
  properties:
    foo:
      type: string
`),
				"/path/to/manifest/tmpl.yaml": []byte(podTemplate),
			},
			expComp: compRef{
				name: "binary-blob-1.2.3",
				ref: bundle.ComponentReference{
					ComponentName: "binary-blob",
					Version:       "1.2.3",
				},
				obj: []objCheck{
					{
						name: "obj-tmpl",
						subStrings: []string{
							"name: {{.foo}}",
							"type: string",
							"type: go-template",
						},
					},
				},
			},
		},

		{
			desc: "success: component, template files",
			data: `
kind: ComponentBuilder
componentName: binary-blob
version: 1.2.3
templateType: go-template
templateFiles:
- url: '/path/to/manifest/tmpl.yaml'`,
			files: map[string][]byte{
				"/path/to/manifest/tmpl.yaml": []byte(podTemplate),
			},
			expComp: compRef{
				name: "binary-blob-1.2.3",
				ref: bundle.ComponentReference{
					ComponentName: "binary-blob",
					Version:       "1.2.3",
				},
				obj: []objCheck{
					{
						subStrings: []string{
							"name: {{.foo}}",
							"type: go-template",
						},
					},
				},
			},
		},

		{
			desc: "success: component, template files: relative path, no templateType",
			data: `
kind: ComponentBuilder
componentName: binary-blob
version: 1.2.3
templateFiles:
- url: 'manifest/tmpl.yaml'`,
			files: map[string][]byte{
				"/path/to/manifest/tmpl.yaml": []byte(podTemplate),
			},
			pathToComponent: "/path/to/builder.yaml",
			expComp: compRef{
				name: "binary-blob-1.2.3",
				ref: bundle.ComponentReference{
					ComponentName: "binary-blob",
					Version:       "1.2.3",
				},
				obj: []objCheck{
					{
						subStrings: []string{
							"name: {{.foo}}",
							"type: go-template",
						},
					},
				},
			},
		},

		// Error cases.
		{
			desc:         "fail: can't read file",
			data:         kubeApiserverComponent,
			files:        make(map[string][]byte),
			expErrSubstr: "reading file",
		},
		{
			desc: "fail: can't read raw text file",
			data: `
kind: ComponentBuilder
componentName: kube-apiserver
version: 1.2.3
rawTextFiles:
- name: foo-group
  files:
  - url: '/path/to/raw-text.yaml'
  - url: '/path/to/rawer-text.yaml'`,
			files:        make(map[string][]byte),
			expErrSubstr: "error reading raw text file for",
		},
		{
			desc: "fail: can't read raw text file group: no name",
			data: `
kind: ComponentBuilder
componentName: kube-apiserver
version: 1.2.3
rawTextFiles:
- files:
  - url: '/path/to/raw-text.yaml'
  - url: '/path/to/rawer-text.yaml'`,
			files:        make(map[string][]byte),
			expErrSubstr: "error reading raw text file group",
		},
		{
			desc: "fail: can't convert object to unstructured",
			data: kubeApiserverComponent,
			files: map[string][]byte{
				"/path/to/kube_apiserver.yaml": []byte("blah"),
			},
			expErrSubstr: "while converting to Unstructured",
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
			expErrSubstr: "converting multi-doc object",
		},
		{
			desc: "error: invalid specified name",
			data: `
kind: ComponentBuilder
metadata:
  name: this-
componentName: kube-apiserver
version: 1.2.3
objectFiles:
- url: '/path/to/kube_apiserver.yaml'`,
			files:        defaultFiles,
			expErrSubstr: "DNS-1123",
		},
		{
			desc: "error: invalid generated name",
			data: `
kind: ComponentBuilder
metadata:
componentName: kube-apiserver
version: 1.2.3-
objectFiles:
- url: '/path/to/kube_apiserver.yaml'`,
			files:        defaultFiles,
			expErrSubstr: "DNS-1123",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			data, err := converter.FromYAMLString(tc.data).ToComponentBuilder()
			if err != nil {
				t.Fatalf("Error converting component: %v", err)
			}

			inliner := NewInlinerWithScheme(files.FileScheme, &fakeLocalReader{tc.files})
			got, err := inliner.ComponentFiles(ctx, data, tc.pathToComponent)
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
		gotObjYAML := make(map[string]string)
		for _, obj := range comp.Spec.Objects {
			objYAML, err := converter.FromObject(obj).ToYAMLString()
			if err != nil {
				t.Fatalf("converting object %q to yaml: %v", obj.GetName(), err)
			}
			gotObjYAML[obj.GetName()] = objYAML
		}

		for _, objCheck := range ec.obj {
			obj := gotObjYAML[objCheck.name]
			if obj == "" {
				t.Errorf("got objs %v, but expected object with name name %q.", gotObjYAML, objCheck.name)
			}
			for _, expSubstring := range objCheck.subStrings {
				if !strings.Contains(obj, expSubstring) {
					t.Errorf("got obj %s, but expected it to contain %s", obj, expSubstring)
				}
			}
		}
	}
}
