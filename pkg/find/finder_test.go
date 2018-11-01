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

package find

import (
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/core"
)

var validComponentExample = `
components:
- spec:
    canonicalName: etcd-server
    objects:
    - apiVersion: v1
      kind: Pod
      metadata:
        name: pod
    - apiVersion: v1
      kind: Pod
      metadata:
        name: dwerp

- spec:
    name: kube-apiserver
    objects:
    - apiVersion: v1
      kind: Pod
      metadata:
        name: pod
`

func TestBundleFinder(t *testing.T) {
	b, err := converter.FromYAMLString(validComponentExample).ToComponentData()
	if err != nil {
		t.Fatalf("error converting bundle: %v", err)
	}
	finder := NewComponentFinder(b.Components)
	if err != nil {
		t.Fatalf("error creating bundle finder: %v", err)
	}
	testCases := []struct {
		desc       string
		compName   string
		objName    string
		shouldFind bool
	}{
		{
			desc:       "success: cluster comp lookup",
			compName:   "etcd-server",
			shouldFind: true,
		},
		{
			desc:       "failure: cluster comp lookup",
			compName:   "etcd-server-bloop",
			shouldFind: false,
		},
		{
			desc:       "success: cluster obj lookup",
			compName:   "etcd-server",
			objName:    "pod",
			shouldFind: true,
		},
		{
			desc:       "failure: cluster obj lookup",
			compName:   "etcd-server",
			objName:    "blorp",
			shouldFind: false,
		},
	}
	for _, tc := range testCases {
		if tc.objName != "" && tc.compName != "" {
			vl := finder.Objects(core.ComponentKey{CanonicalName: tc.compName}, core.ObjectRef{Name: tc.objName})
			var v *unstructured.Unstructured
			if len(vl) > 0 {
				v = vl[0]
			}
			if v == nil && tc.shouldFind {
				t.Errorf("Test %v: Got unexpected nil response for cluster comp object lookup", tc.desc)
			} else if v != nil && !tc.shouldFind {
				t.Errorf("Test %v: Got unexpected non-nil response %v for cluster comp object lookup", tc.desc, v)
			}

		} else if tc.compName != "" {
			v := finder.UniqueComponentFromName(tc.compName)
			if v == nil && tc.shouldFind {
				t.Errorf("Test %v: Got unexpected nil response for cluster comp lookup", tc.desc)
			} else if v != nil && !tc.shouldFind {
				t.Errorf("Test %v: Got unexpected non-nil response %v for cluster comp lookup", tc.desc, v)
			}
		} else {
			t.Errorf("Unexpected fallthrough for testcase %v", tc)
		}
	}
}

var validComponent = `
apiVersion: 'bundle.gke.io/v1alpha1'
kind: ComponentPackage
spec:
  canonicalName: kube-apiserver
  objects:
  - apiVersion: v1
    kind: Pod
    metadata:
      name: pody

  - apiVersion: v1
    kind: Pod
    metadata:
      name: dodo

  - apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: kube-proxy

  - apiVersion: extensions/v1beta1
    kind: DaemonSet
    metadata:
      name: kube-proxy
`

func TestComponentFinder_PartialLookup(t *testing.T) {
	c, err := converter.FromYAMLString(validComponent).ToComponentPackage()
	if err != nil {
		t.Fatalf("error converting componente: %v", err)
	}
	finder := NewObjectFinder(c)

	testCases := []struct {
		desc string
		ref  core.ObjectRef
		exp  []string
	}{
		{
			desc: "get everything",
			ref:  core.ObjectRef{},
			exp:  []string{"pody", "dodo", "kube-proxy", "kube-proxy"},
		},
		{
			desc: "get apiversion",
			ref:  core.ObjectRef{APIVersion: "v1"},
			exp:  []string{"pody", "dodo", "kube-proxy"},
		},
		{
			desc: "get kind",
			ref:  core.ObjectRef{Kind: "Pod"},
			exp:  []string{"pody", "dodo"},
		},
		{
			desc: "get name",
			ref:  core.ObjectRef{Name: "pody"},
			exp:  []string{"pody"},
		},
		{
			desc: "get name, multiple hits",
			ref:  core.ObjectRef{Name: "kube-proxy"},
			exp:  []string{"kube-proxy", "kube-proxy"},
		},
		{
			desc: "get specific",
			ref:  core.ObjectRef{Name: "kube-proxy", APIVersion: "v1"},
			exp:  []string{"kube-proxy"},
		},
		{
			desc: "get none",
			ref:  core.ObjectRef{Name: "kube-proxy", APIVersion: "zed"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			found := finder.Objects(tc.ref)
			names := getObjNames(found)
			if !reflect.DeepEqual(names, tc.exp) {
				t.Errorf("CluusterObjects(): got %v but wanted %v", names, tc.exp)
			}
		})
	}
}

func getObjNames(objs []*unstructured.Unstructured) []string {
	var names []string
	for _, o := range objs {
		names = append(names, o.GetName())
	}
	return names
}
