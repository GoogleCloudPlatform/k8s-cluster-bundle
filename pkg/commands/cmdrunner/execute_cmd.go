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

// Package cmdrunner is a utility for running integration tests for commands.
package cmdrunner

import (
	"context"
	"flag"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/cmdlib"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/cmdtest"
)

// ExecuteCommand executes a fake command.
func ExecuteCommand(fakeio *cmdtest.FakeCmdIO, args []string) error {
	ctx := context.Background()

	cmdio := &cmdlib.CmdIO{
		StdIO:  fakeio.StdIO,
		FileIO: fakeio.FileIO,
		ExitIO: fakeio.ExitIO,
	}

	flagset := flag.NewFlagSet("test-flagset", flag.ContinueOnError)

	cmd := commands.AddCommandsInternal(ctx, cmdio, flagset, args)
	return cmd.Execute()
}
