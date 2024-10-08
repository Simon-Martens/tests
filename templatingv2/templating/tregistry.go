package templating

import (
	"errors"
	"html/template"
	"io/fs"
	"os"
	"strings"

	"github.com/pocketbase/pocketbase/tools/store"
	"github.com/yalue/merged_fs"
)

var InvalidTemplateError = errors.New("invalid template")

type TemplateRegistry struct {
	routesFS fs.FS
	parsed   bool
	// INFO: Template & cache keys are directory routing paths, with '/' as root
	templates map[string]TemplateContext
	cache     *store.Store[*template.Template]
	funcs     template.FuncMap
}

func NewTemplateRegistry(routes fs.FS) *TemplateRegistry {
	return &TemplateRegistry{
		routesFS:  routes,
		parsed:    false,
		templates: make(map[string]TemplateContext),
		cache:     store.New[*template.Template](nil),
		funcs: template.FuncMap{
			"safe": func(s string) template.HTML {
				return template.HTML(s)
			},
		},
	}
}

// This returns a new TemplateRegistry with the new fs added to the existing fs,
// merging with the existing FS, possibly overwriting existing files.
func (r *TemplateRegistry) Register(path string, fs fs.FS) *TemplateRegistry {
	return NewTemplateRegistry(merged_fs.MergeMultiple(fs, r.routesFS))
}

func (r *TemplateRegistry) RegisterFuncs(funcs template.FuncMap) {
	for k, v := range funcs {
		r.funcs[k] = v
	}
}

func (r *TemplateRegistry) Parse() {
	// INFO: setting parsed first is important, as it avoids infinite loops in Add() below
	r.parsed = true

	fs.WalkDir(r.routesFS, ".", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			return nil
		}

		url := FSPathToPath(path)
		tc := NewTemplateContext(url)

		if path != "." {
			pathelem := strings.Split(path, string(os.PathSeparator))
			pathabove := strings.Join(pathelem[:len(pathelem)-1], string(os.PathSeparator))
			pathabove = FSPathToPath(pathabove)

			globals, ok := r.templates[pathabove]
			if ok {
				tc.SetGlobals(globals.GetGlobals())
			}
		}

		tc.Parse(r.routesFS)

		r.templates[url] = tc

		return nil
	})
}

func (r *TemplateRegistry) Add(path string, t *template.Template) error {
	tc := r.cache.Get(path)
	if tc == nil {

		return NewError(NoTemplateError, path)
	}

	temp, err := tc.Get(r.routesFS)
	if err != nil {
		return err
	}

	for _, st := range temp.Templates() {
		_, err = t.AddParseTree(st.Name(), st.Tree)
		if err != nil {
			return err
		}
	}

	return nil
}

// TODO: get for a specific component
func (r *TemplateRegistry) Get(path string) error {
	return nil
}

func PathToFSPath(p string) string {
	if p == "/" {
		return "."
	}

	p = strings.TrimPrefix(p, "/")
	p = strings.TrimSuffix(p, "/")

	return p
}

func FSPathToPath(p string) string {
	if p == "." {
		return "/"
	}

	p = strings.TrimPrefix(p, ".")

	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}

	if !strings.HasSuffix(p, "/") {
		p = p + "/"
	}

	return p
}
