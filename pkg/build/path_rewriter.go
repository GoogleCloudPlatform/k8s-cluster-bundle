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

package build

import (
	"net/url"
	"path/filepath"
)

// RelativePathRewriter rewrites paths when the paths are relative paths.
type RelativePathRewriter struct{}

// MakeAbs rewrites file paths if the path is relative, from the
// original object path, to an object path that's based on a parent component
// path. Note that the parent path must be an absolute path.
//
// For example, if the component path is foo/bar/biff.yaml and the object path
// is zed/fred.yaml, the object will be rewritten as foo/bar/zed/fred.yaml
func (rw *RelativePathRewriter) MakeAbs(parent, obj *url.URL) *url.URL {
	if parent == nil || obj == nil {
		return obj
	}
	if parent.Scheme != "file" && parent.Scheme != "" {
		// Only file schemes are supported.
		return obj
	}
	if !filepath.IsAbs(parent.Path) {
		return obj
	}
	if filepath.IsAbs(obj.Path) {
		return obj
	}
	return &url.URL{
		Scheme: parent.Scheme,
		Host:   parent.Host,
		Path:   filepath.Clean(filepath.Join(filepath.Dir(parent.Path), obj.Path)),
	}
}

// DefaultPathRewriter provides a RelativePathRewriter instance.
var DefaultPathRewriter = &RelativePathRewriter{}
