package main

import (
	"bytes"

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

	tr.Parse()

	e.GET("/*", getEverything(lr, tr))

	e.Logger.Fatal(e.Start("127.0.0.1:1323"))
}

func getEverything(layouts *templating.LayoutRegistry, routes *templating.TemplateRegistry) func(c echo.Context) error {
	return func(c echo.Context) error {

		path := c.Request().URL.Path

		layout, err := layouts.Get(DEFAULT_LAYOUT_NAME)
		if err != nil {
			return c.String(500, err.Error())
		}

		// We clone this here, since driver templates can only be executed once
		layout, err = layout.Clone()
		if err != nil {
			return c.String(500, err.Error())
		}

		err = routes.Add(path, layout)
		// FIXME: we should react do different error types differntly
		if err != nil {
			return c.String(500, err.Error())
		}

		// TODO: we should probably reuse buffers here to avoid allocations
		var buffer bytes.Buffer

		err = layout.Execute(&buffer, nil)
		if err != nil {
			return c.String(500, err.Error())
		}

		return c.HTML(200, buffer.String())

	}
}
