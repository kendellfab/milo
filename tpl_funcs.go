package milo

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
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

// Title gets the title for the word.
func Title(word string) string {
	return strings.Title(word)
}

// Gravatar builds links to gravatar.
func Gravatar(email string, s int) string {
	return fmt.Sprintf("https://www.gravatar.com/avatar/%x?s=%d", md5.Sum([]byte(email)), s)
}
