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

// Below is a worked example of a matcher using the annotations on the
// component itself.

// AnnotationMetadata contains annotation metadata, used for matching against
// components.
type AnnotationMetadata struct {
	Annotations map[string]string
}

// AnnotationProcessor extracts annotations from a component, to be used for Matching
func AnnotationProcessor(c *bundle.Component) (MatchMetadata, error) {
	am := &AnnotationMetadata{
		Annotations: make(map[string]string),
	}
	for key, val := range c.ObjectMeta.Annotations {
		am.Annotations[key] = val
	}
	return am, nil
}

// AnnotationCriteria are options for resolving dependencies
type AnnotationCriteria struct {
	// Match are component annotations that must to be present. This is useful
	// for matching positive characteristics of a component (example: this
	// component passed qualification). Only one of the list of values need be
	// present. In other words, given values A, B, C for Annotation K, the logic
	// is equivalent to A || B || C).
	//
	// If there are multiple keys (annotations) specified, then all annotations
	// must match.  In other words, this is a logical AND operation of the passed
	// in Annotation and the component Annotation.
	Match map[string][]string

	// Exclude are component annotations that must not be present.  This is
	// useful for matching positive characteristics of a component (this
	// component passed qualification). If any of the the list of values match,
	// the component is excluded.
	//
	// Unlike Match, if a there are multiple keys (annotations) specified, then
	// only one of the annotations need be matched in order for the component to
	// be excluded.
	Exclude map[string][]string
}

// AnnotationMatcher constructs a Matcher that relies on a components annotations.
func AnnotationMatcher(criteria *AnnotationCriteria) Matcher {
	return func(ref bundle.ComponentReference, mm MatchMetadata) bool {
		if mm == nil {
			return true
		}

		m, ok := mm.(*AnnotationMetadata)
		if !ok {
			return true
		}

		matchesAnnot := func(key string, vals []string, annots map[string]string) bool {
			matchesAny := false
			for _, val := range vals {
				if m.Annotations[key] == val {
					matchesAny = true
					break
				}
			}
			return matchesAny
		}

		match := true
		for key, vals := range criteria.Match {
			if !matchesAnnot(key, vals, m.Annotations) {
				match = false
				break
			}
		}

		exclude := false
		for key, vals := range criteria.Exclude {
			if matchesAnnot(key, vals, m.Annotations) {
				exclude = true
				break
			}
		}
		return match && !exclude
	}
}

// sliceContains is a helper to indicate whethera slice of strings contains a string.
func sliceContains(ls []string, st string) bool {
	for _, s := range ls {
		if s == st {
			return true
		}
	}
	return false
}
