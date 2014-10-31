package main

import (
	"github.com/kendellfab/milo"
	"github.com/kendellfab/milo/render"
	"net/http"
)

var rend render.MiloRenderer

func main() {

	rend = render.NewDefaultMiloRenderer("tpls", false)

	app := milo.NewMiloApp(nil)

	app.Route("/", []string{"Get"}, handleRoot)
	app.Run()
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	rend.RenderTemplates(w, r, nil, "index.tpl")
}
