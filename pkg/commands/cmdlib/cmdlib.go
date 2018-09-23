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

package cmdlib

import (
	"os"

	"context"
	log "github.com/golang/glog"
	"github.com/spf13/cobra"
)

func ExitWithHelp(cmd *cobra.Command, err string) {
	log.Error(err)
	cmd.Help()
	os.Exit(1)
}

// DefaultFilePermissions is the permission used for the bundle output
// file (rw-r--r--).
const DefaultFilePermissions = os.FileMode(0644)

// ContextActionFunc is a common type for providing a context to a Cobra function.
type ContextActionFunc func(ctx context.Context, cmd *cobra.Command, args []string)

// CobraActionFunc provides a common type for all Cobra commands.
type CobraActionFunc func(cmd *cobra.Command, args []string)

// ContextAction returns a CobraActionFunc for a provided ContextActionFunc.
func ContextAction(ctx context.Context, f ContextActionFunc) CobraActionFunc {
	return func(cmd *cobra.Command, args []string) {
		f(ctx, cmd, args)
	}
}
