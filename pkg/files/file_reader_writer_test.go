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

package files

import "testing"

func TestExtractPath(t *testing.T) {
	tests := []struct {
		workDir string
		fileURL string

		wantErr  bool
		wantPath string
	}{
		{
			workDir:  "examples",
			fileURL:  "./foo",
			wantPath: "examples/foo",
		},
		{
			workDir:  "examples",
			fileURL:  "../foo",
			wantPath: "foo",
		},
		{
			workDir:  "examples",
			fileURL:  "/foo",
			wantPath: "/foo",
		},
		{
			workDir:  "examples",
			fileURL:  "file:///foo",
			wantPath: "examples/foo",
		},
		{
			workDir: "examples",
			fileURL: "file://./foo",
			wantErr: true,
		},
		{
			workDir: "examples",
			fileURL: "https://google.com/foo",
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			r := &LocalFileObjReader{WorkingDir: test.workDir}
			path, err := r.extractPath(test.fileURL)
			if wantErr, gotErr := test.wantErr, err != nil; wantErr != gotErr {
				t.Fatalf("unexpected err: %v", err)
			}
			if want, got := test.wantPath, path; want != got {
				t.Fatalf("unexpected path: want=%q, got=%q", want, got)
			}
		})
	}
}
