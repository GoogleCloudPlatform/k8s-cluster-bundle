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

package patchtmpl

import (
	"bytes"
	"fmt"
	"text/template"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	bundleext "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundleext/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/filter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/options"
	jsonpatch "github.com/evanphx/json-patch"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

// applier applies options via patch-templates
type applier struct {
	scheme *PatcherScheme

	// TODO(kashomon): Maybe we don't need a full options-filter? Maybe we could
	// just have annotation values?
	tmplFilter *filter.Options
}

// NewApplier creates a new options applier instance. The filter indicates
// keep-only options for what subsets of patches to look for.
func NewApplier(pt *PatcherScheme, opts *filter.Options) options.Applier {
	return &applier{
		scheme:     pt,
		tmplFilter: opts,
	}
}

// NewDefaultApplier creates a default patch template applier.
func NewDefaultApplier() options.Applier {
	return NewApplier(DefaultPatcherScheme(), nil)
}

// ApplyOptions looks for PatchTemplates and applies them to the component objects.
func (a *applier) ApplyOptions(comp *bundle.Component, p options.JSONOptions) (*bundle.Component, error) {
	patches, err := a.makePatches(comp, p)
	if err != nil {
		return nil, err
	}
	return options.ApplyCommon(comp, p, objectApplier(a.scheme, patches))
}

// A parsedPatch has had options applied and has been converted both into raw
// bytes and unstructured.
type parsedPatch struct {
	obj *bundleext.PatchTemplate
	raw []byte
	uns *unstructured.Unstructured
}

func (a *applier) makePatches(comp *bundle.Component, opts options.JSONOptions) ([]*parsedPatch, error) {
	tfil := a.tmplFilter
	if tfil == nil {
		tfil = &filter.Options{}
	}
	tfil.Kinds = append(tfil.Kinds, "PatchTemplate")
	tfil.KeepOnly = true

	// Filter all the objects to include just the patch templates + any additional values.
	objs := filter.NewFilter().Objects(comp.Spec.Objects, tfil)

	// First parse the objects back into go-objects.
	var pts []*bundleext.PatchTemplate
	for _, o := range objs {
		pto := &bundleext.PatchTemplate{}
		err := converter.FromUnstructured(o).ToObject(pto)
		if err != nil {
			return nil, fmt.Errorf("while converting object %v to PatchTemplate: %v", pto, err)
		}
		pts = append(pts, pto)
	}

	// Next, de-templatize the templates.
	var patches []*parsedPatch
	for j, pto := range pts {
		tmpl, err := template.New(fmt.Sprintf("patch-tmpl-%d", j)).Parse(pto.Template)
		if err != nil {
			return nil, fmt.Errorf("parsing patch template %d, %s: %v", j, pto.Template, err)
		}

		// TODO(kashomon): Should this be configurable? This seems like the safest option.
		tmpl = tmpl.Option("missingkey=error")

		var buf bytes.Buffer
		err = tmpl.Execute(&buf, opts)
		if err != nil {
			return nil, fmt.Errorf("while applying options to patch template %d: %v", j, err)
		}

		raw := buf.Bytes()
		uns, err := converter.FromYAML(raw).ToUnstructured()
		if err != nil {
			return nil, fmt.Errorf("while converting patch template %d: %v", j, err)
		}

		rawJson, err := converter.FromObject(uns).ToJSON()
		if err != nil {
			return nil, fmt.Errorf("while converting patch template %d back to json: %v", j, err)
		}

		patches = append(patches, &parsedPatch{
			obj: pto,
			raw: rawJson,
			uns: uns,
		})
	}
	return patches, nil
}

// objectApplier creates a patch object-handler. For each patch, the object
// applier function checks whether a patch can be applied, and if so, then
// applies it.
func objectApplier(scheme *PatcherScheme, patches []*parsedPatch) options.ObjHandler {
	return func(obj *unstructured.Unstructured, ref bundle.ComponentReference, _ options.JSONOptions) (*unstructured.Unstructured, error) {
		objJSON, err := converter.FromObject(obj).ToJSON()
		if err != nil {
			return nil, fmt.Errorf("while converting to JSON: %v", err)
		}
		if len(objJSON) == 0 {
			return nil, fmt.Errorf("converted object JSON was empty")
		}
		fmt.Errorf("obj: %s", objJSON)

		deserializer := scheme.Codecs.UniversalDeserializer()
		for _, pat := range patches {
			if !canApplyPatch(pat, obj) {
				continue
			}

			kubeObj, err := runtime.Decode(deserializer, objJSON)
			_, isUnstructured := kubeObj.(*unstructured.Unstructured)
			var newObjJSON []byte
			if runtime.IsNotRegisteredError(err) || isUnstructured {
				// Strategic merge patch can't handle unstructured.Unstructured or
				// unregistered objects, so defer to normal merge-patch.
				newObjJSON, err = jsonpatch.MergePatch(objJSON, pat.raw)
			} else if err != nil {
				return nil, fmt.Errorf("while decoding object via scheme: %v", err)
			} else {
				newObjJSON, err = strategicpatch.StrategicMergePatch(objJSON, pat.raw, kubeObj)
			}
			if err != nil {
				return nil, fmt.Errorf("while applying patch\n%sto \n%s: %v", pat.raw, objJSON, err)
			}
			objJSON = newObjJSON
		}
		obj, err = converter.FromJSON(objJSON).ToUnstructured()
		if err != nil {
			return nil, fmt.Errorf("while converting object %s back to unstructured: %v", string(objJSON), err)
		}
		return obj, nil
	}
}

// canApplyPatch determines whether a patch can be applied to an object. It
// checks to ensure that if the patch defines a name,
func canApplyPatch(pat *parsedPatch, obj *unstructured.Unstructured) bool {
	// TODO(kashomon): Use the filter-library logic for this.
	if pat.uns.GetAPIVersion() != "" && pat.uns.GetAPIVersion() != obj.GetAPIVersion() {
		// Patch defined an apiversion , but didn't match the object apiversion.
		return false
	}

	if pat.uns.GetKind() != "" && pat.uns.GetKind() != obj.GetKind() {
		// Patch defined a kind, but didn't match the object kind.
		return false
	}

	if pat.uns.GetName() != "" && pat.uns.GetName() != obj.GetName() {
		// Patch defined a name, but didn't match the object name.
		return false
	}

	return true
}
