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
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type overlayExample struct {
	overlays       []string
	obj            string
	customResource interface{}
}

func applyPatch(ex *overlayExample) (*structpb.Struct, error) {
	patcher := &Patcher{
		Scheme: DefaultPatcherScheme(),
	}
	o, err := converter.Struct.YAMLToProto([]byte(ex.obj))
	var overlays []*bpb.Overlay
	for _, pstr := range ex.overlays {
		p, err := converter.Overlay.YAMLToProto([]byte(pstr))
		if err != nil {
			return nil, fmt.Errorf("Error parsing overlay %s: %v", pstr, err)
		}
		overlays = append(overlays, converter.ToOverlay(p))
	}

	if err != nil {
		return nil, err
	}
	return patcher.ApplyToClusterObjects(overlays, ex.customResource, converter.ToStruct(o))
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
    image: gcr.io/google_containers/kube-apiserver:v1.9.7
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
	overlay := `
name: apiserver-advertise-address
custom_resource_key:
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
	ex := &overlayExample{
		overlays:       []string{overlay},
		obj:            apiServerExample,
		customResource: cr,
	}
	out, err := applyPatch(ex)
	if err != nil {
		t.Fatalf("Error applying overlay: %v", err)
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
	overlay := `
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
	ex := &overlayExample{
		overlays: []string{overlay},
		obj:      apiServerExample,
	}
	out, err := applyPatch(ex)
	if err != nil {
		t.Fatalf("Error applying overlay: %v", err)
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

func TestPatchApplication(t *testing.T) {
	appYaml := `
name: my-app
clusterObjects:
- name: object1
  inlined:
    apiVersion: v1
    kind: Pod
    spec:
      ip: PLACEHOLDER
  overlayCollection:
    overlays:
    - name: ip-overlay
      customResourceKey:
        group: bundles
        version: v1alpha1
        kind: IPAddress
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
  overlayCollection:
    overlays:
    - name: ip-overlay
      customResourceKey:
        group: bundles
        version: v1alpha1
        kind: IPAddress
      templateString: |
        spec:
          containers:
          - name: my-container
            env:
            - name: IP
              value: '{{.spec.ip}}'
    - name: image-overlay
      customResourceKey:
        group: bundles
        version: v1alpha1
        kind: ImageName
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
spec:
  image: %s
`, imageVal)
	imageCR, err := converter.CustomResourceYAMLToMap([]byte(imageCRYAML))
	if err != nil {
		t.Fatalf("CustomResourceYAMLToMap(%v) returned error: %v", imageCRYAML, err)
	}

	apppb, err := converter.ClusterApplication.YAMLToProto([]byte(appYaml))
	if err != nil {
		t.Fatalf("Error parsing cluster application: %v", err)
	}
	app := converter.ToClusterApplication(apppb)
	patcher := &Patcher{
		Scheme: DefaultPatcherScheme(),
	}
	patched, err := patcher.PatchApplication(app, []map[string]interface{}{ipCR, imageCR})
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
  clusterApps:
  - name: my-app
    clusterObjects:
    - name: foo
      inlined:
        apiVersion: v1
        kind: Pod
        spec:
          fooVal: PLACEHOLDER
      overlayCollection:
        overlays:
        - name: foo-overlay
          customResourceKey:
            group: bundles
            version: v1alpha1
            kind: Foo
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
      overlayCollection:
        overlays:
        - name: bar-overlay
          customResourceKey:
            group: bundles
            version: v1alpha1
            kind: Bar
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

func TestGVKFromCustomResource(t *testing.T) {
	testCases := []struct {
		desc        string
		cr          string
		gvk         schema.GroupVersionKind
		errContains string
	}{
		{
			desc: "success case",
			cr: `
apiVersion: bundles/v1alpha1
kind: BundleOptions
foo: Bar
`,
			gvk: schema.GroupVersionKind{
				Group:   "bundles",
				Version: "v1alpha1",
				Kind:    "BundleOptions",
			},
		},
		{
			desc: "bad apiVersion format",
			cr: `
apiVersion: bundles
kind: BundleOptions
foo: Bar
`,
			errContains: "not formatted as group/version",
		},
		{
			desc: "missing apiVersion",
			cr: `
kind: BundleOptions
foo: Bar
`,
			errContains: "no apiVersion field",
		},
		{
			desc: "missing kind",
			cr: `
apiVersion: bundles/v1alpha1
foo: Bar"
`,
			errContains: "no kind field",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			cr, err := converter.CustomResourceYAMLToMap([]byte(tc.cr))
			if err != nil {
				t.Fatalf("CustomResourceYAMLToMap(%s) returned err: %v", tc.cr, err)
			}
			gvk, err := gvkFromCustomResource(cr)
			if !reflect.DeepEqual(gvk, tc.gvk) {
				t.Errorf("gvkFromCustomResource(%v) returned %+v, want %+v", tc.cr, gvk, tc.gvk)
			}
			if tc.errContains != "" {
				if err == nil {
					t.Fatalf("gvkFromCustomResource(%v) should have returned an error but error was nil", tc.cr)
				}
				if !strings.Contains(err.Error(), tc.errContains) {
					t.Fatalf("gvkFromCustomResource(%v) error message should have contained: %v, Got: %v", tc.cr, tc.errContains, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("gvkFromCustomResource(%v) returned unexpected error: %v", tc.cr, err)
			}
		})
	}
}
