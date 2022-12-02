package internal

import (
	"io"
	"text/template"

	"github.com/google/safetext/yamltemplate"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SafeYAMLAnnotation is an annotation that can be put on a component that
// indicates that the component should be processed with the SafeYAML library
// rather than the default library.
const SafeYAMLAnnotation = "bundle.gke.io/safe-yaml"

// HasSafeYAMLAnnotation returns whether ObjectMeta contains the
// SafeYAMLAnnotation.
func HasSafeYAMLAnnotation(objmeta metav1.ObjectMeta) bool {
	annot := objmeta.GetAnnotations()
	if annot == nil {
		return false
	}
	_, ok := annot[SafeYAMLAnnotation]
	return ok
}

// Templater provides templating functionality for YAML templating.
type Templater struct {
	useSafeYAMLTemplater bool
	yamlTemplater        *yamltemplate.Template
	standardTemplater    *template.Template
}

// NewTemplater creates a new Templater.
func NewTemplater(tmplName, templateDoc string, funcs map[string]interface{}, useSafeYAMLTemplater bool) (*Templater, error) {
	if useSafeYAMLTemplater {
		t := yamltemplate.New(tmplName + "-safetmpl")
		if funcs != nil {
			t.Funcs(funcs)
		}
		t, err := t.Parse(templateDoc)
		if err != nil {
			return nil, err
		}
		return &Templater{useSafeYAMLTemplater: true, yamlTemplater: t}, nil
	}

	t := template.New(tmplName + "-tmpl")
	if funcs != nil {
		t.Funcs(funcs)
	}
	t, err := t.Parse(templateDoc)
	if err != nil {
		return nil, err
	}
	return &Templater{standardTemplater: t}, nil
}

// Option sets an option on the underlying Templater.
func (t *Templater) Option(opt ...string) *Templater {
	if t.useSafeYAMLTemplater {
		t.yamlTemplater = t.yamlTemplater.Option(opt...)
	} else {
		t.standardTemplater.Option(opt...)
	}
	return t
}

// Execute executes the template rendering.
func (t *Templater) Execute(wr io.Writer, data any) error {
	if t.useSafeYAMLTemplater {
		return t.yamlTemplater.Execute(wr, data)
	}
	return t.standardTemplater.Execute(wr, data)
}
