package render

import (
	"encoding/json"
	"html/template"
	"net/http"
)

// Get the host for the given http request.
// Can be used like {{ host .request }}
func Host(r *http.Request) string {
	host := r.URL.Host
	if host == "" {
		host = r.Host
	}
	return host
}

// Get a json encoding of an object from the backend.
// Can be used like {{ marshal .user }}
func Marshal(v interface{}) template.JS {
	a, _ := json.Marshal(v)
	return template.JS(a)
}
