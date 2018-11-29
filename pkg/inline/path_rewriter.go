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

package inline

import (
	"net/url"
	"path/filepath"
)

// TODO(kashomon): Do we really need this? Maybe we can just say that relative
// paths are not supported and ignore the entire problem of path rewriting.

// PathRewriter can rewrite URL paths.
type PathRewriter interface {
	// Rewrite URL paths for objects.
	RewriteObjectPath(componentURL, objectURL *url.URL) string
}

// RelativePathRewriter rewrites paths when the paths are relative paths.
type RelativePathRewriter struct{}

func (rw *RelativePathRewriter) RewriteObjectPath(comp, obj *url.URL) string {
	if filepath.IsAbs(obj.Path) {
		// The rewriter only needs to rewrite relative paths.
		// This is a path of the form file:///, and so is not a relative path.
		return obj.String()
	}
	return "file://" + filepath.Join(filepath.Dir(comp.Path), obj.Path)
}

// DefaultPathRewriter provides a RelativePathRewriter instance.
var DefaultPathRewriter = &RelativePathRewriter{}
