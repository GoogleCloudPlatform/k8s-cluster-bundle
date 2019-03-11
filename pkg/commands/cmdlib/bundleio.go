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
	"strings"

	log "k8s.io/klog"

	bundle "github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/apis/bundle/v1alpha1"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/build"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/converter"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/files"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/options"
	"github.com/GoogleCloudPlatform/k8s-cluster-bundle/pkg/wrapper"
)

type makeInliner func(rw files.FileReaderWriter, inputFile string) fileInliner

func realInlinerMaker(rw files.FileReaderWriter, inputFile string) fileInliner {
	return build.NewInlinerWithScheme(
		files.FileScheme,
		&files.LocalFileObjReader{
			WorkingDir: filepath.Dir(inputFile),
			Rdr:        rw,
		})
}

// StdioReaderWriter can read from STDIN and write to STDOUT.
type StdioReaderWriter interface {
	ReadAll() ([]byte, error)
	io.Writer
}

// RealStdioReaderWriter provides a real STDIN / STDOUT implementation.
type RealStdioReaderWriter struct{}

// ReadAll reads all the input from STDIN.
func (r *RealStdioReaderWriter) ReadAll() ([]byte, error) {
	return ioutil.ReadAll(os.Stdin)
}

// Write writes content to STDOUT.
func (r *RealStdioReaderWriter) Write(b []byte) (int, error) {
	return os.Stdout.Write(b)
}

type fileInliner interface {
	BundleFiles(context.Context, *bundle.BundleBuilder, string) (*bundle.Bundle, error)
	ComponentFiles(context.Context, *bundle.ComponentBuilder, string) (*bundle.Component, error)
}

// BundleReaderWriter provides a mockable reading / writing interface for
// reading / writing bundle data.
type BundleReaderWriter interface {
	ReadBundleData(context.Context, *GlobalOptions) (*wrapper.BundleWrapper, error)
	WriteBundleData(context.Context, *wrapper.BundleWrapper, *GlobalOptions) error
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
func (brw *realBundleReaderWriter) ReadBundleData(ctx context.Context, g *GlobalOptions) (*wrapper.BundleWrapper, error) {
	var bytes []byte
	var err error
	inFmt := g.InputFormat

	if g.InputFile != "" {
		log.V(4).Infof("Reading input file %v", g.InputFile)
		bytes, err = brw.rw.ReadFile(ctx, g.InputFile)
		if err != nil {
			return nil, err
		}
	} else {
		log.V(4).Info("No component data file, reading from stdin")
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

	bw, err := wrapper.FromRaw(inFmt, bytes)
	if err != nil {
		return nil, err
	}

	// For now, we can only inline component data files because we need the path
	// context.
	if g.InputFile != "" && (bw.BundleBuilder() != nil || bw.ComponentBuilder() != nil) {
		return brw.inlineData(ctx, bw, g)
	}
	return bw, nil
}

// inlineData inlines a cluster bundle before processing
func (brw *realBundleReaderWriter) inlineData(ctx context.Context, bw *wrapper.BundleWrapper, g *GlobalOptions) (*wrapper.BundleWrapper, error) {
	inliner := brw.makeInlinerFn(brw.rw, g.InputFile)
	switch bw.Kind() {
	case "BundleBuilder":
		newBun, err := inliner.BundleFiles(ctx, bw.BundleBuilder(), g.InputFile)
		if err != nil {
			return nil, fmt.Errorf("error inlining component data files: %v", err)
		}
		return wrapper.FromBundle(newBun), nil
	case "ComponentBuilder":
		newComp, err := inliner.ComponentFiles(ctx, bw.ComponentBuilder(), g.InputFile)
		if err != nil {
			return nil, fmt.Errorf("error inlining objects: %v", err)
		}
		return wrapper.FromComponent(newComp), nil
	default:
		// Bundle and Component types can't be inlined.
		return bw, nil
	}
}

// WriteBundleData writes either the component or bundle object from the BundleWrapper.
func (brw *realBundleReaderWriter) WriteBundleData(ctx context.Context, bw *wrapper.BundleWrapper, g *GlobalOptions) error {
	if bw == nil {
		return fmt.Errorf("bundle wrapper was nil")
	}

	obj := bw.Object()
	if obj == nil {
		return fmt.Errorf("wrapped bundle object was nil")
	}

	return brw.WriteStructuredContents(ctx, obj, g)
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

// ParseStringMap parses a CLI flag value into a map of string to string. It
// expects the raw flag value to have the form "key1=value1,key2=value2,etc".
func ParseStringMap(p string) map[string]string {
	if p == "" {
		return nil
	}

	m := make(map[string]string)
	splat := strings.Split(p, ",")
	for _, v := range splat {
		kv := strings.Split(v, "=")
		if len(kv) == 2 {
			m[kv[0]] = kv[1]
		}
	}
	return m
}

// MergeOptions reads multiple options files and combines them into a single map.
// Option values will be overwritten if a later file has the same key.
func MergeOptions(ctx context.Context, rw files.FileReaderWriter, files []string) (options.JSONOptions, error) {
	optData := make(map[string]interface{})
	for _, f := range files {
		bytes, err := rw.ReadFile(ctx, f)
		if err != nil {
			return nil, err
		}
		data, err := converter.FromFileName(f, bytes).ToJSONMap()
		if err != nil {
			return nil, err
		}
		for k, v := range data {
			optData[k] = v
		}
	}
	return optData, nil
}
