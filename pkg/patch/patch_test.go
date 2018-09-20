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
	"fmt"
	"reflect"
	"strings"
	"testing"

	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	structpb "github.com/golang/protobuf/ptypes/struct"
	corev1 "k8s.io/api/core/v1"
)

type patchExample struct {
	patches        []string
	obj            string
	customResource interface{}
}

func applyPatch(ex *patchExample) (*structpb.Struct, error) {
	patcher := &Patcher{
		Scheme: DefaultPatcherScheme(),
	}
	o, err := converter.Struct.YAMLToProto([]byte(ex.obj))
	var patches []*bpb.Patch
	for _, pstr := range ex.patches {
		p, err := converter.Patch.YAMLToProto([]byte(pstr))
		if err != nil {
			return nil, fmt.Errorf("Error parsing patch %s: %v", pstr, err)
		}
		patches = append(patches, converter.ToPatch(p))
	}

	if err != nil {
		return nil, err
	}
	return patcher.ApplyToClusterObjects(patches, ex.customResource, converter.ToStruct(o))
}

const apiServerExample = `
apiVersion: v1
kind: Pod
metadata:
  name: kube-apiserver
  namespace: kube-system
spec:
  containers:
  - name: kube-apiserver
    command:
    - /bin/sh
    - -c
    - exec /usr/local/bin/kube-apiserver
      --advertise-address=$(ADVERTISE_ADDRESS)
      --storage-backend=$(ETCD_BACKEND)
    image: k8s.gcr.io/kube-apiserver:v1.9.7
    env:
    - name: ADVERTISE_ADDRESS
      value: 10.0.0.1
    - name: ETCD_BACKEND
      value: etcd2
    volumeMounts:
    - mountPath: /etc/gce.conf
      name: cloudconfigmount
      readOnly: true
    - mountPath: /etc/gcp_authz.config
      name: webhookconfigmount
      readOnly: false
  volumes:
  - hostPath:
      path: /etc/gce.conf
      type: FileOrCreate
    name: cloudconfigmount
  - hostPath:
      path: /etc/gcp_authz.config
      type: FileOrCreate
    name: webhookconfigmount
`

func TestApplyPatchEnvVar(t *testing.T) {
	patchBase := `
name: apiserver-advertise-address
objectRef:
  kind: ApiServerAddress
templateString: |
  spec:
    containers:
    - name: kube-apiserver
      env:
       - name: ADVERTISE_ADDRESS
         value: '{{.advertiseAddress}}'
`
	val := "8.8.8.8"
	crYAML := fmt.Sprintf(`
kind: ApiServerAddress
advertiseAddress: %s
`, val)
	cr, err := converter.CustomResourceYAMLToMap([]byte(crYAML))
	if err != nil {
		t.Fatalf("CustomResourceYAMLToMap(%v) returned error: %v", crYAML, err)
	}
	ex := &patchExample{
		patches:        []string{patchBase},
		obj:            apiServerExample,
		customResource: cr,
	}
	out, err := applyPatch(ex)
	if err != nil {
		t.Fatalf("Error applying patch: %v", err)
	}
	if out == nil {
		t.Errorf("Patch file was unexpectedly nil")
	}
	outby, _ := converter.Struct.ProtoToYAML(out)
	if !strings.Contains(string(outby), val) {
		t.Errorf("yaml did not contain %v", val)
	}
}

func TestApplyDeleteVolumes(t *testing.T) {
	patchBase := `
name: apiserver-delete-volumes
templateString: |
  spec:
    containers:
    - name: kube-apiserver
      volumeMounts:
      - mountPath: /etc/gce.conf
        $patch: delete
    volumes:
    - name: cloudconfigmount
      $patch: delete
`
	ex := &patchExample{
		patches: []string{patchBase},
		obj:     apiServerExample,
	}
	out, err := applyPatch(ex)
	if err != nil {
		t.Fatalf("Error applying patch: %v", err)
	}
	if out == nil {
		t.Errorf("Patch file was unexpectedly nil")
	}
	outby, _ := converter.Struct.ProtoToYAML(out)
	if val := "cloudconfigmount"; strings.Contains(string(outby), val) {
		t.Errorf("yaml contained %q", val)
	}
	if val := "webhookconfigmount"; !strings.Contains(string(outby), val) {
		t.Errorf("yaml did not contain %q", val)
	}
}

func TestPatchComponent(t *testing.T) {
	compYAML := `
name: my-app
clusterObjects:
- name: object1
  inlined:
    apiVersion: v1
    kind: Pod
    spec:
      ip: PLACEHOLDER
  patchCollection:
    patches:
    - name: ip-patch
      objectRef:
        apiVersion: bundles/v1alpha1
        kind: IPAddress
        name: Zed
      templateString: |
        spec:
          ip: '{{.spec.ip}}'
- name: object2
  inlined:
    apiVersion: v1
    kind: Pod
    spec:
      containers:
      - name: my-container
        image: PLACEHOLDER
        env:
        - name: IP
          value: PLACEHOLDER
  patchCollection:
    patches:
    - name: ip-patch
      objectRef:
        apiVersion: bundles/v1alpha1
        kind: IPAddress
        name: Zed
      templateString: |
        spec:
          containers:
          - name: my-container
            env:
            - name: IP
              value: '{{.spec.ip}}'
    - name: image-patch
      objectRef:
        apiVersion: bundles/v1alpha1
        kind: ImageName
        name: Zed
      templateString: |
        spec:
          containers:
          - name: my-container
            image: '{{.spec.image}}'
`
	ipVal := "8.8.8.8"
	ipCRYAML := fmt.Sprintf(`
apiVersion: bundles/v1alpha1
kind: IPAddress
metadata:
  name: Zed
spec:
  ip: %s
`, ipVal)
	ipCR, err := converter.CustomResourceYAMLToMap([]byte(ipCRYAML))
	if err != nil {
		t.Fatalf("CustomResourceYAMLToMap(%v) returned error: %v", ipCRYAML, err)
	}

	imageVal := "path/to/image"
	imageCRYAML := fmt.Sprintf(`
apiVersion: bundles/v1alpha1
kind: ImageName
metadata:
  name: Zed
spec:
  image: %s
`, imageVal)
	imageCR, err := converter.CustomResourceYAMLToMap([]byte(imageCRYAML))
	if err != nil {
		t.Fatalf("CustomResourceYAMLToMap(%v) returned error: %v", imageCRYAML, err)
	}

	compPb, err := converter.ClusterComponent.YAMLToProto([]byte(compYAML))
	if err != nil {
		t.Fatalf("Error parsing cluster application: %v", err)
	}
	comp := converter.ToClusterComponent(compPb)
	patcher := &Patcher{
		Scheme: DefaultPatcherScheme(),
	}
	patched, err := patcher.PatchComponent(comp, []map[string]interface{}{ipCR, imageCR})
	if err != nil {
		t.Fatalf("PatchApplication returned error: %v", err)
	}
	if patched == nil {
		t.Fatalf("Patched application was unexpectedly nil")
	}
	out, _ := converter.Struct.ProtoToYAML(patched)
	if !strings.Contains(string(out), ipVal) {
		t.Errorf("yaml did not contain %v", ipVal)
	}
	if !strings.Contains(string(out), imageVal) {
		t.Errorf("yaml did not contain %v", imageVal)
	}
	// Check for the placeholder string since we're using the same CR in two places,
	// so checking for the presence of the value is insufficient.
	if val := "PLACEHOLDER"; strings.Contains(string(out), val) {
		t.Errorf("yaml contained %v", val)
	}
}

func TestPatchBundle(t *testing.T) {
	bundleYAML := `
apiVersion: 'bundle.k8s.io/v1alpha1'
kind: ClusterBundle
metadata:
  name: bundle-example
spec:
  components:
  - name: my-app
    clusterObjects:
    - name: foo
      inlined:
        apiVersion: v1
        kind: Pod
        spec:
          fooVal: PLACEHOLDER
      patchCollection:
        patches:
        - name: foo-patch
          objectRef:
            apiVersion: bundles/v1alpha1
            kind: Foo
            name: Zed
          templateString: |
            spec:
              fooVal: '{{.spec.fooVal}}'
  - name: another-app
    clusterObjects:
    - name: bar
      inlined:
        apiVersion: v1
        kind: Pod
        spec:
          barVal: PLACEHOLDER
      patchCollection:
        patches:
        - name: bar-patch
          objectRef:
            apiVersion: bundles/v1alpha1
            kind: Bar
            name: Zod
          templateString: |
            spec:
              barVal: '{{.spec.barVal}}'
`
	b, err := converter.Bundle.YAMLToProto([]byte(bundleYAML))
	if err != nil {
		t.Fatalf("error parsing bundle yaml: %v", err)
	}
	bp := converter.ToBundle(b)

	fooVal := "path/to/foo"
	fooCRYAML := fmt.Sprintf(`
apiVersion: bundles/v1alpha1
kind: Foo
metadata:
  name: Zed
spec:
  fooVal: %s
`, fooVal)
	fooCR, err := converter.CustomResourceYAMLToMap([]byte(fooCRYAML))
	if err != nil {
		t.Fatalf("CustomResourceYAMLToMap(%v) returned error: %v", fooCRYAML, err)
	}

	barVal := "xyz"
	barCRYAML := fmt.Sprintf(`
apiVersion: bundles/v1alpha1
kind: Bar
metadata:
  name: Zod
spec:
  barVal: %s
`, barVal)
	barCR, err := converter.CustomResourceYAMLToMap([]byte(barCRYAML))
	if err != nil {
		t.Fatalf("CustomResourceYAMLToMap(%v) returned error: %v", barCRYAML, err)
	}

	patcher := &Patcher{
		Bundle: bp,
		Scheme: DefaultPatcherScheme(),
	}
	patched, err := patcher.PatchBundle([]map[string]interface{}{fooCR, barCR})
	if err != nil {
		t.Fatalf("PatchBundle returned error: %v", err)
	}
	if patched == nil {
		t.Fatalf("Patched bundle was unexpectedly nil")
	}
	out, _ := converter.Bundle.ProtoToYAML(patched)
	if !strings.Contains(string(out), fooVal) {
		t.Errorf("yaml did not contain %v", fooVal)
	}
	if !strings.Contains(string(out), barVal) {
		t.Errorf("yaml did not contain %v", barVal)
	}
}

func TestObjectRefFromCustomResource(t *testing.T) {
	testCases := []struct {
		desc        string
		cr          string
		ref         corev1.ObjectReference
		errContains string
	}{
		{
			desc: "success case",
			cr: `
apiVersion: bundles/v1alpha1
kind: BundleOptions
metadata:
  name: Zoinks
foo: Bar
`,
			ref: corev1.ObjectReference{
				APIVersion: "bundles/v1alpha1",
				Kind:       "BundleOptions",
				Name:       "Zoinks",
			},
		},
		{
			desc: "bad apiVersion format",
			cr: `
apiVersion: bundles
kind: BundleOptions
metadata:
  name: Zoinks
foo: Bar
`,
			errContains: "not formatted as group/version",
		},
		{
			desc: "missing apiVersion",
			cr: `
kind: BundleOptions
metadata:
  name: Zoinks
foo: Bar
`,
			errContains: "no apiVersion field",
		},
		{
			desc: "missing kind",
			cr: `
apiVersion: bundles/v1alpha1
metadata:
  name: Zoinks
foo: Bar"
`,
			errContains: "no kind field",
		},
		{
			desc: "missing metadata",
			cr: `
apiVersion: bundles/v1alpha1
kind: BundleOptions
foo: Bar"
`,
			errContains: "no metadata field",
		},
		{
			desc: "missing metadata.name",
			cr: `
apiVersion: bundles/v1alpha1
kind: BundleOptions
metadata:
  UUID: zoinks
foo: Bar"
`,
			errContains: "no metadata.name field",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			cr, err := converter.CustomResourceYAMLToMap([]byte(tc.cr))
			if err != nil {
				t.Fatalf("CustomResourceYAMLToMap(%s) returned err: %v", tc.cr, err)
			}
			ref, err := objectRefFromCustomResource(cr)
			if !reflect.DeepEqual(ref, tc.ref) {
				t.Errorf("objectRefFromCustomResource(%v) returned %+v, want %+v", tc.cr, ref, tc.ref)
			}
			if tc.errContains != "" {
				if err == nil {
					t.Fatalf("objectRefFromCustomResource(%v) should have returned an error but error was nil", tc.cr)
				}
				if !strings.Contains(err.Error(), tc.errContains) {
					t.Fatalf("objectRefFromCustomResource(%v) error message should have contained: %v, Got: %v", tc.cr, tc.errContains, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("objectRefFromCustomResource(%v) returned unexpected error: %v", tc.cr, err)
			}
		})
	}
}
