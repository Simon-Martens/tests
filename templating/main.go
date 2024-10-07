package main

import (
	"bytes"
	"fmt"

	"github.com/Simon-Martens/misc_tests/templating"
	"github.com/Simon-Martens/misc_tests/views"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

	e.GET("/*", getEverything(lr, tr), middleware.RemoveTrailingSlash())

	e.Logger.Fatal(e.Start("127.0.0.1:1323"))
}

func getEverything(layouts *templating.LayoutRegistry, routes *templating.TemplateRegistry) func(c echo.Context) error {
	return func(c echo.Context) error {
		t := routes.Get(c.Request().RequestURI)

		fmt.Printf("Got a request for: %s\n", c.Request().RequestURI)

		if t == nil {
			return c.String(404, "Not found")
		}

		l, err := layouts.Get(DEFAULT_LAYOUT_NAME).GetRoot()
		if err != nil {
			fmt.Println("Could not get root")
			fmt.Println(err)
		}

		l, _ = l.Clone()

		b, err := t.GetBody()

		if err != nil {
			fmt.Println(err)
			fmt.Println("Could not get body")
		} else {
			fmt.Println("Got body")
			_, err = l.AddParseTree("body", b.Tree)
			if err != nil {
				fmt.Println("Could not add body")
				fmt.Println(err)
			}
		}

		h, err := t.GetHead()
		if err != nil {
			fmt.Println(err)
			fmt.Println("Could not get head")
		} else {
			l.AddParseTree("head", h.Tree)
		}

		var buff bytes.Buffer

		for _, g := range l.Templates() {
			fmt.Println("Executing template: " + g.Name())
		}

		l.Execute(&buff, "hello")

		return c.HTMLBlob(200, buff.Bytes())
	}
}
