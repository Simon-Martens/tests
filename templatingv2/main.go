package main

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/Simon-Martens/misc_tests/templating"
	"github.com/Simon-Martens/misc_tests/views"
	"github.com/labstack/echo/v4"
)

const ROOT_LAYOUT_NAME = "root"
const DEFAULT_LAYOUT_NAME = "default"

var lr *templating.LayoutRegistry
var tr *templating.TemplateRegistry

func main() {
	e := echo.New()

	lr = templating.NewLayoutRegistry(views.LayoutFS)
	tr = templating.NewTemplateRegistry(views.RoutesFS)

	lr.Parse()
	tr.Parse()

	e.GET("/*", getEverything(lr, tr))

	e.Logger.Fatal(e.Start("127.0.0.1:1323"))
}

func getEverything(layouts *templating.LayoutRegistry, routes *templating.TemplateRegistry) func(c echo.Context) error {
	return func(c echo.Context) error {

		path := c.Request().URL.Path

		// TODO: getting single components is not supported ATM
		if !strings.HasSuffix(path, "/") {
			path += "/"
		}

		layout, err := layouts.Get(DEFAULT_LAYOUT_NAME)
		if err != nil {
			return c.String(500, err.Error())
		}

		fmt.Println(layout.Name())

		// We clone this here, since driver templates can only be executed once
		layout, err = layout.Clone()
		if err != nil {
			return c.String(500, err.Error())
		}

		// tc, err := routes.Add(path, layout)

		tc, err := routes.Get(path)
		if err != nil {
			return c.String(500, err.Error())
		}

		// FIXME: Here we don't know the behaviour of the htm/template package
		// Are there sub-templates? Do I need a for loop to add them to the global name space?
		// INFO: since we don't clone the template, we can't execute it multiple times, right?
		// INFO: AddParseTree deletes all sub-templates of the root template.
		for _, st := range tc.Templates() {
			_, err = layout.AddParseTree(st.Name(), st.Tree)
			if err != nil {
				return c.String(500, err.Error())
			}
		}

		// TODO: we should probably reuse buffers here to avoid allocations
		var buffer bytes.Buffer

		err = tc.Execute(&buffer, nil)
		if err != nil {
			return c.String(500, err.Error())
		}

		return c.HTML(200, buffer.String())

	}
}
