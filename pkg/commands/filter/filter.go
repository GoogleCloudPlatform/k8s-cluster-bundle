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
	"fmt"
	"strings"

	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/cmdlib"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/filter"
	log "github.com/golang/glog"
	"github.com/spf13/cobra"
)

// options represents options flags for the filter command.
type options struct {
	// Either 'components' or 'objects'. Defaults to components.
	filterType string

	// Comma-separated kinds to filter
	kinds string

	// Comma-separated names to filter
	names string

	// Comma-separated namespaces to filter
	namespaces string

	// Comma + semicolon separated annotations to filter
	// Example: foo,bar;biff,bam
	annotations string

	// Comma + semicolon separated annotations to filter
	// Example: foo,bar;biff,bam
	labels string

	// Whether to keep matches rather then remove them.
	keepOnly bool
}

// opts is a global options instance for reference via the add commands.
var opts = &options{}

func action(ctx context.Context, cmd *cobra.Command, _ []string) {
	gopt := cmdlib.GlobalOptionsValues.Copy()
	gopt.Inline = true
	brw := converter.NewFileSystemBundleReaderWriter()
	if err := run(ctx, opts, brw, gopt); err != nil {
		log.Exit(err)
	}
}

func run(ctx context.Context, o *options, brw *converter.BundleReaderWriter, gopt *cmdlib.GlobalOptions) error {
	b, err := cmdlib.ReadBundleContents(ctx, brw.RW, gopt)
	if err != nil {
		return fmt.Errorf("error reading bundle contents: %v", err)
	}

	fopts := &filter.Options{}
	if o.kinds != "" {
		fopts.Kinds = strings.Split(o.kinds, ",")
	}
	if o.names != "" {
		fopts.Names = strings.Split(o.names, ",")
	}
	if o.namespaces != "" {
		fopts.Namespaces = strings.Split(o.namespaces, ",")
	}
	if o.annotations != "" {
		m := make(map[string]string)
		splat := strings.Split(o.annotations, ";")
		for _, v := range splat {
			kv := strings.Split(v, ",")
			if len(kv) == 2 {
				m[kv[0]] = kv[1]
			}
		}
		fopts.Annotations = m
	}
	if o.labels != "" {
		m := make(map[string]string)
		splat := strings.Split(o.labels, ";")
		for _, v := range splat {
			kv := strings.Split(v, ",")
			if len(kv) == 2 {
				m[kv[0]] = kv[1]
			}
		}
		fopts.Labels = m
	}
	fopts.KeepOnly = o.keepOnly

	if o.filterType == "components" {
		b = (&filter.Filterer{b}).FilterComponents(fopts)
	} else {
		b = (&filter.Filterer{b}).FilterObjects(fopts)
	}

	return cmdlib.WriteStructuredContents(ctx, b, brw.RW, gopt)
}
