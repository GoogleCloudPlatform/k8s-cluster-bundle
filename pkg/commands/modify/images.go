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

package modify

import (
	"context"
	"fmt"
	"strings"

	log "github.com/golang/glog"
	"github.com/spf13/cobra"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/cmdlib"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/files"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/transformer"
)

// options represents options flags for the images command.
type options struct {
	// Pairs of find-replace pairs. Should have the form foo,bar;biff,bam;
	findReplacePairs string
}

// opts is a global options flags instance for reference via the cobra command
// installation.
var opts = &options{}

func modifyImagesAction(ctx context.Context, cmd *cobra.Command, _ []string) {
	rw := &files.LocalFileSystemReaderWriter{}
	gopts := cmdlib.GlobalOptionsValues.Copy()
	if err := runModifyImages(ctx, opts, rw, gopts); err != nil {
		log.Exitf("error in runModifyImages: %v", err)
	}
}

func runModifyImages(ctx context.Context, opts *options, rw files.FileReaderWriter, gopt *cmdlib.GlobalOptions) error {
	brw := cmdlib.NewBundleReaderWriter(rw)

	bw, err := brw.ReadBundleData(ctx, gopt)
	if err != nil {
		return fmt.Errorf("error reading contents: %v", err)
	}

	var rules []*transformer.ImageSubRule
	splat := strings.Split(opts.findReplacePairs, ";")
	for _, s := range splat {
		sp := strings.Split(s, ",")
		if len(sp) != 2 {
			return fmt.Errorf("find-replace-pairs must have the form <find,replace> -- found %q", s)
		}
		rules = append(rules, &transformer.ImageSubRule{
			Find:    sp[0],
			Replace: sp[1],
		})
	}

	repl := transformer.NewImageTransformer(bw.AllComponents()).TransformImagesStringSub(rules)
	if bw.Bundle != nil {
		bw.Bundle.Components = repl
	} else if bw.Component != nil {
		if len(repl) != 1 {
			return fmt.Errorf("got %d components, but expected exactly one after component-image transform.", len(repl))
		}
		bw.Component = repl[0]
	}

	return brw.WriteBundleData(ctx, bw, gopt)
}
