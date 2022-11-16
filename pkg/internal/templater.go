package internal

import (
	"io"
	"text/template"

	"github.com/google/safetext/yamltemplate"
)

// A Templater provides templating functionality
type Templater struct {
	useSafeYAMLTemplater bool
	yamlTemplater        *yamltemplate.Template
	standardTemplater    *template.Template
}

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
