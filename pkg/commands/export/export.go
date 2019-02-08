// Copyright 2019 Google LLC
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

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/cmdlib"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/files"
	log "github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// options represents options flags for the export command.
type options struct{}

// opts is a global options instance for reference via the add commands.
var opts = &options{}

func action(ctx context.Context, cmd *cobra.Command, _ []string) {
	gopt := cmdlib.GlobalOptionsValues.Copy()
	stdio := cmdlib.RealStdioReaderWriter{}
	brw := cmdlib.NewBundleReaderWriter(
		&files.LocalFileSystemReaderWriter{},
		&stdio,
	)
	if err := run(ctx, opts, brw, stdio, gopt); err != nil {
		log.Exit(err)
	}
}

func run(ctx context.Context, o *options, brw cmdlib.BundleReaderWriter, stdio cmdlib.RealStdioReaderWriter, gopt *cmdlib.GlobalOptions) error {
	bw, err := brw.ReadBundleData(ctx, gopt)
	if err != nil {
		return fmt.Errorf("error reading bundle contents: %v", err)
	}

	comps := bw.AllComponents()
	if len(comps) == 0 {
		return nil
	}

	var objs []*unstructured.Unstructured
	for _, c := range comps {
		objs = append(objs, c.Spec.Objects...)
	}
	exporter := converter.ObjectExporter{Objects: objs}
	s, err := exporter.ExportAsYAML()
	if err != nil {
		return err
	}
	_, err = stdio.Write([]byte(s))
	return err
}
