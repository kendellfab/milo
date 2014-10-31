package render

import (
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
)

type MiloRenderer interface {
	RenderTemplates(w http.ResponseWriter, r *http.Request, data map[string]interface{}, tpls ...string)
	RenderJson(w http.ResponseWriter, r *http.Request, data interface{})
	RegisterTemplateFunc(key string, fn interface{})
}

type DefaultMiloRenderer struct {
	templateCache map[string]*template.Template
	tplDir        string
	tplFuncs      map[string]interface{}
	cacheTpls     bool
}

func NewDefaultMiloRenderer(tplDir string, cache bool) MiloRenderer {
	r := &DefaultMiloRenderer{templateCache: make(map[string]*template.Template), tplDir: tplDir, tplFuncs: make(map[string]interface{}), cacheTpls: cache}
	return r
}

// Takes care of rendering templates from file.
func (mr *DefaultMiloRenderer) RenderTemplates(w http.ResponseWriter, r *http.Request, data map[string]interface{}, tpls ...string) {
	if len(tpls) < 1 {
		w.WriteHeader(500)
		w.Write([]byte("Error: Template required!"))
		return
	}

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
		if tpl, ok = mr.templateCache[key]; ok {
			return tpl, nil
		}
	}

	tpl, loadErr = template.New(filepath.Base(tpls[0])).Funcs(mr.tplFuncs).ParseFiles(tpls...)
	if loadErr != nil {
		return nil, loadErr
	}

	if mr.cacheTpls {
		mr.templateCache[key] = tpl
	}
	return tpl, nil
}

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

func (mr *DefaultMiloRenderer) RegisterTemplateFunc(key string, fn interface{}) {
	mr.tplFuncs[key] = fn
}
