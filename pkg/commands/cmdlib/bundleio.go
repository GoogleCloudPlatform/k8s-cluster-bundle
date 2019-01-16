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
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/golang/glog"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/files"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/inline"
)

// BundleWrapper represents a collection of possibly specified bundle types. It
// is meant to be internal -- only used / usable by the commands.
type BundleWrapper struct {
	Bundle    *bundle.Bundle
	Component *bundle.Component
}

// AllComponents returns all the components in the BundleWrapper, with the
// assumption that only one of Bundle or Component is filled out.
func (bw *BundleWrapper) AllComponents() []*bundle.Component {
	if bw.Bundle != nil {
		return bw.Bundle.Components
	} else if bw.Component != nil {
		return []*bundle.Component{bw.Component}
	}
	return []*bundle.Component{}
}

type makeInliner func(rw files.FileReaderWriter, inputFile string) fileInliner

func realInlinerMaker(rw files.FileReaderWriter, inputFile string) fileInliner {
	return inline.NewInlinerWithScheme(
		files.FileScheme,
		&files.LocalFileObjReader{filepath.Dir(inputFile), rw},
		inline.DefaultPathRewriter,
	)
}

// StdioReaderWriter can read from STDIN and write to STDOUT.
type StdioReaderWriter interface {
	ReadAll() ([]byte, error)
	io.Writer
}

type RealStdioReaderWriter struct{}

func (r *RealStdioReaderWriter) ReadAll() ([]byte, error) {
	return ioutil.ReadAll(os.Stdin)
}

func (r *RealStdioReaderWriter) Write(b []byte) (int, error) {
	return os.Stdout.Write(b)
}

type fileInliner interface {
	InlineBundleFiles(context.Context, *bundle.Bundle) (*bundle.Bundle, error)
	InlineComponentsInBundle(context.Context, *bundle.Bundle) (*bundle.Bundle, error)
	InlineComponent(context.Context, *bundle.Component) (*bundle.Component, error)
}

type BundleReaderWriter interface {
	ReadBundleData(context.Context, *GlobalOptions) (*BundleWrapper, error)
	WriteBundleData(context.Context, *BundleWrapper, *GlobalOptions) error
	WriteStructuredContents(context.Context, interface{}, *GlobalOptions) error
}

// BundleReaderWriter is an object that can read and write bundle information,
// in the context of CLI flags.
type realBundleReaderWriter struct {
	rw            files.FileReaderWriter
	stdio         StdioReaderWriter
	makeInlinerFn makeInliner
}

// NewBundleReaderWriter creates a new BundleReaderWriter.
func NewBundleReaderWriter(rw files.FileReaderWriter, stdio StdioReaderWriter) BundleReaderWriter {
	return &realBundleReaderWriter{
		rw:            rw,
		stdio:         stdio,
		makeInlinerFn: realInlinerMaker,
	}
}

// ReadBundleData reads either data file contents from a file or stdin.
func (brw *realBundleReaderWriter) ReadBundleData(ctx context.Context, g *GlobalOptions) (*BundleWrapper, error) {
	var bytes []byte
	var err error
	inFmt := g.InputFormat

	if g.InputFile != "" {
		log.Infof("Reading input file %v", g.InputFile)
		bytes, err = brw.rw.ReadFile(ctx, g.InputFile)
		if err != nil {
			return nil, err
		}
	} else {
		log.Info("No component data file, reading from stdin")
		if bytes, err = brw.stdio.ReadAll(); err != nil {
			return nil, err
		}
	}

	fileFmt := formatFromFile(g.InputFile)
	if fileFmt != "" && inFmt != "" {
		inFmt = fileFmt
	}
	if inFmt == "" {
		inFmt = "yaml"
	}

	bw := &BundleWrapper{}
	uns, err := converter.FromContentType(inFmt, bytes).ToUnstructured()
	if err != nil {
		return nil, fmt.Errorf("error converting bundle content to unstructured: %v", err)
	}

	kind := uns.GetKind()
	if kind == "Bundle" {
		b, err := converter.FromContentType(inFmt, bytes).ToBundle()
		if err != nil {
			return nil, err
		}
		bw.Bundle = b
	} else if kind == "Component" {
		c, err := converter.FromContentType(inFmt, bytes).ToComponent()
		if err != nil {
			return nil, err
		}
		bw.Component = c
	} else {
		return nil, fmt.Errorf("unrecognized bundle-kind %s for content %s", kind, string(bytes))
	}

	// For now, we can only inline component data files because we need the path
	// context.
	if (g.InlineComponents || g.InlineObjects) && g.InputFile != "" {
		return brw.inlineData(ctx, bw, g)
	}
	return bw, nil
}

// inlineData inlines a cluster bundle before processing
func (brw *realBundleReaderWriter) inlineData(ctx context.Context, bw *BundleWrapper, g *GlobalOptions) (*BundleWrapper, error) {
	var err error
	inliner := brw.makeInlinerFn(brw.rw, g.InputFile)
	if g.InlineComponents && bw.Bundle != nil {
		bw.Bundle, err = inliner.InlineBundleFiles(ctx, bw.Bundle)
		if err != nil {
			return nil, fmt.Errorf("error inlining component data files: %v", err)
		}
	}
	if g.InlineObjects && bw.Bundle != nil {
		bw.Bundle, err = inliner.InlineComponentsInBundle(ctx, bw.Bundle)
		if err != nil {
			return nil, fmt.Errorf("error inlining objects: %v", err)
		}
	}
	if g.InlineObjects && bw.Component != nil {
		bw.Component, err = inliner.InlineComponent(ctx, bw.Component)
		if err != nil {
			return nil, fmt.Errorf("error inlining objects: %v", err)
		}
	}
	return bw, nil
}

// WriteBundleData writes either the component or bundle object from the BundleWrapper.
func (brw *realBundleReaderWriter) WriteBundleData(ctx context.Context, bw *BundleWrapper, g *GlobalOptions) error {
	if bw.Bundle != nil && bw.Component != nil {
		return fmt.Errorf("both the bundle and the component fields were non-nil in the BundleWrapper")
	}

	if bw.Bundle != nil {
		return brw.WriteStructuredContents(ctx, bw.Bundle, g)
	} else if bw.Component != nil {
		return brw.WriteStructuredContents(ctx, bw.Component, g)
	} else {
		return fmt.Errorf("both the bundle and the component fields were nil")
	}
}

// WriteStructuredContents writes some structured contents from some object
// `obj`. The contents must be serializable to both JSON and YAML.
func (brw *realBundleReaderWriter) WriteStructuredContents(ctx context.Context, obj interface{}, g *GlobalOptions) error {
	outFmt := g.OutputFormat
	if outFmt == "" {
		outFmt = "yaml"
	}

	bytes, err := converter.FromObject(obj).ToContentType(outFmt)
	if err != nil {
		return fmt.Errorf("error writing contents: %v", err)
	}
	return brw.writeContents(ctx, bytes, brw.rw)
}

// writeContents writes some bytes to stdout. if outPath is empty, write to
func (brw *realBundleReaderWriter) writeContents(ctx context.Context, bytes []byte, rw files.FileReaderWriter) error {
	_, err := brw.stdio.Write(bytes)
	return err
}

// formatFromFile gets the content format from a file-extension and returns
// empty string if the extension couldn't be mapped
func formatFromFile(path string) string {
	ext := filepath.Ext(path)
	switch ext {
	case ".yaml", ".yml":
		return "yaml"
	case ".json":
		return "json"
	}
	return ""
}
