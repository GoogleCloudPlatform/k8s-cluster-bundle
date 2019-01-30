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

// Package partial patch provides functionality for applying a subset of options to patches.
package partialpatch

type applier struct{}

// ApplyOptions looks for PatchTemplates and applies options to the
// PatchTemplates, returning components with the partially filled out
// PatchTemplates, but not actually applying them.
func (a *applier) ApplyOptions(comp *bundle.Component, p options.JSONOptions) (*bundle.Component, error) {
	return nil, fmt.Errorf("not implemented")
}
