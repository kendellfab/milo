package render

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
)

// An interface of suggested methods for rendering output in Go
type MiloRenderer interface {
	RenderTemplates(w http.ResponseWriter, r *http.Request, data map[string]interface{}, tpls ...string)
	RenderJson(w http.ResponseWriter, r *http.Request, data interface{})
	RegisterTemplateFunc(key string, fn interface{})
	Redirect(w http.ResponseWriter, r *http.Request, url string, code int)
}

// Default milo renderer that can cache templates, sets a base template directory.
type DefaultMiloRenderer struct {
	templateCache map[string]*template.Template
	tplDir        string
	tplFuncs      map[string]interface{}
	cacheTpls     bool
	sync.RWMutex
}

// Create a new default milo renderer.
func NewDefaultMiloRenderer(tplDir string, cache bool) MiloRenderer {
	r := &DefaultMiloRenderer{templateCache: make(map[string]*template.Template), tplDir: tplDir, tplFuncs: make(map[string]interface{}), cacheTpls: cache}
	r.tplFuncs["host"] = Host
	r.tplFuncs["marshal"] = Marshal
	return r
}

// Takes care of rendering templates from file.
func (mr *DefaultMiloRenderer) RenderTemplates(w http.ResponseWriter, r *http.Request, data map[string]interface{}, tpls ...string) {
	if len(tpls) < 1 {
		w.WriteHeader(500)
		w.Write([]byte("Error: Template required!"))
		return
	}

	log.Println("Rendering the templates")

	list := make([]string, 0)
	for _, elem := range tpls {
		list = append(list, filepath.Join(mr.tplDir, elem))
	}

	if tpl, loadErr := mr.acquireTemplate(strings.Join(tpls, ""), list...); loadErr != nil {
		w.WriteHeader(500)
		w.Write([]byte(loadErr.Error()))
	} else {
		tpl.Execute(w, data)
	}
}

// Unexported method to help handle template parsing.  If the cache template bool is set on the config
// struct this method with look in the cache & load the cache upon subsequent encounters.
// This should lower disk access penalties useful for production instances.
func (mr *DefaultMiloRenderer) acquireTemplate(key string, tpls ...string) (*template.Template, error) {
	var tpl *template.Template
	var loadErr error
	var ok bool

	if mr.cacheTpls {
		mr.RLock()
		tpl, ok = mr.templateCache[key]
		mr.RUnlock()
		if ok {
			return tpl, nil
		}
	}

	tpl, loadErr = template.New(filepath.Base(tpls[0])).Funcs(mr.tplFuncs).ParseFiles(tpls...)
	if loadErr != nil {
		return nil, loadErr
	}

	if mr.cacheTpls {
		mr.Lock()
		mr.templateCache[key] = tpl
		mr.Unlock()
	}
	return tpl, nil
}

// Render json output
func (mr *DefaultMiloRenderer) RenderJson(w http.ResponseWriter, r *http.Request, data interface{}) {
	if data, err := json.Marshal(data); err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
	} else {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

// Register a template function with the MiloRenderer
func (mr *DefaultMiloRenderer) RegisterTemplateFunc(key string, fn interface{}) {
	mr.tplFuncs[key] = fn
}

// Setup an http redirect on the request.
func (mr *DefaultMiloRenderer) Redirect(w http.ResponseWriter, r *http.Request, url string, code int) {
	http.Redirect(w, r, url, code)
}
