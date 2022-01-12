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

// Package wrapper provides a union type for expressing various different
// bundle-types.
package wrapper

import (
	"fmt"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/options"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/options/multi"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// BundleWrapper represents one of several possible bundle types -- a union type.
type BundleWrapper struct {
	bundle    *bundle.Bundle
	component *bundle.Component

	bundleBuilder    *bundle.BundleBuilder
	componentBuilder *bundle.ComponentBuilder
}

// FromRaw creates a BundleWrapper object from a byte source, by converting
// based off of the Kind.
func FromRaw(inFmt string, bytes []byte) (*BundleWrapper, error) {
	if len(bytes) == 0 {
		return nil, fmt.Errorf("in FromRaw, content was empty")
	}
	if inFmt == "" {
		return nil, fmt.Errorf("in FromRaw, format was empty")
	}

	uns, err := converter.FromContentType(inFmt, bytes).ToUnstructured()
	if err != nil {
		return nil, err
	}

	kind := uns.GetKind()
	switch kind {
	case "BundleBuilder":
		b, err := converter.FromContentType(inFmt, bytes).ToBundleBuilder()
		if err != nil {
			return nil, err
		}
		return FromBundleBuilder(b), nil
	case "Bundle":
		b, err := converter.FromContentType(inFmt, bytes).ToBundle()
		if err != nil {
			return nil, err
		}
		return FromBundle(b), nil
	case "ComponentBuilder":
		c, err := converter.FromContentType(inFmt, bytes).ToComponentBuilder()
		if err != nil {
			return nil, err
		}
		return FromComponentBuilder(c), nil
	case "Component":
		c, err := converter.FromContentType(inFmt, bytes).ToComponent()
		if err != nil {
			return nil, err
		}
		return FromComponent(c), nil
	default:
		return nil, fmt.Errorf("unrecognized bundle-kind %s for content %s", kind, string(bytes))
	}
}

// FromBundle creates a BundleWrapper from a bundle object.
func FromBundle(c *bundle.Bundle) *BundleWrapper {
	return &BundleWrapper{bundle: c}
}

// FromComponent creates a BundleWrapper from a component object
func FromComponent(c *bundle.Component) *BundleWrapper {
	return &BundleWrapper{component: c}
}

// FromBundleBuilder creates a BundleWrapper from a bundle builder object.
func FromBundleBuilder(c *bundle.BundleBuilder) *BundleWrapper {
	return &BundleWrapper{bundleBuilder: c}
}

// FromComponentBuilder creates a BundleWrapper from a component builder object
func FromComponentBuilder(c *bundle.ComponentBuilder) *BundleWrapper {
	return &BundleWrapper{componentBuilder: c}
}

// Bundle returns the wrapped bundle object.
func (bw *BundleWrapper) Bundle() *bundle.Bundle { return bw.bundle }

// Component returns the wrapped Component object.
func (bw *BundleWrapper) Component() *bundle.Component { return bw.component }

// BundleBuilder returns the wrapped BundleBuilder object.
func (bw *BundleWrapper) BundleBuilder() *bundle.BundleBuilder { return bw.bundleBuilder }

// ComponentBuilder returns the wrapped ComponentBuilder object.
func (bw *BundleWrapper) ComponentBuilder() *bundle.ComponentBuilder { return bw.componentBuilder }

// Kind gets the type of the underlying object.
func (bw *BundleWrapper) Kind() string {
	if bw.Bundle() != nil {
		return bw.Bundle().Kind
	} else if bw.BundleBuilder() != nil {
		return bw.BundleBuilder().Kind
	} else if bw.Component() != nil {
		return bw.Component().Kind
	} else if bw.ComponentBuilder() != nil {
		return bw.ComponentBuilder().Kind
	}
	return ""
}

// Object returns the wrapped object.
func (bw *BundleWrapper) Object() interface{} {
	if bw.Bundle() != nil {
		return bw.Bundle()
	} else if bw.BundleBuilder() != nil {
		return bw.BundleBuilder()
	} else if bw.Component() != nil {
		return bw.Component()
	} else if bw.ComponentBuilder() != nil {
		return bw.ComponentBuilder()
	}
	return nil
}

// AllComponents returns all the components in the BundleWrapper, with the
// assumption that only one of Bundle or Component is filled out.
func (bw *BundleWrapper) AllComponents() []*bundle.Component {
	if bw.bundle != nil {
		return bw.bundle.Components
	} else if bw.component != nil {
		return []*bundle.Component{bw.component}
	}
	return nil
}

// ExportAsObjects will export a Bundle as a ComponentSet and Components,
// or a Component as unstructured Objects.
func (bw *BundleWrapper) ExportAsObjects(opts options.JSONOptions) ([]*unstructured.Unstructured, error) {
	switch bw.Kind() {
	case "Component":
		components := bw.Component()
		if opts != nil {
			applier := multi.NewDefaultApplier()
			comp, err := applier.ApplyOptions(components, opts)
			if err != nil {
				return nil, fmt.Errorf("failed to apply options: %v", err)
			}
			components = comp
		}
		return components.Spec.Objects, nil
	case "Bundle":
		bun := bw.Bundle()
		y, err := converter.FromObject(bun.ComponentSet()).ToYAML()
		if err != nil {
			return nil, err
		}
		o, err := converter.FromYAML(y).ToUnstructured()
		if err != nil {
			return nil, err
		}
		var objs []*unstructured.Unstructured
		objs = append(objs, o)
		for _, c := range bun.Components {
			y, err := converter.FromObject(c).ToYAML()
			if err != nil {
				return nil, err
			}
			o, err := converter.FromYAML(y).ToUnstructured()
			if err != nil {
				return nil, err
			}
			objs = append(objs, o)
		}
		return objs, nil
	default:
		return nil, fmt.Errorf("bundle kind %q not supported for exporting", bw.Kind())
	}
}
