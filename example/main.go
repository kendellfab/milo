package main

import (
	"github.com/kendellfab/milo"
	"github.com/kendellfab/milo/render"
	"log"
	"net/http"
	"time"
)

var rend render.MiloRenderer

func main() {
	log.Println("Milo")

	app := milo.NewMiloApp(milo.SetPort(3030))
	log.Println(app)

	app.RegisterBefore(func(w http.ResponseWriter, r *http.Request) bool {
		log.Println("First Global Before Middleware")
		r.ParseForm()
		if r.Form.Get("redirect") != "" {
			rend.Redirect(w, r, "/landing", http.StatusSeeOther)
			log.Println("And We're Redirecting")
			return false
		}
		return true
	})

	app.RegisterBefore(func(w http.ResponseWriter, r *http.Request) bool {
		log.Println("Second Global Before Middleware")
		return true
	})

	rend = render.NewDefaultMiloRenderer("tpls", false)

	app.Route("/", []string{"Get"}, handleRoot)
	app.Route("/demo", []string{"Get"}, miloMiddleware(redirectMiddleware(handleDemo)))
	app.Route("/landing", []string{"Get"}, miloMiddleware(handleLanding))

	app.RouteAsset("/css", "static")
	app.RouteAsset("/", "./")
	app.Run()

}

func miloMiddleware(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Milo Middleware")
		fn(w, r)
	}
}

func redirectMiddleware(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Redirect middleware")
		if time.Now().Unix()%2 == 0 {
			log.Println("Caught Redirect Trap")
			rend.Redirect(w, r, "/landing", http.StatusSeeOther)
			return
		}
		fn(w, r)
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	rend.RenderTemplates(w, r, nil, "index.tpl")
}

func handleDemo(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Demo"))
}

func handleLanding(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Landing"))
}
