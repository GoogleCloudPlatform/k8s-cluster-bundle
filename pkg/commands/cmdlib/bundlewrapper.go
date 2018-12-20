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

package cmdlib

import (
	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
)

// BundleWrapper represents one of several possible bundle types -- a union type.
type BundleWrapper struct {
	bundle    *bundle.Bundle
	component *bundle.Component

	bundleBuilder    *bundle.BundleBuilder
	componentBuilder *bundle.ComponentBuilder
}

// NewWrapper creates a WrapperConstructor.
func NewWrapper() *WrapperConstructor {
	return &WrapperConstructor{}
}

// WrapperConstructor provides syntactic sugar for creating a bundle wrapper.
type WrapperConstructor struct{}

// FromBundle creates a BundleWrapper from a bundle object.
func (*WrapperConstructor) FromBundle(c *bundle.Bundle) *BundleWrapper {
	return &BundleWrapper{bundle: c}
}

// FromComponent creates a BundleWrapper from a component object
func (*WrapperConstructor) FromComponent(c *bundle.Component) *BundleWrapper {
	return &BundleWrapper{component: c}
}

// FromBundleBuilder creates a BundleWrapper from a bundle builder object.
func (*WrapperConstructor) FromBundleBuilder(c *bundle.BundleBuilder) *BundleWrapper {
	return &BundleWrapper{bundleBuilder: c}
}

// FromComponentBuilder creates a BundleWrapper from a component builder object
func (*WrapperConstructor) FromComponentBuilder(c *bundle.ComponentBuilder) *BundleWrapper {
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
