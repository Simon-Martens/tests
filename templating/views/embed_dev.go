//go:build dev

package views

import (
	"os"
)

const (
	STATIC_FILEPATH = "./views/assets"
	ROUTES_FILEPATH = "./views/routes"
	LAYOUT_FILEPATH = "./views/layouts"
)

var StaticFS = os.DirFS(STATIC_FILEPATH)
var RoutesFS = os.DirFS(ROUTES_FILEPATH)
var LayoutFS = os.DirFS(LAYOUT_FILEPATH)
