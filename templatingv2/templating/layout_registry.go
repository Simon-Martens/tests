package templating

import (
	"html/template"
	"io/fs"

	"github.com/pocketbase/pocketbase/tools/store"
	"github.com/yalue/merged_fs"
)

// TODO: Implement Handler interface, maybe in template? But then template would need to know about the layout registry
// Function signature: func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request)
// A static Handler could incoporate both the layout registry and the template registry and serve templates that dont need any data
type LayoutRegistry struct {
	layoutsFS fs.FS
	parsed    bool
	// INFO: Layout & cache keys are template directory names
	layouts map[string]TemplateContext
	cache   *store.Store[*template.Template]
	funcs   template.FuncMap
}

func NewLayoutRegistry(routes fs.FS) *LayoutRegistry {
	return &LayoutRegistry{
		layoutsFS: routes,
		parsed:    false,
		layouts:   make(map[string]TemplateContext),
		cache:     store.New[*template.Template](nil),
		funcs: template.FuncMap{
			"safe": func(s string) template.HTML {
				return template.HTML(s)
			},
		},
	}
}

// NOTE: Upon registering a new layout dir, we return a new LayoutRegistry
func (r *LayoutRegistry) Register(fs fs.FS) *LayoutRegistry {
	return NewLayoutRegistry(merged_fs.MergeMultiple(fs, r.layoutsFS))
}

// TODO: Funcs are not used in executing the templates yet
func (r *LayoutRegistry) RegisterFuncs(funcs template.FuncMap) {
	for k, v := range funcs {
		r.funcs[k] = v
	}
}

func (r *LayoutRegistry) Parse() error {
	// INFO: setting parsed first is important, as it avoids infinit loops in Get() below
	r.parsed = true

	rootcontext := NewTemplateContext(".")
	err := rootcontext.Parse(r.layoutsFS)
	if err != nil {
		return err
	}

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

		r.layouts[e.Name()] = context
	}

	return nil
}

// TODO: this pattern here might not be thread save, since parsed can be false for concurrent handlers when calling Get()
// Maybe use the std/library once package to ensure Parse() is only called once
func (r *LayoutRegistry) Get(name string) (*template.Template, error) {
	cached := r.cache.Get(name)
	// This makes sense bc it is very likely cached on most requests
	if cached != nil {
		return cached, nil
	}

	context, ok := r.layouts[name]
	if !ok {
		// This makes only sense to check ONCE, we will have to call NewError(NoTemplateError, name) 99.9999999999% of the requests
		if !r.parsed {
			err := r.Parse()
			if err != nil {
				return nil, err
			}

			return r.Get(name)
		}

		return nil, NewError(NoTemplateError, name)
	}

	t, err := context.Get(r.layoutsFS)
	if err != nil {
		return nil, err
	}

	r.cache.Set(name, t)

	return context.Get(r.layoutsFS)
}
