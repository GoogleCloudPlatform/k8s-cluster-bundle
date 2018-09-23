package converter

import (
	"reflect"
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/core"
)

func TestObjectRefFromCustomResource(t *testing.T) {
	testCases := []struct {
		desc        string
		cr          string
		ref         core.ObjectReference
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
			ref: core.ObjectReference{
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
		t.Run(tc.desc+" via Raw", func(t *testing.T) {
			cr, err := KubeResourceYAMLToMap([]byte(tc.cr))
			if err != nil {
				t.Fatalf("KubeResourceYAMLToMap(%s) returned err: %v", tc.cr, err)
			}
			ref, err := ObjectRefFromRawKubeResource(cr)
			if !reflect.DeepEqual(ref, tc.ref) {
				t.Errorf("ObjectRefFromRawKubeResource(%v) returned %+v, want %+v", tc.cr, ref, tc.ref)
			}
			if tc.errContains != "" {
				if err == nil {
					t.Fatalf("ObjectRefFromRawKubeResource(%v) should have returned an error but error was nil", tc.cr)
				}
				if !strings.Contains(err.Error(), tc.errContains) {
					t.Fatalf("ObjectRefFromRawKubeResource(%v) error message should have contained: %v, Got: %v", tc.cr, tc.errContains, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("ObjectRefFromRawKubeResource(%v) returned unexpected error: %v", tc.cr, err)
			}
		})

		// These two methods of creating object refs should be identical. So, ensure they are.
		t.Run(tc.desc+" via Structpb", func(t *testing.T) {
			s, err := Struct.YAMLToProto([]byte(tc.cr))
			if err != nil {
				t.Fatalf("Error converting yaml to struct: %v", err)
			}
			ref, err := FromStruct(ToStruct(s)).ToObjectRef()
			if !reflect.DeepEqual(ref, tc.ref) {
				t.Errorf("ToObjectRef(%v) returned %+v, want %+v", tc.cr, ref, tc.ref)
			}
			if tc.errContains != "" {
				if err == nil {
					t.Fatalf("ToObjectRef(%v) should have returned an error but error was nil", tc.cr)
				}
				if !strings.Contains(err.Error(), tc.errContains) {
					t.Fatalf("ToObjectRef(%v) error message should have contained: %v, Got: %v", tc.cr, tc.errContains, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("ToObjectRef(%v) returned unexpected error: %v", tc.cr, err)
			}
		})
	}
}
