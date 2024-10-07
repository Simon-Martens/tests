package templating

import (
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"path/filepath"
	"strings"
)

var NoTemplateError = errors.New("no template found for this path")
var InvalidPathError = errors.New("invalid path. Could not create a sub.FS")
var TEMPLATE_FORMATS = [...]string{"html", "tmpl", "gotmpl", "gotemplate"}

const TEMPLATE_GLOBAL_CONTEXT_NAME = "globals"
const TEMPLATE_LOCAL_CONTEXT_NAME = "locals"
const TEMPLATE_GLOBAL_PREFIX = "_"
const TEMPLATE_COMPONENT_DIRECTORY = "components"
const TEMPLATE_HEAD = "head"
const TEMPLATE_BODY = "body"
const TEMPLATE_HEADERS = "headers"

type TemplateContext struct {
	Path    string
	locals  map[string]*template.Template
	globals map[string]*template.Template
	body    *template.Template
	head    *template.Template
	Headers *template.Template
}

func NewTemplateContext(path string) *TemplateContext {
	return &TemplateContext{
		Path:    path,
		locals:  make(map[string]*template.Template),
		globals: make(map[string]*template.Template),
	}
}

func (c *TemplateContext) Parse(fsys fs.FS) error {
	globs := Globs()
	fsys, err := fs.Sub(fsys, c.Path)
	if err != nil {
		return InvalidPathError
	}

	for _, g := range globs {
		matches, err := fs.Glob(fsys, g)
		if err != nil {
			fmt.Println("Invalid glob " + g)
			return err
		}

		// We do not use the template.ParseFS method here because we use
		// template names without file extension.
		for _, m := range matches {
			f, err := fs.Stat(fsys, m)
			if err != nil {
				fmt.Println("Could not stat " + m)
				continue
			}

			if f.IsDir() {
				continue
			}

			text, err := fs.ReadFile(fsys, m)
			if err != nil {
				fmt.Println("Could not read " + m)
				continue
			}

			name := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))

			fmt.Println("Parsing local template with name: " + name)
			t, err := template.New(name).Parse(string(text))
			if err != nil {
				fmt.Println("Could not parse " + m)
				continue
			}

			if strings.HasPrefix(g, TEMPLATE_GLOBAL_PREFIX) {
				c.globals[name] = t
			} else {
				c.locals[name] = t
			}
		}
	}

	return nil
}

func (c *TemplateContext) SetGlobals(globals map[string]*template.Template) error {
	for k, v := range globals {
		c.globals[k] = v
	}
	return nil
}

func (c *TemplateContext) GetGlobals() map[string]*template.Template {
	return c.globals
}

func (c *TemplateContext) GetRoot() (*template.Template, error) {
	if c.body != nil {
		return c.body, nil
	}

	t, ok := c.locals["root"]
	if !ok {
		return nil, NoTemplateError
	}

	for _, g := range c.globals {
		_, err := t.AddParseTree(g.Name(), g.Tree)
		if err != nil {
			return nil, err
		}
	}

	for _, g := range c.locals {
		_, err := t.AddParseTree(g.Name(), g.Tree)
		if err != nil {
			return nil, err
		}
	}

	c.body = t

	return t, nil
}

func (c *TemplateContext) GetBody() (*template.Template, error) {
	if c.body != nil {
		return c.body, nil
	}

	t, ok := c.locals["body"]
	if !ok {
		return nil, NoTemplateError
	}

	for _, g := range c.globals {
		_, err := t.AddParseTree(g.Name(), g.Tree)
		if err != nil {
			return nil, err
		}
	}

	for _, g := range c.locals {
		_, err := t.AddParseTree(g.Name(), g.Tree)
		if err != nil {
			return nil, err
		}
	}

	c.body = t

	return t, nil
}

func (c *TemplateContext) GetHead() (*template.Template, error) {
	if c.head != nil {
		return c.head, nil
	}

	t, ok := c.locals["head"]
	if !ok {
		return nil, NoTemplateError
	}

	for _, g := range c.globals {
		_, err := t.AddParseTree(g.Name(), g.Tree)
		if err != nil {
			return nil, err
		}
	}

	for _, g := range c.locals {
		_, err := t.AddParseTree(g.Name(), g.Tree)
		if err != nil {
			return nil, err
		}
	}

	c.head = t

	return t, nil
}

func Globs() []string {
	g := []string{}
	for _, f := range TEMPLATE_FORMATS {
		g = append(g, TEMPLATE_GLOBAL_PREFIX+TEMPLATE_COMPONENT_DIRECTORY+"/*."+f)
		g = append(g, TEMPLATE_COMPONENT_DIRECTORY+"/*."+f)
		g = append(g, "*."+f)
	}

	return g
}
