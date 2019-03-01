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
	"github.com/spf13/cobra"
	log "k8s.io/klog"
)

// options represents options flags for the build command.
type options struct {
	// optionsFile contains yaml or json structured data containing options to
	// apply to PatchTemplates
	// TODO(jbelamaric): Make this a list of files
	optionsFile string
}


func action(ctx context.Context, fio files.FileReaderWriter, sio cmdlib.StdioReaderWriter, cmd *cobra.Command, _ []string, opts *options) {
	gopt := cmdlib.GlobalOptionsValues.Copy()
	brw := cmdlib.NewBundleReaderWriter(fio, sio)
	if err := run(ctx, opts, brw, fio, gopt); err != nil {
		log.Exit(err)
	}
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

	bw, err = build.AllPatchTemplates(bw, &filter.Options{}, buildOpts)
	if err != nil {
		return err
	}

	return brw.WriteBundleData(ctx, bw, gopt)
}
