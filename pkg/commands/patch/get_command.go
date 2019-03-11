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

// Package patch contains commands for patching objects.
package patch

import (
	"context"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/cmdlib"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/files"
	"github.com/spf13/cobra"
)

// GetCommand returns the command for patching.
func GetCommand(ctx context.Context, fio files.FileReaderWriter, sio cmdlib.StdioReaderWriter, gopts *cmdlib.GlobalOptions) *cobra.Command {
	opts := &options{}
	cmd := &cobra.Command{
		Use:   "patch",
		Short: "Apply patch templates to component objects",
		Long: "Apply patch templates to component objects. " +
			"Options are usually applied to the templates before application.",
		Run: func(cmd *cobra.Command, args []string) {
			action(ctx, fio, sio, cmd, opts, gopts)
		},
	}
	// While options-file is technically optional, it is usually provided to detemplatize the patch templates.
	cmd.Flags().StringArrayVar(&opts.optionsFiles, "options-file", []string{}, "File containing options to apply to patch templates. May be repeated, later values override earlier ones.")
	cmd.Flags().StringVar(&opts.patchAnnotations, "patch-annotations", "", "Select a subset of patches to apply based on a list of annotations of the form \"key1=val1,key2=val2\"")
	cmd.Flags().BoolVar(&opts.keepTemplates, "keep-templates", false, "Do not remove templates that have been applied from the component.")
	return cmd
}
