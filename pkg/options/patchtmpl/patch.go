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
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/filter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/options"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/options/openapi"
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

	// If includeTemplates is true, applied patch templates will be included in the
	// component objects.
	includeTemplates bool

	// Set the template options
	templateOpts []string
}

// NewApplier creates a new options applier instance. The filter indicates
// keep-only options for what subsets of patches to look for.
func NewApplier(pt *PatcherScheme, opts *filter.Options, includeTemplates bool, templateOpts ...string) options.Applier {
	if templateOpts == nil || len(templateOpts) == 0 {
		templateOpts = []string{options.MissingKeyError}
	}

	return &applier{
		scheme:           pt,
		tmplFilter:       opts,
		includeTemplates: includeTemplates,
		templateOpts:     templateOpts,
	}
}

// NewDefaultApplier creates a default patch template applier.
func NewDefaultApplier() options.Applier {
	return NewApplier(DefaultPatcherScheme(), nil, false)
}

// ApplyOptions looks for PatchTemplates and applies them to the component objects.
func (a *applier) ApplyOptions(comp *bundle.Component, p options.JSONOptions) (*bundle.Component, error) {
	comp = comp.DeepCopy()
	patchTemplates, objects := a.getPatchTemplates(comp)
	if len(patchTemplates) < 1 {
		return comp, nil
	}
	patches, objs, err := a.makePatches(patchTemplates, objects, p)
	if err != nil {
		return nil, err
	}
	newObjs, err := options.ApplyCommon(comp.ComponentReference(), objs, p, objectApplier(a.scheme, patches))
	comp.Spec.Objects = newObjs
	return comp, err
}

// A parsedPatch has had options applied and has been converted both into raw
// bytes and unstructured.
type parsedPatch struct {
	// raw parsed patch.
	raw []byte

	// the parsed patch as a JSON Map
	jsonMap map[string]interface{}

	// selector for determining which objects to patch.
	selector *bundle.ObjectSelector

	// type of merging logic to use for the patch
	patchType bundle.PatchType
}

// String returns the string form of the parsedPatch.
func (p *parsedPatch) String() string {
	return string(p.raw)
}

// getPatchTemplates returns all patch templates fom particular component
func (a *applier) getPatchTemplates(comp *bundle.Component) ([]*unstructured.Unstructured, []*unstructured.Unstructured) {
	tfil := a.tmplFilter
	if tfil == nil {
		tfil = &filter.Options{}
	}
	tfil.Kinds = []string{"PatchTemplate"}

	// Filter all the objects to include just the patch templates + any additional values.
	ptObjs, objs := filter.NewFilter().PartitionObjects(comp.Spec.Objects, tfil)

	// PartitionObjects will exclude *matching* patch templates from objs; if we actually
	// want to include them, then use the original component objects for our object list
	if a.includeTemplates {
		objs = comp.Spec.Objects
	}

	return ptObjs, objs
}

func (a *applier) makePatches(ptObjs, objs []*unstructured.Unstructured, opts options.JSONOptions) ([]*parsedPatch, []*unstructured.Unstructured, error) {
	// First parse the objects back into go-objects.
	var pts []*bundle.PatchTemplate
	for _, o := range ptObjs {
		pto := &bundle.PatchTemplate{}
		err := converter.FromUnstructured(o).ToObject(pto)
		if err != nil {
			return nil, nil, fmt.Errorf("while converting object %v to PatchTemplate: %v", pto, err)
		}
		pts = append(pts, pto)
	}

	if opts == nil {
		opts = options.JSONOptions{}
	}

	// Next, de-templatize the templates.
	var patches []*parsedPatch
	for j, pto := range pts {
		patchType := bundle.PatchType(pto.PatchType)
		switch patchType {
		case bundle.StrategicMergePatch, bundle.JSONPatch:
			// known types

		case "":
			// use default
			patchType = bundle.StrategicMergePatch

		default:
			return nil, nil, fmt.Errorf("bad patch type: %s", patchType)
		}

		tmpl, err := template.New(fmt.Sprintf("patch-tmpl-%d", j)).Parse(pto.Template)
		if err != nil {
			return nil, nil, fmt.Errorf("parsing patch template %d, %s: %v", j, pto.Template, err)
		}

		tmpl = tmpl.Option(a.templateOpts...)

		newOpts := opts
		if pto.OptionsSchema != nil {
			newOpts, err = openapi.ApplyDefaults(opts, pto.OptionsSchema)
			if err != nil {
				return nil, nil, fmt.Errorf("applying schema defaults for patch template %d, %s: %v", j, pto.Template, err)
			}
		}

		// Detemplatize the patch
		var buf bytes.Buffer
		err = tmpl.Execute(&buf, newOpts)
		if err != nil {
			return nil, nil, fmt.Errorf("while applying options to patch template %d: %v", j, err)
		}

		// Convert the patch into a JSONMap to prepare for Strategic Merge Patch.
		by := buf.Bytes()
		jsonMap := make(map[string]interface{})
		err = converter.FromYAML(by).ToObject(&jsonMap)
		if err != nil {
			return nil, nil, fmt.Errorf("while converting patch template %d: %v", j, err)
		}

		// Neither Kind nor APIVersion are allowed as patchable fields in a
		// PatchTemplate -- we don't want to change the schema of the objects we're
		// patching. So, instead remove them from the PatchTemplate and add them as
		// an additional selector parameter (which supports the previous behavior).
		pKind := ""
		pAPIVersion := ""
		if jsonMap["kind"] != nil {
			var ok bool
			pKind, ok = jsonMap["kind"].(string)
			if !ok {
				return nil, nil, fmt.Errorf("found non-string type %T for Kind field for patch %s", jsonMap["kind"], string(by))
			}
			delete(jsonMap, "kind")
		}
		if jsonMap["apiVersion"] != nil {
			var ok bool
			pAPIVersion, ok = jsonMap["apiVersion"].(string)
			if !ok {
				return nil, nil, fmt.Errorf("found a non-string APIVersion field for patch %s", string(by))
			}
			delete(jsonMap, "apiVersion")
		}

		selector := pto.Selector
		if pKind != "" {
			if selector == nil {
				selector = &bundle.ObjectSelector{}
			}
			if pAPIVersion != "" {
				pKind = pAPIVersion + "," + pKind
			}
			selector.Kinds = append(selector.Kinds, pKind)
		}

		patches = append(patches, &parsedPatch{
			raw:       by,
			jsonMap:   jsonMap,
			selector:  selector,
			patchType: patchType,
		})
	}
	return patches, objs, nil
}

// objectApplier creates a patch object-handler. For each patch, the object
// applier function checks whether a patch can be applied, and if so, then
// applies it.
func objectApplier(scheme *PatcherScheme, patches []*parsedPatch) options.ObjHandler {
	return func(obj *unstructured.Unstructured, ref bundle.ComponentReference, _ options.JSONOptions) ([]*unstructured.Unstructured, error) {
		objJSON := obj.Object

		if obj.GetKind() == "PatchTemplate" {
			// Don't process PatchTemplates (possible if includeTemplates is set). In
			// other words, it's not allowed to apply patch templates to patch
			// templates.
			return []*unstructured.Unstructured{obj}, nil
		}

		// TODO(kashomon): Is there a faster way to convert from JSON-Map to string?
		objByt, err := converter.FromObject(objJSON).ToJSON()
		if err != nil {
			// This would be pretty unlikely
			return nil, err
		}

		deserializer := scheme.Codecs.UniversalDeserializer()

		kubeObj, decodeErr := runtime.Decode(deserializer, objByt)
		_, isUnstructured := kubeObj.(*unstructured.Unstructured)
		strategicWillFail := runtime.IsNotRegisteredError(decodeErr) || isUnstructured
		objSchema, objSchemaErr := strategicpatch.NewPatchMetaFromStruct(kubeObj)
		for _, pat := range patches {
			if !canApplyPatch(pat, obj) {
				continue
			}

			var newObjJSON map[string]interface{}
			switch pat.patchType {
			case bundle.JSONPatch:
				if oByt, err := converter.FromObject(objJSON).ToJSON(); err != nil {
					return nil, fmt.Errorf("while converting JSON obj\n%s to bytes: %v", objJSON, err)
				} else if pByt, err := converter.FromObject(pat.jsonMap).ToJSON(); err != nil {
					return nil, fmt.Errorf("while converting patch JSON obj\n%s to bytes: %v", pat.jsonMap, err)
				} else if newObjByt, err := jsonpatch.MergePatch(oByt, pByt); err != nil {
					return nil, fmt.Errorf("while applying JSON merge patch\n%s to \n%s: %v", pat.raw, oByt, err)
				} else if newObjJSON, err = converter.FromJSON(newObjByt).ToJSONMap(); err != nil {
					return nil, fmt.Errorf("while converting bytes\n%s to JSON: %v", newObjByt, err)
				}

			case bundle.StrategicMergePatch:
				if strategicWillFail {
					// Strategic merge patch can't handle unstructured.Unstructured or
					// unregistered objects, so return an error.
					return nil, fmt.Errorf("while converting object %q of kind %q and apiVersion %q: type not registered in scheme", obj.GetName(), obj.GetKind(), obj.GetAPIVersion())
				}
				if objSchemaErr != nil {
					return nil, fmt.Errorf("while getting patch meta from object %s: %v", string(objByt), objSchemaErr)
				}
				if newObjJSON, err = strategicpatch.StrategicMergeMapPatchUsingLookupPatchMeta(objJSON, pat.jsonMap, objSchema); err != nil {
					return nil, fmt.Errorf("while applying strategic merge patch\n%sto \n%s: %v", pat.raw, objJSON, err)
				}

			default:
				return nil, fmt.Errorf("unknown patch type: %s", pat.patchType)
			}

			objJSON = newObjJSON
		}

		obj = &unstructured.Unstructured{
			Object: runtime.DeepCopyJSON(objJSON),
		}
		return []*unstructured.Unstructured{obj}, nil
	}
}

// canApplyPatch determines whether a patch can be applied to an object. It
// checks to ensure that if the patch defines a name,
func canApplyPatch(pat *parsedPatch, obj *unstructured.Unstructured) bool {
	return filter.MatchesObject(obj, filter.OptionsFromObjectSelector(pat.selector))
}
