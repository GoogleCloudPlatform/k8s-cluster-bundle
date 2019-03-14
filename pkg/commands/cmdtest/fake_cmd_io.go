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

// Package cmdtest provides test utilities for testing commands.
package cmdtest

import (
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/testutil"
)

// FakeCmdIO contains fake implementations of the CmdIo dependencies
type FakeCmdIO struct {
	StdIO  *FakeStdioReaderWriter
	FileIO *testutil.FakeFileReaderWriter
	ExitIO *FakeExiter
}

// NewFakeCmdIO creates a new FakeCmdIO
func NewFakeCmdIO() *FakeCmdIO {
	return &FakeCmdIO{
		StdIO:  &FakeStdioReaderWriter{},
		FileIO: testutil.NewEmptyReaderWriter(),
		ExitIO: &FakeExiter{},
	}
}
