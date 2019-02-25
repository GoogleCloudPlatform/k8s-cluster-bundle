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
	"os"
	"path/filepath"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/commands/cmdlib"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/files"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/wrapper"
	log "k8s.io/klog"
)

// pathTemplates are go-template paths to indicate how and where to generate
// component files.  Possible values are {ComponentName, Version, BuildTag,
// SetName}.  Only applies when the write-to flag is also set.
var pathTemplates = []string{
	"{{.ComponentName}}/{{.ComponentName}}-{{.Version}}.yaml",
	"{{.ComponentName}}/{{.BuildTag}}/{{.ComponentName}}.yaml",
	"sets/{{.SetName}}-{{.Version}}.yaml",
}

const defaultPerms = 0664

// compOptions represents options flags for the export command.
type compOptions struct {
	// writeTo indicates the directory to write the components to. This is joined with the pa
	writeTo string

	// overwrite indicates whether to overwrite existing files.
	overwrite bool

	// exportSet determines whether to export a set for all the known components.
	exportSet bool

	// setName is an optional set name for the set of components.
	setName string

	// version is an optional version for the components
	setVersion string
}

func componentAction(ctx context.Context, fio files.FileReaderWriter, sio cmdlib.StdioReaderWriter, o *compOptions, gopts *cmdlib.GlobalOptions) {
	brw := cmdlib.NewBundleReaderWriter(fio, sio)
	if err := runComponent(ctx, brw, sio, fio, o, gopts); err != nil {
		log.Exit(err)
	}
}

func runComponent(ctx context.Context, brw cmdlib.BundleReaderWriter, stdio cmdlib.StdioReaderWriter, fio files.FileReaderWriter, o *compOptions, gopt *cmdlib.GlobalOptions) error {
	bw, err := brw.ReadBundleData(ctx, gopt)
	if err != nil {
		return fmt.Errorf("error reading bundle contents: %v", err)
	}

	comps := bw.AllComponents()

	var set *bundle.ComponentSet
	if o.exportSet {
		setName, setVersion := getSetNameVersion(bw, o)
		set, err = converter.NewExporter(comps...).ComponentSet(setName, setVersion)
		if err != nil {
			return err
		}
	}

	if o.writeTo != "" {
		return writeToFiles(ctx, comps, o, fio, set)
	} else {
		return writeToStdout(comps, stdio, set)
	}
}

func getSetNameVersion(bw *wrapper.BundleWrapper, o *compOptions) (string, string) {
	setName := o.setName
	setVersion := o.setVersion
	if bw.Kind() == "bundle" {
		bun := bw.Bundle()
		setName = bun.SetName
		setVersion = bun.Version
	}
	return setName, setVersion
}

func writeToFiles(ctx context.Context, comps []*bundle.Component, o *compOptions, fio files.FileReaderWriter, set *bundle.ComponentSet) error {
	pathMap, err := converter.NewExporter(comps...).ComponentsWithPathTemplates(pathTemplates, set)
	if err != nil {
		return err
	}
	for cpath, yaml := range pathMap {
		fpath := filepath.Join(o.writeTo, cpath)
		dir := filepath.Dir(fpath)
		// TODO(kashomon): This is going to be a problem for mocking.
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			os.MkdirAll(dir, os.ModePerm)
		}
		if !o.overwrite {
			if _, err := os.Stat(fpath); err == nil {
				return fmt.Errorf("file %q already exists", fpath)
			}
		}
		if err := fio.WriteFile(ctx, fpath, []byte(yaml), defaultPerms); err != nil {
			return err
		}
	}
	return nil
}

func writeToStdout(comps []*bundle.Component, stdio cmdlib.StdioReaderWriter, set *bundle.ComponentSet) error {
	data, err := converter.NewExporter(comps...).ComponentsAsSingleYAML(set)
	if err != nil {
		return err
	}
	_, err = stdio.Write([]byte(data))
	return err
}
