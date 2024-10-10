package templating

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
)

type TemplateContext struct {
	//  WARNING: Path is a URL path, NOT a filesystem path
	Path string
	//  WARNING: The keys of these maps are template names, NOT filesystem paths
	// The values are FS paths absolute from the root directory of the templates FS
	locals  map[string]string
	globals map[string]string
}

func NewTemplateContext(path string) TemplateContext {
	return TemplateContext{
		Path:    path,
		locals:  make(map[string]string),
		globals: make(map[string]string),
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
			// INFO: components in the components directory can be overwritten
			// by components in the base directory down below
			// TODO: Maybe allow for subdirectories in the components directory?
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
	// Make sure to call this appopriately before or after Parse(), depending on your use case
	for k, v := range globals {
		c.globals[k] = v
	}
	return nil
}

func (c *TemplateContext) GetGlobals() map[string]string {
	return c.globals
}

func (c *TemplateContext) Get(fsys fs.FS) (*template.Template, error) {
	t, err := readTemplates(fsys, nil, c.globals)
	if err != nil {
		return nil, err
	}

	t, err = readTemplates(fsys, t, c.locals)
	if err != nil {
		return nil, err
	}

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

		if t == nil {
			t = temp
			continue
		}

		for _, template := range temp.Templates() {
			_, err = t.AddParseTree(template.Name(), template.Tree)
			if err != nil {
				return nil, err
			}
		}

		_, err = t.AddParseTree(temp.Name(), temp.Tree)
		if err != nil {
			return nil, err
		}

	}

	return t, nil
}
