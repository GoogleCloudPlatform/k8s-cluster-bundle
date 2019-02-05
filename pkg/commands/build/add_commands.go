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

package build

import (
	"context"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/cmdlib"
	"github.com/spf13/cobra"
)

// AddCommandsTo adds commands to a root cobra command.
func AddCommandsTo(ctx context.Context, root *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build the bundle files",
		Long:  `Build all the files in the given bundle yaml`,
		Run:   cmdlib.ContextAction(ctx, action),
	}

	// Optional flags

	// While options-file is technically optional, it is usually provided to
	// detemplatize the patch templates.
	cmd.Flags().StringVarP(&opts.optionsFile, "options-file", "", "",
		"File containing options to apply to patch templates")

	cmd.Flags().StringVarP(&opts.annotations, "annotations", "", "",
		"Select a subset of patch templates to build to apply based on a list of annotations of the form \"key1,val1;key2,val2;\"")

	root.AddCommand(cmd)
}
