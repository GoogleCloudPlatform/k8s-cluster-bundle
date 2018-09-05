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

package commands

import (
	"os"
	"path/filepath"

	"context"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/transformer"
	log "github.com/golang/glog"
	"github.com/spf13/cobra"
)

// outFilePermissions is the permission used for the bundle output file (rw-r--r--).
const outFilePermissions = os.FileMode(0644)

type inlineOptions struct {
	bundle string
	output string
}

var inlineOpts = &inlineOptions{}

// addInlineCommand sets up the inline command with context and adds it to the root command.
func addInlineCommand(ctx context.Context) {
	inlineCmd := &cobra.Command{
		Use:   "inline",
		Short: "Inline the bundle files",
		Long:  `Inline all the files in the given bundle yaml`,
		Run:   ContextAction(ctx, inlineAction),
	}

	// Required flags
	// Note: the path to the bundle must be absolute when running with bazel.
	inlineCmd.Flags().StringVarP(&inlineOpts.bundle, "bundle", "b", "",
		"The path to the bundle to inline")

	// Optional flags
	// Note: the path to the output bundle is required when running with bazel,
	// and it must be absolute.
	inlineCmd.Flags().StringVarP(&inlineOpts.output, "output", "o", "bundle_inline.yaml",
		"Where to output the inline bundle. By default it outputs to the current working directory")

	rootCmd.AddCommand(inlineCmd)
}

func inlineAction(ctx context.Context, cmd *cobra.Command, _ []string) {
	if inlineOpts.bundle == "" {
		exitWithHelp(cmd, "Please provide yaml file for bundle.")
	}
	if err := runInline(ctx, inlineOpts, &realReaderWriter{}); err != nil {
		log.Exit(err)
	}
}

// createInlinerFn creates an Inliner that works with the given current working directory.
var createInlinerFn = func(cwd string) *transformer.Inliner {
	return transformer.NewInliner(cwd)
}

func runInline(ctx context.Context, opts *inlineOptions, brw BundleReaderWriter) error {
	path, err := filepath.Abs(opts.bundle)
	if err != nil {
		return err
	}

	b, err := brw.ReadBundleFile(path)
	if err != nil {
		return err
	}

	inliner := createInlinerFn(filepath.Dir(path))
	inlined, err := inliner.Inline(ctx, b)
	if err != nil {
		return err
	}
	return brw.WriteBundleFile(opts.output, inlined, outFilePermissions)
}
