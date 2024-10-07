package templating

import (
	"errors"
	"html/template"
	"io/fs"

	"github.com/pocketbase/pocketbase/tools/store"
	"github.com/yalue/merged_fs"
)

const ROOT_LAYOUT_NAME = "root"

var NoTemplateError = errors.New("No template found for this path")

type LayoutRegistry struct {
	layoutsFS fs.FS
	// INFO: We store the cache keys as URL paths, but actually for compatibility
	// with the TemplateContext, as really they should be FS paths
	cache *store.Store[*TemplateContext]
	funcs template.FuncMap
}

func NewLayoutRegistry(routes fs.FS) *LayoutRegistry {
	return &LayoutRegistry{
		layoutsFS: routes,
		cache:     store.New[*TemplateContext](nil),
		funcs: template.FuncMap{
			"safe": func(s string) template.HTML {
				return template.HTML(s)
			},
		},
	}
}

func (r *LayoutRegistry) Register(path string, fs fs.FS) {
	r.layoutsFS = merged_fs.MergeMultiple(fs, r.layoutsFS)
	r.cache = store.New[*TemplateContext](nil)
}

func (r *LayoutRegistry) RegisterFuncs(funcs template.FuncMap) {
	for k, v := range funcs {
		r.funcs[k] = v
	}
}

func (r *LayoutRegistry) Parse() error {
	rootcontext := NewTemplateContext(".")
	rootcontext.Parse(r.layoutsFS)
	globals := rootcontext.GetGlobals()

	entries, err := fs.ReadDir(r.layoutsFS, ".")
	if err != nil {
		return NewError(FileAccessError, ".")
	}

	for _, e := range entries {
		if !e.IsDir() || e.Name() == TEMPLATE_COMPONENT_DIRECTORY {
			continue
		}

		url := FSPathToPath(e.Name())
		context := NewTemplateContext(url)
		context.SetGlobals(globals)
		context.Parse(r.layoutsFS)

		r.cache.Set(url, &context)
	}

	return nil
}

func (r *LayoutRegistry) Get(name string) (*template.Template, error) {
	url := FSPathToPath(name)
	tc := r.cache.Get(url)
	if tc == nil {
		return nil, NewError(NoTemplateError, name)
	}

	availables, err := tc.Get(r.layoutsFS)
	if err != nil {
		return nil, err
	}

	t := availables.Lookup(ROOT_LAYOUT_NAME)
	if t == nil {
		return nil, NewError(NoTemplateError, name)
	}

	return t, nil

}
