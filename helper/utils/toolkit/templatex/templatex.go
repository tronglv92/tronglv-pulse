package templatex

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
)

const (
	// DefaultFolderName is the default folder where templates are located.
	DefaultFolderName = "resources/templates"
)

// Template defines an interface to execute a named template file with given data.
type Template interface {
	// Execute parses the template file and executes it with the provided data.
	// Returns the rendered template as a string or an error.
	Execute(fileName string, data any) (string, error)
}

// Options holds configuration for the template service.
type Options struct {
	FuncMap    template.FuncMap // Optional custom functions for templates
	FolderName string           // Template files folder path
}

// service implements the Template interface.
type service struct {
	*Options
}

// Option is a functional option to configure the template service.
type Option func(opts *Options)

// WithFuncMap sets custom template functions.
func WithFuncMap(f template.FuncMap) Option {
	return func(opts *Options) {
		opts.FuncMap = f
	}
}

// WithFolderName sets the template folder path.
func WithFolderName(folder string) Option {
	return func(opts *Options) {
		opts.FolderName = folder
	}
}

// New creates a new Template service with optional configurations.
func New(opts ...Option) Template {
	options := &Options{
		FolderName: DefaultFolderName,
	}
	for _, opt := range opts {
		opt(options)
	}
	return &service{options}
}

// new parses the specified template file applying any custom FuncMap.
func (s *service) new(fileName string) *template.Template {
	tmpl := template.New(fileName)
	if s.FuncMap != nil {
		tmpl = tmpl.Funcs(s.FuncMap)
	}
	tmpl = template.Must(tmpl.ParseFiles(filepath.Join(s.FolderName, fileName)))
	return tmpl
}

// Execute loads, parses, and executes the template file with the provided data,
// returning the rendered output as a string.
func (s *service) Execute(fileName string, data any) (string, error) {
	var buf bytes.Buffer
	tmpl := s.new(fileName)
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template %q: %w", fileName, err)
	}
	return buf.String(), nil
}
