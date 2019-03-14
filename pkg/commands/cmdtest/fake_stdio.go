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

package cmdtest

// FakeStdioReaderWriter is a fake implementation of a reader/writer to
// stdout.
type FakeStdioReaderWriter struct {
	// ReadBytes contains the bytes to be read.
	ReadBytes []byte

	// ReadErr is an error that occurs during Read. If ReadErr is present, then
	// ReadErr will be returned instead of ReadBytes
	ReadErr error

	// WriteBytes contains a record of the most-recently written bytes.
	WriteBytes []byte

	// WriteErr is an error that occurs during Write. If WriteErr is present, then
	// Write will write not write content and return WriteErr instead.
	WriteErr error
}

// ReadAll returns all the ReadBytes or returns an error.
func (f *FakeStdioReaderWriter) ReadAll() ([]byte, error) {
	if f.ReadErr != nil {
		return nil, f.ReadErr
	}
	return f.ReadBytes, nil
}

// Write stores all the WriteBytes or returns an error.
func (f *FakeStdioReaderWriter) Write(b []byte) (int, error) {
	if f.WriteErr != nil {
		return 0, f.WriteErr
	}
	f.WriteBytes = b
	return len(f.WriteBytes), nil
}
