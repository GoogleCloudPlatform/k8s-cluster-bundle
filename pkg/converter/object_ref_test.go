package converter

import (
	"reflect"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

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
			cr, err := CustomResourceYAMLToMap([]byte(tc.cr))
			if err != nil {
				t.Fatalf("CustomResourceYAMLToMap(%s) returned err: %v", tc.cr, err)
			}
			ref, err := ObjectRefFromRawKubeResource(cr)
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
