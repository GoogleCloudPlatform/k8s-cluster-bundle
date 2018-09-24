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

package inline

import (
	"context"
	"path/filepath"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/cmdlib"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/core"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/transformer"
	log "github.com/golang/glog"
	"github.com/spf13/cobra"
)

// options represents options flags for the inline command.
type options struct {
	bundle string
	output string
}

// opts is a global options instance for reference via the add commands.
var opts = &options{}

func action(ctx context.Context, cmd *cobra.Command, _ []string) {
	if opts.bundle == "" {
		cmdlib.ExitWithHelp(cmd, "Please provide yaml file for bundle.")
	}

	brw := converter.NewFileSystemBundleReaderWriter()
	fpb := core.NewLocalFilePBReader(filepath.Dir(opts.bundle))
	if err := run(ctx, opts, brw, fpb); err != nil {
		log.Exit(err)
	}
}

// createInlinerFn creates an Inliner that works with the given current working
// directory for the purposes of dependency injection.
var createInlinerFn = func(pbr core.FilePBReader) *transformer.Inliner {
	return &transformer.Inliner{pbr}
}

func run(ctx context.Context, o *options, brw *converter.BundleReaderWriter, pbr core.FilePBReader) error {
	b, err := brw.ReadBundleFile(ctx, o.bundle)
	if err != nil {
		return err
	}

	inlined, err := createInlinerFn(pbr).Inline(ctx, b)
	if err != nil {
		return err
	}
	return brw.WriteBundleFile(ctx, o.output, inlined, cmdlib.DefaultFilePermissions)
}
