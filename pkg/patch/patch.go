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
	"bytes"
	"fmt"
	"regexp"
	"text/template"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/validation"
	log "github.com/golang/glog"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/strategicpatch"

	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	structpb "github.com/golang/protobuf/ptypes/struct"
)

// Patcher applies overlays to objects.
type Patcher struct {
	// Bundle is the bundle from which the Patcher was created.
	Bundle *bpb.ClusterBundle

	// CustomResourceValidator is a validator for CRDs.
	CustomResourceValidator *validation.CustomResourceValidator

	// Scheme wraps the Runtime Kubernetes schema.
	Scheme *PatcherScheme
}

// NewPatcherFromBundleYAML creates a patcher object from a Bundle YAML.
func NewPatcherFromBundleYAML(bundle string) (*Patcher, error) {
	bun, err := converter.Bundle.YAMLToProto([]byte(bundle))
	if err != nil {
		return nil, err
	}
	return NewPatcherFromBundle(converter.ToBundle(bun))
}

// NewPatcherFromBundle creates a patcher object from a Bundle. Note that all
// objects must, at this point, be inlined.
func NewPatcherFromBundle(bundle *bpb.ClusterBundle) (*Patcher, error) {
	crds := make(map[string]*apiextv1beta1.CustomResourceDefinition)
	for _, app := range bundle.GetSpec().GetClusterApps() {
		appName := app.GetName()
		for _, obj := range app.ClusterObjects {
			objName := obj.GetName()
			if obj.GetFile() != nil {
				return nil, fmt.Errorf("for cluster app %q and cluster object %q, found a non-inlined file; all files must be inlined to perform patching", appName, objName)
			}
			inline := obj.GetInlined()
			if inline == nil {
				continue
			}
			if kind := inline.GetFields()["kind"]; kind == nil || kind.GetStringValue() != "CustomResourceDefinition" {
				continue
			}
			crd, err := converter.FromStruct(inline).ToCRD()
			if err != nil {
				return nil, fmt.Errorf("error converting object struct %v to CustomResourceDefinition", crd)
			}
			customKind := crd.Spec.Names.Kind // All types are non-pointer
			if customKind == "" {
				continue
			}
			crds[customKind] = crd
		}
	}

	crdValidator, err := validation.NewCustomResourceValidator(bundle, &validation.CustomResourceOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create CRD validator: %v", err)
	}
	return &Patcher{
		Bundle:                  bundle,
		CustomResourceValidator: crdValidator,
		Scheme:                  DefaultPatcherScheme(),
	}, nil
}

// overlayRequired returns whether the overlay is required or is merely optional.
func (t *Patcher) overlayRequired(overlay *bpb.Overlay) bool {
	if overlay.OverlayConstraint == bpb.OverlayConstraint_REQUIRED ||
		overlay.OverlayConstraint == bpb.OverlayConstraint_UNKNOWN_PATCH_USAGE {
		// Required overlays must always be applied.
		return true
	}
	return false
}

// Detemplatize takes a overlay message, and with the params specified in the
// patcher, and applies them to the overlay template, and returns the overlay.
// - Returns an error if the overlay doesn't exist
// - Returns an error there aren't sufficient parameters.
func (t *Patcher) Detemplatize(overlay *bpb.Overlay, customResource interface{}) (*structpb.Struct, error) {
	key := overlay.Name
	if key == "" {
		return nil, fmt.Errorf("overlay name was empty for overlay %v", overlay)
	}

	err := t.CustomResourceValidator.Validate(customResource)
	if err != nil {
		return nil, fmt.Errorf("error validating custom resource: %v", err)
	}

	tmpl := overlay.GetTemplateString()
	if tmpl == "" {
		return nil, fmt.Errorf("template string was empty for overlay %v", overlay)
	}

	overlayTmpl, err := template.New("overlay:" + key).Parse(tmpl)
	if err != nil {
		return nil, fmt.Errorf("for overlay %q, template parsing failed: %v ", key, err)
	}
	overlayTmpl.Option("missingkey=error")

	var doc bytes.Buffer
	if err := overlayTmpl.Execute(&doc, customResource); err != nil {
		if t.overlayRequired(overlay) {
			return nil, fmt.Errorf("for overlay %q, error applying template params: %v", key, err)
		}
		log.Infof("Missing parameters in overlay, but overlay is optional: %v. not applying overlay: %v", err, overlay.Name)
		return nil, nil
	}

	pb, err := converter.Struct.YAMLToProto(doc.Bytes())
	if err != nil {
		return nil, fmt.Errorf("for overlay %q, error parsing YAML: %v", key, err)
	}
	return converter.ToStruct(pb), nil
}

// ApplyToClusterObjects applies overlays to cluster objects, always returning a
// new object copy of the original object with the overlays (if any) applied.
//
// Note: every cluster object must have it's type registered in with the K8S
// scheme / codec factory.
func (t *Patcher) ApplyToClusterObjects(overlays []*bpb.Overlay, customResource interface{}, kubeObj *structpb.Struct) (*structpb.Struct, error) {
	objJSON, err := converter.Struct.ProtoToJSON(kubeObj)
	if err != nil {
		return nil, err
	}
	for _, pat := range overlays {
		p, err := t.Detemplatize(pat, customResource)
		if err != nil {
			return nil, err
		}
		if p == nil {
			// No patch to apply for whatever reason. Continue on with applying other overlays.
			continue
		}
		patchJSON, err := converter.Struct.ProtoToJSON(p)
		if err != nil {
			return nil, err
		}

		deserializer := t.Scheme.Codecs.UniversalDeserializer()
		kubeObj, err := runtime.Decode(deserializer, objJSON)
		if err != nil {
			return nil, fmt.Errorf("could not decode kube object: %v", err)
		}

		// The Magic! Strategic Merge Patch Away!
		objJSON, err = strategicpatch.StrategicMergePatch(objJSON, patchJSON, kubeObj)
		if err != nil {
			return nil, fmt.Errorf("for overlay %q, error applying overlay: %v", pat.Name, err)
		}
	}
	o, err := converter.Struct.JSONToProto(objJSON)
	return converter.ToStruct(o), err
}

// PatchApplication applies overlays to all the cluster objects in a given ClusterApplication,
// always returning a new application copy of the original application with the overlays (if any)
// applied.
func (t *Patcher) PatchApplication(app *bpb.ClusterApplication, customResources []map[string]interface{}) (*bpb.ClusterApplication, error) {
	crMap, err := createCustomResourceMap(customResources)
	if err != nil {
		return nil, fmt.Errorf("could not patch application %q: %v", app.GetName(), err)
	}
	return t.patchApplication(app, crMap)
}

// patchApplication applies overlays to all the cluster objects in a given ClusterApplication,
// always returning a new application copy of the original application with the overlays (if any)
// applied. It takes in a map from custom resource kind to custom resource instance for ease of
// lookup.
func (t *Patcher) patchApplication(app *bpb.ClusterApplication, crMap map[schema.GroupVersionKind]interface{}) (*bpb.ClusterApplication, error) {
	clonedApp := converter.CloneApplication(app)
	for _, co := range clonedApp.GetClusterObjects() {
		obj := co.GetInlined()
		if obj == nil {
			return nil, fmt.Errorf("object %q was not inlined", co.GetName())
		}
		for _, overlay := range co.GetOverlayCollection().GetOverlays() {
			var err error
			// Pass each overlay separately to update the cluster object along with the corresponding CR.
			cr, found := crMap[converter.ToGVK(overlay.GetCustomResourceKey())]
			if !found {
				return nil, fmt.Errorf("could not patch object %q: no custom resource found for %q", co.GetName(), overlay.GetCustomResourceKey())
			}
			obj, err = t.ApplyToClusterObjects([]*bpb.Overlay{overlay}, cr, obj)
			if err != nil {
				return nil, fmt.Errorf("could not patch object %q: %s", co.GetName(), err)
			}
		}
		co.KubeData = &bpb.ClusterObject_Inlined{obj}
	}
	return clonedApp, nil
}

// PatchBundle applies overlays to all the cluster objects in the Patcher Bundle. It returns a new
// bundle copy of the original bundle with the overlays (if any) applied.
func (t *Patcher) PatchBundle(customResources []map[string]interface{}) (*bpb.ClusterBundle, error) {
	bundle := converter.CloneBundle(t.Bundle)
	apps := bundle.GetSpec().GetClusterApps()[:]

	crMap, err := createCustomResourceMap(customResources)
	if err != nil {
		return nil, fmt.Errorf("could not patch bundle: %s", err)
	}

	for i, app := range t.Bundle.Spec.GetClusterApps() {
		patched, err := t.patchApplication(app, crMap)
		if err != nil {
			return nil, fmt.Errorf("could not patch application %q: %s", app.GetName(), err)
		}
		// Replace each app in the bundle with the patched app.
		apps[i] = patched
	}
	bundle.GetSpec().ClusterApps = apps
	return bundle, nil
}

// createCustomResourceMap creates a map from the custom resource GVK to the custom resource
// instance.
func createCustomResourceMap(customResources []map[string]interface{}) (map[schema.GroupVersionKind]interface{}, error) {
	crMap := make(map[schema.GroupVersionKind]interface{})
	for _, cr := range customResources {
		gvk, err := gvkFromCustomResource(cr)
		if err != nil {
			return nil, fmt.Errorf("error creating custom resource map: %v", err)
		}
		crMap[gvk] = cr
	}
	return crMap, nil
}

// gvkFromCustomResource extracts the GroupVersionKind out of a custom resource.
// - returns an error if there is no apiVersion or kind field in the custom resource
// - returns an error if the apiVersion value is not of the format "group/version"
// TODO: parse CustomResource into a RawExtension instead of a map.
func gvkFromCustomResource(cr map[string]interface{}) (schema.GroupVersionKind, error) {
	gvk := schema.GroupVersionKind{}
	apiVersion := cr["apiVersion"]
	if apiVersion == nil {
		return schema.GroupVersionKind{}, fmt.Errorf("no apiVersion field was found for custom resource %v", cr)
	}
	matches := regexp.MustCompile("(.+)/(.+)").FindStringSubmatch(apiVersion.(string))
	// The number of matches should be 3 - FindSubstringMatch returns the full matched string in
	// addition to the matched subexpressions.
	if matches == nil || len(matches) != 3 {
		return schema.GroupVersionKind{}, fmt.Errorf("custom resource apiVersion is not formatted as group/version: got %q", apiVersion)
	}
	gvk.Group, gvk.Version = matches[1], matches[2]
	kind := cr["kind"]
	if kind == nil {
		return schema.GroupVersionKind{}, fmt.Errorf("no kind field was found for custom resource %v", cr)
	}
	gvk.Kind = kind.(string)
	return gvk, nil
}
