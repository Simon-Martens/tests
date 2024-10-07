package templating

import (
	"fmt"
	"html/template"
	"io/fs"
	"path/filepath"

	"github.com/pocketbase/pocketbase/tools/store"
	"github.com/yalue/merged_fs"
)

const ROOT_LAYOUT_NAME = "root"

type LayoutRegistry struct {
	layoutsFS fs.FS
	cache     *store.Store[*TemplateContext]
	funcs     template.FuncMap
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

func (r *LayoutRegistry) Parse() {
	fmt.Println("Parsing layouts")
	rootcontext := NewTemplateContext(".")
	rootcontext.Parse(r.layoutsFS)
	globals := rootcontext.GetGlobals()

	fs.WalkDir(r.layoutsFS, ".", func(path string, d fs.DirEntry, err error) error {

		if path == "." {
			return nil
		}

		if !d.IsDir() {
			return nil
		}

		fmt.Println("Parsing layout path: " + path)

		pathelements := filepath.SplitList(path)
		// We dont parse subdirectories ATM
		if len(pathelements) > 1 {
			return nil
		}

		context := NewTemplateContext(path)
		context.Parse(r.layoutsFS)
		context.SetGlobals(globals)

		r.cache.Set(path, context)

		return nil
	})
}

func (r *LayoutRegistry) Get(path string) *TemplateContext {
	return r.cache.Get(path)
}
