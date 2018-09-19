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
	"text/template"

	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/scheme"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/validation"
	log "github.com/golang/glog"
	structpb "github.com/golang/protobuf/ptypes/struct"
	corev1 "k8s.io/api/core/v1"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
)

// Patcher applies patches to objects.
type Patcher struct {
	// Bundle is the bundle from which the Patcher was created.
	Bundle *bpb.ClusterBundle

	// CustomResourceValidator is a validator for CRDs.
	CustomResourceValidator *validation.CustomResourceValidator

	// Scheme wraps the Runtime Kubernetes schema.
	Scheme *scheme.PatcherScheme
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
	for _, app := range bundle.GetSpec().GetComponents() {
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
		Scheme:                  scheme.DefaultPatcherScheme(),
	}, nil
}

// patchRequired returns whether the patch is required or is merely optional.
func (t *Patcher) patchRequired(pat *bpb.Patch) bool {
	if pat.IsRequired {
		return true
	}
	return false
}

// Detemplatize takes a patch message, and with the params specified in the
// patcher, and applies them to the patch template, and returns the patch.
// - Returns an error if the patch doesn't exist
// - Returns an error there aren't sufficient parameters.
func (t *Patcher) Detemplatize(pat *bpb.Patch, customResource interface{}) (*structpb.Struct, error) {
	key := pat.Name
	if key == "" {
		return nil, fmt.Errorf("patch name was empty for patch %v", pat)
	}

	err := t.CustomResourceValidator.Validate(customResource)
	if err != nil {
		return nil, fmt.Errorf("error validating custom resource: %v", err)
	}

	tmpl := pat.GetTemplateString()
	if tmpl == "" {
		return nil, fmt.Errorf("template string was empty for patch %v", pat)
	}

	patchTmpl, err := template.New("patch:" + key).Parse(tmpl)
	if err != nil {
		return nil, fmt.Errorf("for patch %q, template parsing failed: %v ", key, err)
	}
	patchTmpl.Option("missingkey=error")

	var doc bytes.Buffer
	if err := patchTmpl.Execute(&doc, customResource); err != nil {
		if t.patchRequired(pat) {
			return nil, fmt.Errorf("for patch %q, error applying template params: %v", key, err)
		}
		log.Infof("Missing parameters in patch , but patch is optional: %v. not applying patch: %v", err, pat.Name)
		return nil, nil
	}

	pb, err := converter.Struct.YAMLToProto(doc.Bytes())
	if err != nil {
		return nil, fmt.Errorf("for patch %q, error parsing YAML: %v", key, err)
	}
	return converter.ToStruct(pb), nil
}

// ApplyToClusterObjects applies patches  to cluster objects, always returning a
// new object copy of the original object with the patches (if any) applied.
//
// Note: every cluster object must have it's type registered in with the K8S
// scheme / codec factory.
func (t *Patcher) ApplyToClusterObjects(patches []*bpb.Patch, customResource interface{}, kubeObj *structpb.Struct) (*structpb.Struct, error) {
	objJSON, err := converter.Struct.ProtoToJSON(kubeObj)
	if err != nil {
		return nil, err
	}
	for _, pat := range patches {
		p, err := t.Detemplatize(pat, customResource)
		if err != nil {
			return nil, err
		}
		if p == nil {
			// No patch to apply for whatever reason. Continue on with applying other patches.
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
			return nil, fmt.Errorf("for patch %q, error applying patch: %v", pat.Name, err)
		}
	}
	o, err := converter.Struct.JSONToProto(objJSON)
	return converter.ToStruct(o), err
}

// PatchComponent applies patches to all the cluster objects in a given ClusterComponent,
// always returning a new component copy of the original component with the patches (if any)
// applied.
func (t *Patcher) PatchComponent(comp *bpb.ClusterComponent, customResources []map[string]interface{}) (*bpb.ClusterComponent, error) {
	crMap, err := converter.KubeResourceMap(customResources)
	if err != nil {
		return nil, fmt.Errorf("could not patch component%q: %v", comp.GetName(), err)
	}
	return t.patchComponent(comp, crMap)
}

// patchComponent applies patches to all the cluster objects in a given ClusterComponent,
// always returning a new component copy of the original component with the patches (if any)
// applied. It takes in a map from custom resource kind to custom resource instance for ease of
// lookup.
func (t *Patcher) patchComponent(comp *bpb.ClusterComponent, crMap map[corev1.ObjectReference]interface{}) (*bpb.ClusterComponent, error) {
	clonedComponent := converter.CloneClusterComponent(comp)
	for _, co := range clonedComponent.GetClusterObjects() {
		obj := co.GetInlined()
		if obj == nil {
			return nil, fmt.Errorf("object %q was not inlined", co.GetName())
		}
		for _, pat := range co.GetPatchCollection().GetPatches() {
			var err error
			// Pass each patch separately to update the cluster object along with the corresponding CR.
			cr, found := crMap[converter.ToObjectReference(pat.GetObjectRef())]
			if !found {
				return nil, fmt.Errorf("could not patch object %q: no custom resource found for %q", co.GetName(), pat.GetObjectRef())
			}
			obj, err = t.ApplyToClusterObjects([]*bpb.Patch{pat}, cr, obj)
			if err != nil {
				return nil, fmt.Errorf("could not patch object %q: %s", co.GetName(), err)
			}
		}
		co.KubeData = &bpb.ClusterObject_Inlined{obj}
	}
	return clonedComponent, nil
}

// PatchBundle applies patches to all the cluster objects in the Patcher
// Bundle. It returns a new bundle copy of the original bundle with the patches
// (if any) applied.
func (t *Patcher) PatchBundle(customResources []map[string]interface{}) (*bpb.ClusterBundle, error) {
	bundle := converter.CloneBundle(t.Bundle)
	comps := bundle.GetSpec().GetComponents()[:]

	crMap, err := converter.KubeResourceMap(customResources)
	if err != nil {
		return nil, fmt.Errorf("could not patch bundle: %s", err)
	}

	for i, comp := range t.Bundle.Spec.GetComponents() {
		patched, err := t.patchComponent(comp, crMap)
		if err != nil {
			return nil, fmt.Errorf("could not patch component %q: %s", comp.GetName(), err)
		}
		// Replace each component in the bundle with the patched component.
		comps[i] = patched
	}
	bundle.GetSpec().Components = comps
	return bundle, nil
}
