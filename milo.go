package milo

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"log"

	"github.com/gorilla/mux"
	"golang.org/x/net/websocket"
)

const (
	BIND_ERR = "bind: address already in use"
)

type MiloMiddlware func(w http.ResponseWriter, r *http.Request) bool

// This is the default application.
type Milo struct {
	bind                string
	port                int
	portIncrement       bool
	router              *mux.Router
	subRoutes           map[string]*mux.Router
	logger              MiloLogger
	beforeMiddleware    []MiloMiddlware
	afterMiddleware     []MiloMiddlware
	defaultErrorHandler http.HandlerFunc
	notFoundHandler     http.HandlerFunc
}

// Create a new milo app.  Uses the config object.
func NewMiloApp(opts ...func(*Milo) error) *Milo {
	milo := &Milo{
		router:           mux.NewRouter(),
		subRoutes:        make(map[string]*mux.Router),
		logger:           newDefaultLogger(),
		beforeMiddleware: make([]MiloMiddlware, 0),
		afterMiddleware:  make([]MiloMiddlware, 0),
	}
	milo.router.NotFoundHandler = milo
	milo.port = 7000
	for _, opt := range opts {
		err := opt(milo)
		if err != nil {
			milo.logger.LogFatal(err)
		}
	}
	return milo
}

// Add after request middleware to the global middleware stack.
func (m *Milo) RegisterAfter(mw MiloMiddlware) {
	m.afterMiddleware = append(m.afterMiddleware, mw)
}

// Add before request middleware to the global middlware stack.
func (m *Milo) RegisterBefore(mw MiloMiddlware) {
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

// Setup a route to be executed when the specific path prefix is matched, uses the gorilla mux router.
func (m *Milo) PathPrefix(path string, methods []string, hf http.HandlerFunc) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		m.runRoute(w, r, hf, path)
	}

	if methods != nil {
		m.router.PathPrefix(path).Methods(methods...).HandlerFunc(fn)
	} else {
		m.router.PathPrefix(path).HandlerFunc(fn)
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

// Handle assets rooted in different directories.
func (m *Milo) RouteAsset(prefix, dir string) {
	m.router.PathPrefix(prefix).Handler(http.FileServer(http.Dir(dir)))
}

// Handle assets rooted in different directories, strips prefix.
func (m *Milo) RouteAssetStripPrefix(prefix, dir string) {
	m.router.PathPrefix(prefix).Handler(http.StripPrefix(prefix, http.FileServer(http.Dir(dir))))
}

// Binds and runs the application on the given config port.
func (m *Milo) Run() {
	port := m.port
	if m.portIncrement {
		for {
			log.Println("Connection:", m.getConnectionString())
			srv := m.getHTTPServer()
			err := srv.ListenAndServe()
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
		log.Println("Connection:", m.getConnectionString())
		srv := m.getHTTPServer()
		if err := srv.ListenAndServe(); err != nil {
			m.logger.LogError(err)
		}
	}
}

func (m *Milo) getHTTPServer() *http.Server {
	return &http.Server{
		Addr:         m.getConnectionString(),
		Handler:      m.router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
}

// Internal handler for running the route, that way different functions can be exposed but all handled the same.
func (m *Milo) runRoute(w http.ResponseWriter, r *http.Request, hf http.HandlerFunc, path string) {
	defer handleError(m, w, r)
	// Writing out a request log
	m.logger.LogInterfaces("Path:", path)
	shouldContinue := m.runBeforeMiddleware(w, r)
	// Something happend in the global middleware and we don't want to continue
	// This is under the assumption that the middleware handled everything.
	if !shouldContinue {
		return
	}
	// Call registered handler
	hf(w, r)
	m.runAfterMiddlware(w, r)
}

// Runs before middleware.
func (m *Milo) runBeforeMiddleware(w http.ResponseWriter, r *http.Request) bool {
	// Running before middleware
	for _, mdw := range m.beforeMiddleware {
		if resp := mdw(w, r); !resp {
			return resp
		}
	}
	return true
}

// Runs after middleware
func (m *Milo) runAfterMiddlware(w http.ResponseWriter, r *http.Request) {
	// Run through after middleware
	for _, mdw := range m.afterMiddleware {
		mdw(w, r)
	}
}

// Get the connection string from the config object.
func (m *Milo) getConnectionString() string {
	return fmt.Sprintf("%s:%d", m.bind, m.port)
}

// ServeHTTP as passed into the notfoundhandler.
func (m *Milo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.logger.Log("404 - Route not found.  " + r.RequestURI)
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

// Stringer implementation
func (m *Milo) String() string {
	return fmt.Sprintf("Bind: %s Port: %d Port Increment: %t", m.bind, m.port, m.portIncrement)
}
