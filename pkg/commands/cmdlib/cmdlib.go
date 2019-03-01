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

	log "k8s.io/klog"
	"github.com/spf13/cobra"
)

// ExitWithHelp exits and prints the help message.
func ExitWithHelp(cmd *cobra.Command, err string) {
	log.Error(err)
	cmd.Help()
	os.Exit(1)
}
