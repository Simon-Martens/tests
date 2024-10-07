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
	// WARNING: The keys of this store are routing paths, NOT filesystem paths
	cache *store.Store[*TemplateContext]
	funcs template.FuncMap
}

func NewTemplateRegistry(routes fs.FS) *TemplateRegistry {
	return &TemplateRegistry{
		routesFS: routes,
		cache:    store.New[*TemplateContext](nil),
		funcs: template.FuncMap{
			"safe": func(s string) template.HTML {
				return template.HTML(s)
			},
		},
	}
}

func (r *TemplateRegistry) Register(path string, fs fs.FS) {
	r.routesFS = merged_fs.MergeMultiple(fs, r.routesFS)

	// We invalidate the cache because the routesFS has changed
	r.cache = store.New[*TemplateContext](nil)
}

func (r *TemplateRegistry) RegisterFuncs(funcs template.FuncMap) {
	for k, v := range funcs {
		r.funcs[k] = v
	}
}

func (r *TemplateRegistry) Parse() {
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

			globals := r.cache.Get(pathabove)
			if globals != nil {
				tc.SetGlobals(globals.GetGlobals())
			}
		}

		tc.Parse(r.routesFS)
		r.cache.Set(url, &tc)

		return nil
	})
}

// INFO: the difference between a trailing slash and no trailing slash is important
// a trailing slash means to get the template context for the directory
// no trailing slash means to get the template context for the component
func (r *TemplateRegistry) Get(path string) (*template.Template, error) {
	tc := r.cache.Get(path)
	if tc == nil {
		return nil, NewError(NoTemplateError, path)
	}

	return tc.Get(r.routesFS)
}

func (r *TemplateRegistry) Add(path string, t *template.Template) (*template.Template, error) {
	tc := r.cache.Get(path)
	if tc == nil {
		return nil, NewError(NoTemplateError, path)
	}

	return tc.Add(r.routesFS, t)
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
