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

package export

import (
	"context"
	"fmt"
	"path/filepath"

	bpb "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/cmdlib"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/files"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/find"
	log "github.com/golang/glog"
	"github.com/spf13/cobra"
)

// compfinder provides an interface for exporting ComponentPackages from a
// ClusterBundle.
type compfinder interface {
	ComponentPackage(compName string) *bpb.ComponentPackage
}

type options struct {
	components []string
	outputDir  string
}

var opts = &options{}

func action(ctx context.Context, cmd *cobra.Command, _ []string) {
	if opts.components == nil {
		cmdlib.ExitWithHelp(cmd, "Please provide at least one component to extract.")
	}
	gopt := cmdlib.GlobalOptionsValues.Copy()
	rw := &files.LocalFileSystemReaderWriter{}
	if err := run(ctx, opts, rw, gopt); err != nil {
		log.Exit(err)
	}
}

// createFinderFn creates an exporter that operates on the given ClusterBundle.
var createFinderFn = func(b *bpb.ClusterBundle) (compfinder, error) {
	return find.NewBundleFinder(b)
}

func run(ctx context.Context, o *options, rw files.FileReaderWriter, gopt *cmdlib.GlobalOptions) error {
	b, err := cmdlib.ReadBundleContents(ctx, rw, gopt)
	if err != nil {
		return fmt.Errorf("error reading bundle contents: %v", err)
	}

	f, err := createFinderFn(b)
	if err != nil {
		return err
	}

	for _, comp := range o.components {
		ea := f.ComponentPackage(comp)
		if ea == nil {
			return fmt.Errorf("could not find cluster component named %q", comp)
		}

		// If a write fails, just return the error and the user can rerun the command and rewrite
		// any files that may have been written or partially written.
		path := fmt.Sprintf("%s/%s.yaml", filepath.Clean(o.outputDir), ea.GetMetadata().GetName())
		bytes, err := converter.ComponentPackage.ProtoToYAML(ea)
		if err != nil {
			return err
		}
		err = rw.WriteFile(ctx, path, bytes, cmdlib.DefaultFilePermissions)
		if err != nil {
			return err
		}
		log.Infof("Wrote file %q", path)
	}
	return nil
}
