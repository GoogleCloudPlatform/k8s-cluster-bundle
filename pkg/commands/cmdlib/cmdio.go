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

package cmdlib

import (
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/files"
	log "k8s.io/klog"
)

// CmdIO contains dependencies for doing I/O in commands.
type CmdIO struct {
	// StdIO preforms I/O to stdout.
	StdIO StdioReaderWriter

	// FileIO performs I/O to files
	FileIO files.FileReaderWriter

	// ExitIO exits the program with some message.
	ExitIO Exiter
}

// Exiter is a thing that can exit with messages.
type Exiter interface {
	// Exit the program with args.
	Exit(args ...interface{})

	// Exit the program with a formatted message.
	Exitf(format string, v ...interface{})
}

// RealExiter exits the program with some message.
type RealExiter struct{}

// Exit calls log.Exit.
func (e *RealExiter) Exit(args ...interface{}) {
	log.Exit(args...)
}

// Exitf calls log.Exitf.
func (e *RealExiter) Exitf(format string, v ...interface{}) {
	log.Exitf(format, v...)
}
