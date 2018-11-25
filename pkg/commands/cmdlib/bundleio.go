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

package cmdlib

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"

	log "github.com/golang/glog"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/files"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/inline"
	"k8s.io/apimachinery/pkg/runtime"
)

// readBytes reads data file contents from a file or stdin.
func readBytes(ctx context.Context, rw files.FileReaderWriter, g *GlobalOptions) ([]byte, error) {
	var bytes []byte
	var err error
	if g.BundleFile != "" {
		log.Infof("Reading bundle file %v", g.BundleFile)
		bytes, err = rw.ReadFile(ctx, g.BundleFile)
		if err != nil {
			return nil, err
		}
	} else {
		log.Info("No component data file, reading from stdin")
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			bytes = append(bytes, scanner.Bytes()...)
			bytes = append(bytes, '\n')
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}
	return bytes, err
}

// ReadBundle reads component data file contents from a file or stdin.
func ReadBundle(ctx context.Context, rw files.FileReaderWriter, g *GlobalOptions) (*bundle.Bundle, error) {
	bytes, err := readBytes(ctx, rw, g)
	if err != nil {
		return nil, err
	}

	b, err := convertBundle(ctx, bytes, g)
	if err != nil {
		return nil, err
	}

	// For now, we can only inline component data files because we need the path
	// context.
	if (g.InlineComponents || g.InlineObjects) && g.BundleFile != "" {
		o, err := inlineData(ctx, b, rw, g)
		if err != nil {
			return nil, err
		}
		b = o.(*bundle.Bundle)
	}
	return b, nil
}

// ReadObject reads a bundle object from a file or stdin.
func ReadObject(ctx context.Context, rw files.FileReaderWriter, g *GlobalOptions) (runtime.Object, error) {
	bytes, err := readBytes(ctx, rw, g)
	if err != nil {
		return nil, err
	}

	o, err := convertToObject(ctx, bytes, g)
	if err != nil {
		return nil, err
	}

	// For now, we can only inline component data files because we need the path
	// context.
	if (g.InlineComponents || g.InlineObjects) && g.BundleFile != "" {
		o, err = inlineData(ctx, o, rw, g)
		if err != nil {
			return nil, err
		}
	}
	return o, nil
}

func convertBundle(ctx context.Context, bt []byte, g *GlobalOptions) (*bundle.Bundle, error) {
	return converter.FromContentType(g.InputFormat, bt).ToBundle()
}

func convertToObject(ctx context.Context, bt []byte, g *GlobalOptions) (runtime.Object, error) {
	return converter.FromContentType(g.InputFormat, bt).ToObject()
}

// inlineData inlines a cluster bundle before processing
func inlineData(ctx context.Context, data runtime.Object, rw files.FileReaderWriter, g *GlobalOptions) (runtime.Object, error) {
	inliner := &inline.Inliner{
		&files.LocalFileObjReader{
			WorkingDir: filepath.Dir(g.BundleFile),
			Rdr:        rw,
		},
	}
	var err error
	if g.InlineComponents {
		switch v := data.(type) {
		case *bundle.Bundle:
			data, err = inliner.InlineBundleFiles(ctx, v)
			if err != nil {
				return nil, fmt.Errorf("error inlining component data files: %v", err)
			}
		case *bundle.ComponentPackage:
			data, err = inliner.InlineComponent(ctx, v)
			if err != nil {
				return nil, fmt.Errorf("error inlining ComponentPackage data files: %v", err)
			}
		default:
			return nil, fmt.Errorf("unhandled type for inlining: %T", data)
		}
	}
	if g.InlineObjects {
		switch v := data.(type) {
		case *bundle.Bundle:
			data, err = inliner.InlineComponentsInBundle(ctx, v)
			if err != nil {
				return nil, fmt.Errorf("error inlining objects: %v", err)
			}
		case *bundle.ComponentPackage:
			// No objects (?)
		default:
			return nil, fmt.Errorf("unhandled type for inlining: %T", data)
		}
	}
	return data, nil
}

// WriteContentsStructured writes some structured contents. The contents must be serializable to both JSON and YAML.
func WriteStructuredContents(ctx context.Context, obj interface{}, rw files.FileReaderWriter, g *GlobalOptions) error {
	bytes, err := converter.FromObject(obj).ToContentType(g.OutputFormat)
	if err != nil {
		return fmt.Errorf("error writing contents: %v", err)
	}
	return WriteContents(ctx, g.OutputFile, bytes, rw)
}

// WriteContents writes some bytes to disk or stdout. if outPath is empty, write to stdout instdea.
func WriteContents(ctx context.Context, outPath string, bytes []byte, rw files.FileReaderWriter) error {
	if outPath == "" {
		_, err := os.Stdout.Write(bytes)
		if err != nil {
			return fmt.Errorf("error writing content %q to stdout", string(bytes))
		}
		return nil
	}
	log.Infof("Writing file to %q", outPath)
	return rw.WriteFile(ctx, outPath, bytes, DefaultFilePermissions)
}
