package milo

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
)

// An interface to define a way to get config items always into template rendering
type Configer interface {
	GetConfig(key string) interface{}
}

// Default milo renderer that can cache templates, sets a base template directory.
type Renderer struct {
	templateCache map[string]*template.Template
	tplDir        string
	tplFuncs      map[string]interface{}
	cacheTpls     bool
	configer      Configer
	sync.RWMutex
}

// Create a new default milo renderer.
func NewRenderer(tplDir string, cache bool, configer Configer) *Renderer {
	r := &Renderer{templateCache: make(map[string]*template.Template), tplDir: tplDir, tplFuncs: make(map[string]interface{}), cacheTpls: cache, configer: configer}
	r.tplFuncs["host"] = Host
	r.tplFuncs["marshal"] = Marshal
	r.tplFuncs["partial"] = r.Partial
	r.tplFuncs["title"] = Title
	r.tplFuncs["gravatar"] = Gravatar
	return r
}

// Takes care of rendering templates from file, passes a status 200.
func (mr *Renderer) RenderTemplates(w http.ResponseWriter, r *http.Request, data map[string]interface{}, tpls ...string) {
	mr.RenderTemplatesCode(w, r, 200, data, tpls...)
}

// Takes the care of rendering templates, with an explicit status code.
func (mr *Renderer) RenderTemplatesCode(w http.ResponseWriter, r *http.Request, code int, data map[string]interface{}, tpls ...string) {
	if len(tpls) < 1 {
		w.WriteHeader(500)
		w.Write([]byte("Error: Template required!"))
		return
	}
	defaults := make(map[string]interface{})
	if mr.configer != nil {
		defaults["config"] = mr.configer
	}
	for k, v := range data {
		defaults[k] = v
	}

	list := make([]string, 0)
	for _, elem := range tpls {
		list = append(list, filepath.Join(mr.tplDir, elem))
	}

	if tpl, loadErr := mr.acquireTemplate(strings.Join(tpls, ""), list...); loadErr != nil {
		w.WriteHeader(500)
		w.Write([]byte(loadErr.Error()))
	} else {
		var doc bytes.Buffer
		err := tpl.Execute(&doc, defaults)
		if err == nil {
			w.WriteHeader(code)
			w.Write(doc.Bytes())
		} else {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
		}
	}
}

// Unexported method to help handle template parsing.  If the cache template bool is set on the config
// struct this method with look in the cache & load the cache upon subsequent encounters.
// This should lower disk access penalties useful for production instances.
func (mr *Renderer) acquireTemplate(key string, tpls ...string) (*template.Template, error) {
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
func (mr *Renderer) RenderJson(w http.ResponseWriter, r *http.Request, data interface{}) {
	if data, err := json.Marshal(data); err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
	} else {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

func (mr *Renderer) RenderError(w http.ResponseWriter, r *http.Request, code int, message string) {
	w.WriteHeader(code)
	w.Write([]byte(message))
}

func (mr *Renderer) RenderMessage(w http.ResponseWriter, r *http.Request, message string) {
	w.WriteHeader(200)
	w.Write([]byte(message))
}

// Register a template function with the MiloRenderer
func (mr *Renderer) RegisterTemplateFunc(key string, fn interface{}) {
	mr.tplFuncs[key] = fn
}

// Setup an http redirect on the request.
func (mr *Renderer) Redirect(w http.ResponseWriter, r *http.Request, url string, code int) {
	http.Redirect(w, r, url, code)
}

// A template function which can include a partial template.
func (mr *Renderer) Partial(name string, payload interface{}) (template.HTML, error) {
	var buff bytes.Buffer
	path := filepath.Join(mr.tplDir, "partials", name)

	tpl, loadErr := mr.acquireTemplate(path, path)
	if loadErr != nil {
		return "", loadErr
	}

	execErr := tpl.Execute(&buff, payload)

	if execErr != nil {
		return "", execErr
	}

	return template.HTML(string(buff.Bytes())), nil
}
