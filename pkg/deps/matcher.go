// Copyright 2019 Google LLC
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

package deps

import (
	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
)

// MatchProcessor is a function that's called on every component during the
// creation of the resolver.
type MatchProcessor func(c *bundle.Component) (MatchMetadata, error)

// NoOpMatchProcessor doesn't do any processing
func NoOpMatchProcessor(c *bundle.Component) (MatchMetadata, error) {
	return nil, nil
}

// MatchMetadata is optional metadata stored on the node to be used for matching.
type MatchMetadata interface{}

// Matcher is a function that performs some additional *unconditional* matching
// on a node. True means that the component should be used; false means it
// shouldn't.
//
// In other words, the matching that is applied should be consistent
// across components during the matching phase. If the matching is conditional,
// which is to say relies on some context of the parents or children, then it
// will cause errors in dependency resolution.
type Matcher func(ref bundle.ComponentReference, m MatchMetadata) bool

// NoOpMatcher doesn't do any matching; it returns true (use component) for all components.
func NoOpMatcher(ref bundle.ComponentReference, m MatchMetadata) bool {
	return true
}
