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

package filter

import (
	"context"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/cmdlib"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/files"
	"github.com/spf13/cobra"
)

// GetCommand filters components or objects from a bundle file
func GetCommand(ctx context.Context, fio files.FileReaderWriter, sio cmdlib.StdioReaderWriter, gopts *cmdlib.GlobalOptions) *cobra.Command{
	opts := &options{}

	cmd := &cobra.Command{
		Use:   "filter",
		Short: "Filter the components or objects in a bundle file",
		Long:  `Filter the components or objects in a bundle file, returning a new bundle file`,
		Run: func(cmd *cobra.Command, args[] string) {
			action(ctx, fio, sio, cmd, opts, gopts)
		},	
	}

	// Optional flags
	cmd.Flags().StringVarP(&opts.filterType, "filter-type", "", "objects", "Whether to filter components or objects")
	cmd.Flags().StringVarP(&opts.kinds, "kinds", "", "", "Comma separated kinds to filter on")
	cmd.Flags().StringVarP(&opts.names, "names", "", "", "Comma separated names to filter on")
	cmd.Flags().StringVarP(&opts.namespaces, "namespaces", "", "", "Comma separated namespaces to filter on")
	cmd.Flags().StringVarP(&opts.annotations, "annotations", "", "", "Comma + semicolon separated annotations to filter on. Ex: 'foo=bar,biff=bam'")
	cmd.Flags().StringVarP(&opts.labels, "labels", "", "", "Comma + semicolon separated labelsto filter on. Ex: 'foo=bar,biff=bam'")
	cmd.Flags().BoolVarP(&opts.keepOnly, "keep-only", "", false, "Whether to keep options instead of filtering them")

	return cmd
}
