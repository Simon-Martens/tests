package templating

import (
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"strings"

	"github.com/pocketbase/pocketbase/tools/store"
	"github.com/yalue/merged_fs"
)

var InvalidTemplateError = errors.New("invalid template")

type TemplateRegistry struct {
	routesFS fs.FS
	cache    *store.Store[*TemplateContext]
	funcs    template.FuncMap
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
	r.cache = store.New[*TemplateContext](nil)
}

func (r *TemplateRegistry) RegisterFuncs(funcs template.FuncMap) {
	for k, v := range funcs {
		r.funcs[k] = v
	}
}

func (r *TemplateRegistry) Parse() {
	rootcontext := NewTemplateContext(".")
	rootcontext.Parse(r.routesFS)

	r.cache.Set("/", rootcontext)

	fs.WalkDir(r.routesFS, ".", func(path string, d fs.DirEntry, err error) error {

		if !d.IsDir() {
			return nil
		}

		fmt.Println("Parsing template directory: " + path)

		context := NewTemplateContext(path)
		err = context.Parse(r.routesFS)
		if err != nil {
			fmt.Println("Could not parse " + path)
			fmt.Println(err)
			return nil
		}

		pathelements := strings.Split(path, "/")

		var head string
		if len(path) > 1 {
			head = strings.Join(pathelements[0:len(pathelements)-1], "/")
		}

		if head == "" {
			head = "/"
		}

		fmt.Println("Template Head: " + head)

		// We trust that the parent is parsed before the child
		parent := r.cache.Get(head)
		if parent == nil {
			fmt.Println("Could not find parent for " + path)
			return nil
		}

		err = context.SetGlobals(parent.GetGlobals())
		if err != nil {
			fmt.Println("Could not set globals for " + path)
			return nil
		}

		r.cache.Set("/"+path, context)

		return nil
	})
}

func (r *TemplateRegistry) Get(path string) *TemplateContext {
	return r.cache.Get(path)
}
