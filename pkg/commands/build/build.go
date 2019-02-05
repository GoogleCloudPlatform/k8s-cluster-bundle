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
	"fmt"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/build"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/cmdlib"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/files"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/filter"
	log "github.com/golang/glog"
	"github.com/spf13/cobra"
)

// options represents options flags for the build command.
type options struct {
	// annotations selects a subset of patch templates on which to apply options
	// Has the form "foo,bar;biff,baz".
	annotations string

	// optionsFile contains yaml or json structured data containing options to
	// apply to PatchTemplates
	// TODO(jbelamaric): Make this a list of files
	optionsFile string
}

// opts is a global options instance for reference via the add commands.
var opts = &options{}

func action(ctx context.Context, cmd *cobra.Command, _ []string) {
	gopt := cmdlib.GlobalOptionsValues.Copy()
	gopt.Inline = true
	rw := &files.LocalFileSystemReaderWriter{}
	brw := cmdlib.NewBundleReaderWriter(
		rw,
		&cmdlib.RealStdioReaderWriter{})
	if err := run(ctx, opts, brw, rw, gopt); err != nil {
		log.Exit(err)
	}
}

// createInlinerFn creates an Inliner that works with the given current working
// directory for the purposes of dependency injection.
var createInlinerFn = func(pbr files.FileObjReader) *build.Inliner {
	return build.NewInlinerWithScheme(files.FileScheme, pbr)
}

func run(ctx context.Context, o *options, brw cmdlib.BundleReaderWriter, rw files.FileReaderWriter, gopt *cmdlib.GlobalOptions) error {
	bw, err := brw.ReadBundleData(ctx, gopt)
	if err != nil {
		return fmt.Errorf("error reading bundle contents: %v", err)
	}

	// the bundle now contains components which may include PatchTemplateBuilder objects
	// that we need to build into PatchTemplates
        optFiles := []string{}
        if o.optionsFile != "" {
                optFiles = []string{o.optionsFile}
        }

        buildOpts, err := cmdlib.MergeOptions(ctx, rw, optFiles)
        if err != nil {
                return err
        }

	bw, err = build.BuildAllPatchTemplates(bw, &filter.Options{}, buildOpts)
	if err != nil {
		return err
	}

	return brw.WriteBundleData(ctx, bw, gopt)
}
