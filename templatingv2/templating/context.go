package templating

import (
	"errors"
	"html/template"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
)

var InvalidPathError = errors.New("Invalid path. Must be a directory.")
var FileAccessError = errors.New("could not stat file or directory")

var TEMPLATE_FORMATS = []string{".html", ".tmpl", ".gotmpl", ".gotemplate", ".gohtml", ".gohtmltemplate"}

const TEMPLATE_GLOBAL_CONTEXT_NAME = "globals"
const TEMPLATE_LOCAL_CONTEXT_NAME = "locals"
const TEMPLATE_GLOBAL_PREFIX = "_"
const TEMPLATE_COMPONENT_DIRECTORY = "components"
const TEMPLATE_HEAD = "head"
const TEMPLATE_BODY = "body"
const TEMPLATE_HEADERS = "headers"

// We define our own template type, to define some methods on it
type Template template.Template

type TemplateContext struct {
	// WARNING: Path is a URL path, NOT a filesystem path
	Path string
	// WARNING: The keys of these maps are template names, NOT filesystem paths
	// The values are FS paths from the root directory of the templates
	locals  map[string]string
	globals map[string]string
	cache   *template.Template
}

func NewTemplateContext(path string) TemplateContext {
	return TemplateContext{
		Path:    path,
		locals:  make(map[string]string),
		globals: make(map[string]string),
		cache:   nil,
	}
}

func (c *TemplateContext) Parse(fsys fs.FS) error {
	fspath := PathToFSPath(c.Path)
	entries, err := fs.ReadDir(fsys, fspath)

	if err != nil {
		return NewError(InvalidPathError, c.Path)
	}

	for _, e := range entries {
		if e.IsDir() {
			if e.Name() == TEMPLATE_COMPONENT_DIRECTORY {
				entries, err := fs.ReadDir(fsys, filepath.Join(fspath, e.Name()))
				if err != nil {
					return NewError(FileAccessError, filepath.Join(fspath, e.Name()))
				}

				for _, e := range entries {
					ext := filepath.Ext(e.Name())

					if !slices.Contains(TEMPLATE_FORMATS, ext) {
						continue
					}

					name := strings.TrimSuffix(e.Name(), ext)
					if strings.HasPrefix(e.Name(), TEMPLATE_GLOBAL_PREFIX) {
						c.globals[name] = filepath.Join(fspath, TEMPLATE_COMPONENT_DIRECTORY, e.Name())
					} else {
						c.locals[name] = filepath.Join(fspath, TEMPLATE_COMPONENT_DIRECTORY, e.Name())
					}
				}
				continue
			}
		}

		ext := filepath.Ext(e.Name())

		if !slices.Contains(TEMPLATE_FORMATS, ext) {
			continue
		}

		name := strings.TrimSuffix(e.Name(), ext)
		if strings.HasPrefix(e.Name(), TEMPLATE_GLOBAL_PREFIX) {
			c.globals[name] = filepath.Join(fspath, e.Name())
		} else {
			c.locals[name] = filepath.Join(fspath, e.Name())
		}
	}

	return nil
}

func (c *TemplateContext) SetGlobals(globals map[string]string) error {
	// INFO: this allows for overwriting of existing global keys.
	for k, v := range globals {
		c.globals[k] = v
	}
	return nil
}

func (c *TemplateContext) GetGlobals() map[string]string {
	return c.globals
}

func (c *TemplateContext) Add(fsys fs.FS, t *template.Template) (*template.Template, error) {
	if c.cache != nil {
		return c.cache, nil
	}

	t, err := readTemplates(fsys, t, c.globals)
	if err != nil {
		return nil, err
	}

	t, err = readTemplates(fsys, t, c.locals)
	if err != nil {
		return nil, err
	}

	c.cache = t
	return t, nil
}

func (c *TemplateContext) GetByName(fsys fs.FS) (*template.Template, error) {
	if c.cache != nil {
		return c.cache, nil
	}

	t := template.New(c.Path)

	t, err := readTemplates(fsys, t, c.globals)
	if err != nil {
		return nil, err
	}

	t, err = readTemplates(fsys, t, c.locals)
	if err != nil {
		return nil, err
	}

	c.cache = t
	return t, nil
}

// Get gets the template namespace for this path. The enerated template is cached and
// reused on subsequent calls. This gives the GC a lot of work to do, but it's fine for now.
// TODO: also, we re-parse global components in every directory, which is not very efficient. But to make sure, parts of the template weren't already executed, we do this (could just simply clone the global templates on parsing, but the context is cached so I guess it's ok for now)
func (c *TemplateContext) Get(fsys fs.FS) (*template.Template, error) {
	if c.cache != nil {
		return c.cache, nil
	}

	t := template.New(c.Path)

	t, err := readTemplates(fsys, t, c.globals)
	if err != nil {
		return nil, err
	}

	t, err = readTemplates(fsys, t, c.locals)
	if err != nil {
		return nil, err
	}

	c.cache = t
	return t, nil
}

func readTemplates(fsys fs.FS, t *template.Template, paths map[string]string) (*template.Template, error) {
	for k, v := range paths {
		text, err := fs.ReadFile(fsys, v)
		if err != nil {
			return nil, NewError(FileAccessError, v)
		}

		temp, err := template.New(k).Parse(string(text))
		if err != nil {
			return nil, err
		}

		_, err = t.AddParseTree(k, temp.Tree)
		if err != nil {
			return nil, err
		}
	}

	return t, nil
}
