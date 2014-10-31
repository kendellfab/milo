package milo

import (
	"code.google.com/p/go.net/websocket"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	BIND_ERR = "bind: address already in use"
)

// This is the default application.
type Milo struct {
	config              *Config
	router              *mux.Router
	subRoutes           map[string]*mux.Router
	logger              MiloLogger
	beforeMiddleware    []http.HandlerFunc
	afterMiddleware     []http.HandlerFunc
	defaultErrorHandler http.HandlerFunc
	notFoundHandler     http.HandlerFunc
}

// Create a new milo app.  Uses the config object.
func NewMiloApp(c *Config) *Milo {
	if c == nil {
		c = DefaultConfig()
	}

	milo := &Milo{
		config:           c,
		router:           mux.NewRouter(),
		subRoutes:        make(map[string]*mux.Router),
		logger:           newDefaultLogger(),
		beforeMiddleware: make([]http.HandlerFunc, 0),
		afterMiddleware:  make([]http.HandlerFunc, 0),
	}
	milo.router.NotFoundHandler = milo
	return milo
}

// Add after request middleware to the global middleware stack.
func (m *Milo) RegisterAfter(mw http.HandlerFunc) {
	m.afterMiddleware = append(m.afterMiddleware, mw)
}

// Add before request middleware to the global middlware stack.
func (m *Milo) RegisterBefore(mw http.HandlerFunc) {
	m.beforeMiddleware = append(m.beforeMiddleware, mw)
}

// Register an error handler for when things go crazy.
func (m *Milo) RegisterDefaultErrorHandler(h http.HandlerFunc) {
	m.defaultErrorHandler = h
}

// Register your own implementation of the milo logger.
func (m *Milo) RegisterLogger(l MiloLogger) {
	m.logger = l
}

// Register a not found handler so you can capture 404 errors.
func (m *Milo) RegisterNotFound(h http.HandlerFunc) {
	m.notFoundHandler = h
}

// Setup a route to be executed when the path is matched, uses the gorilla mux router.
func (m *Milo) Route(path string, methods []string, hf http.HandlerFunc) {

	fn := func(w http.ResponseWriter, r *http.Request) {
		m.runRoute(w, r, hf, path)
	}

	if methods != nil {
		m.router.Path(path).Methods(methods...).HandlerFunc(fn)
	} else {
		m.router.Path(path).HandlerFunc(fn)
	}
}

// Setup sub routes for more efficient routing of requests inside of gorilla mux.
func (m *Milo) SubRoute(prefix, path string, methods []string, hf http.HandlerFunc) {
	var subRouter *mux.Router
	var ok bool

	subRouter, ok = m.subRoutes[prefix]
	if !ok {
		subRouter = m.router.PathPrefix(prefix).Subrouter()
		m.subRoutes[prefix] = subRouter
	}

	fn := func(w http.ResponseWriter, r *http.Request) {
		m.runRoute(w, r, hf, prefix+path)
	}

	if methods != nil {
		subRouter.Path(path).Methods(methods...).HandlerFunc(fn)
	} else {
		subRouter.Path(path).HandlerFunc(fn)
	}
}

// Handling websocket connection.
func (m *Milo) RouteWebsocket(path string, hf func(ws *websocket.Conn)) {
	m.router.Path(path).Handler(websocket.Handler(hf))
}

// Binds and runs the application on the given config port.
func (m *Milo) Run() {
	m.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(m.config.AssetDirectory))))
	if m.config.CatchAll {
		m.router.PathPrefix("/").Handler(http.FileServer(http.Dir(m.config.AssetDirectory)))
	}
	port := m.config.Port

	if m.config.PortIncrement {
		for {
			err := http.ListenAndServe(m.config.GetConnectionString(), m.router)
			if err != nil {
				if strings.Contains(err.Error(), BIND_ERR) {
					port++
					time.Sleep(100 * time.Millisecond)
				} else {
					m.logger.LogError(err)
					os.Exit(2)
				}
			}
		}
	} else {
		if err := http.ListenAndServe(m.config.GetConnectionString(), m.router); err != nil {
			m.logger.LogError(err)
		}
	}
}

// Internal handler for running the route, that way different functions can be exposed but all handled the same.
func (m *Milo) runRoute(w http.ResponseWriter, r *http.Request, hf http.HandlerFunc, path string) {
	defer handleError(m, w, r)
	m.runBeforeMiddleware(w, r)
	// Call registered handler
	hf(w, r)
	m.runAfterMiddlware(w, r)
	// Writing out a request log
	m.logger.LogInterfaces("Path:", path)
}

// Runs before middleware.
func (m *Milo) runBeforeMiddleware(w http.ResponseWriter, r *http.Request) {
	// Running before middleware
	for _, mdw := range m.beforeMiddleware {
		mdw(w, r)
	}
}

// Runs after middleware
func (m *Milo) runAfterMiddlware(w http.ResponseWriter, r *http.Request) {
	// Run through after middleware
	for _, mdw := range m.afterMiddleware {
		mdw(w, r)
	}
}

// ServeHTTP as passed into the notfoundhandler.
func (m *Milo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	if m.notFoundHandler == nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - MILO: Route not found."))
	} else {
		m.notFoundHandler(w, r)
	}
}

// Internal error handler for the multiple places that could cause a crashing error.
func handleError(m *Milo, w http.ResponseWriter, r *http.Request) {
	if err := recover(); err != nil {
		m.logger.LogInterfaces("milo.Route", r.URL.RequestURI(), err, mux.Vars(r))
		m.logger.LogStackTrace()

		if m.defaultErrorHandler != nil {
			m.defaultErrorHandler(w, r)
		} else {
			http.Error(w, "500 - Internal Server Error.", http.StatusInternalServerError)
		}
	}
}
