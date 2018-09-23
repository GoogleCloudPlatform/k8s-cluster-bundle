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
	"path/filepath"

	"context"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/cmdlib"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/transformer"
	log "github.com/golang/glog"
	"github.com/spf13/cobra"
)

// options represents options flags for the inline command.
type options struct {
	bundle string
	output string
}

// opts is a global Options instance for referenc via the
var opts = &options{}

func action(ctx context.Context, cmd *cobra.Command, _ []string) {
	if opts.bundle == "" {
		cmdlib.ExitWithHelp(cmd, "Please provide yaml file for bundle.")
	}
	if err := run(ctx, opts, &cmdlib.RealReaderWriter{}); err != nil {
		log.Exit(err)
	}
}

// createInlinerFn creates an Inliner that works with the given current working
// directory for the purposes of dependency injection.
var createInlinerFn = func(cwd string) *transformer.Inliner {
	return transformer.NewInliner(cwd)
}

func run(ctx context.Context, opts *options, brw cmdlib.BundleReaderWriter) error {
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
	return brw.WriteBundleFile(opts.output, inlined, cmdlib.DefaultFilePermissions)
}
