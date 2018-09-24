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
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/core"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/transformer"
	log "github.com/golang/glog"
	"github.com/spf13/cobra"
)

// exporter provides an interface for exporting ClusterComponents from a
// ClusterBundle.
type exporter interface {
	Export(b *bpb.ClusterBundle, compName string) (*transformer.ExportedComponent, error)
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
	rw := &core.LocalFileSystemReaderWriter{}
	if err := run(ctx, opts, rw, gopt); err != nil {
		log.Exit(err)
	}
}

// createExporterFn creates an exporter that operates on the given ClusterBundle.
var createExporterFn = func(b *bpb.ClusterBundle) (exporter, error) {
	return transformer.NewComponentExporter(b)
}

func run(ctx context.Context, o *options, rw core.FileReaderWriter, gopt *cmdlib.GlobalOptions) error {
	b, err := cmdlib.ReadBundleContents(ctx, rw, gopt)
	if err != nil {
		return fmt.Errorf("error reading bundle contents: %v", err)
	}

	exporter, err := createExporterFn(b)
	if err != nil {
		return err
	}

	for _, comp := range o.components {
		ea, err := exporter.Export(b, comp)
		if err != nil {
			return err
		}
		// If a write fails, just return the error and the user can rerun the command and rewrite
		// any files that may have been written or partially written.
		path := fmt.Sprintf("%s/%s.yaml", filepath.Clean(o.outputDir), ea.Name)
		outComp := &bpb.ClusterComponent{
			Name:           ea.Name,
			ClusterObjects: ea.Objects,
		}
		bytes, err := converter.ClusterComponent.ProtoToYAML(outComp)
		if err != nil {
			return err
		}
		err = rw.WriteFile(ctx, path, bytes, cmdlib.DefaultFilePermissions)
		if err != nil {
			return err
		}
		log.Infof("Wrote file %q", ea.Name)
	}
	return nil
}
