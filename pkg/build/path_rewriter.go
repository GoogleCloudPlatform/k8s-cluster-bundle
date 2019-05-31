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
	"path"
	"path/filepath"
)

// makeAbsWithParent rewrites file paths if the path is relative, from the
// original object path, to an object path that's based on a parent component
// path. Note that the parent path must be an absolute path.
//
// For example, if the component path is foo/bar/biff.yaml and the object path
// is zed/fred.yaml, the object will be rewritten as foo/bar/zed/fred.yaml
func makeAbsWithParent(parent, obj *url.URL) *url.URL {
	if parent == nil || obj == nil {
		return obj
	}
	if path.IsAbs(obj.Path) {
		return obj
	}
	return &url.URL{
		Scheme: parent.Scheme,
		Host:   parent.Host,
		Path:   path.Clean(path.Join(path.Dir(parent.Path), obj.Path)),
	}
}

// makeAbsForFileScheme makes an absolute url for URL that has an empty or
// file-based schemes.
func makeAbsForFileScheme(obj *url.URL) (*url.URL, error) {
	if (obj.Scheme == "file" || obj.Scheme == "") && !filepath.IsAbs(obj.Path) {
		s, err := filepath.Abs(obj.Path)
		if err != nil {
			return nil, err
		}
		return &url.URL{
			Scheme: obj.Scheme,
			Host:   obj.Host,
			Path:   s,
		}, nil
	}
	return obj, nil
}
