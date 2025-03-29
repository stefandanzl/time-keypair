package api

import (
	"data-cron-server/auth"
	"data-cron-server/config"
	"data-cron-server/cron"
	"net/http"
	"strings"
)

// Router handles HTTP routing
type Router struct {
	mux        *http.ServeMux
	config     *config.Config
	scheduler  *cron.Scheduler
	auth       *auth.Authenticator
}

// NewRouter creates a new router
func NewRouter(cfg *config.Config, scheduler *cron.Scheduler, superAdminKey string) http.Handler {
	router := &Router{
		mux:       http.NewServeMux(),
		config:    cfg,
		scheduler: scheduler,
		auth:      auth.NewAuthenticator(cfg, superAdminKey),
	}

	// Setup routes
	router.setupAdminRoutes()
	router.setupCronRoutes()
	router.setupDataRoutes()
	router.setupHealthCheck()

	return router.mux
}

// setupAdminRoutes sets up admin routes
func (r *Router) setupAdminRoutes() {
	// Admin routes - require super admin authentication
	adminHandler := r.auth.RequireSuperAdmin(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		path := req.URL.Path

		// Route based on path pattern
		switch {
		case matchPath(path, "/admin/*/users"):
			r.handleAdminUsers(w, req)
		case matchPath(path, "/admin/*/config"):
			r.handleAdminConfig(w, req)
		default:
			http.NotFound(w, req)
		}
	}))

	r.mux.Handle("/admin/", adminHandler)
}

// setupCronRoutes sets up cron routes
func (r *Router) setupCronRoutes() {
	// Cron routes - require user authentication
	cronHandler := r.auth.RequireUser(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		path := req.URL.Path

		// Route based on path pattern
		switch {
		case matchPath(path, "/status/*"):
			r.handleStatus(w, req)
		case matchPath(path, "/cron/*/jobs"):
			r.handleCronJobs(w, req)
		case matchPath(path, "/cron/*/job/*"):
			r.handleCronJob(w, req)
		default:
			http.NotFound(w, req)
		}
	}))

	r.mux.Handle("/status/", cronHandler)
	r.mux.Handle("/cron/", cronHandler)
}

// setupDataRoutes sets up data routes
func (r *Router) setupDataRoutes() {
	// Data routes - require user authentication
	dataHandler := r.auth.RequireUser(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		path := req.URL.Path

		// Route based on path pattern
		switch {
		case matchPath(path, "/data/*/keys"):
			r.handleDataKeys(w, req)
		case matchPath(path, "/data/*/*"):
			r.handleData(w, req)
		default:
			http.NotFound(w, req)
		}
	}))

	r.mux.Handle("/data/", dataHandler)
}

// setupHealthCheck sets up health check route
func (r *Router) setupHealthCheck() {
	r.mux.HandleFunc("/health", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}

// matchPath checks if a path matches a pattern
func matchPath(path, pattern string) bool {
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")

	// Special case for user deletion path: /admin/*/users/user1
	if len(patternParts) == 3 && len(pathParts) == 4 && 
	   patternParts[0] == "admin" && patternParts[2] == "users" && 
	   pathParts[0] == "admin" && pathParts[2] == "users" {
		return true
	}

	if len(pathParts) != len(patternParts) {
		return false
	}

	for i, patternPart := range patternParts {
		if patternPart == "*" {
			continue
		}
		if pathParts[i] != patternPart {
			return false
		}
	}

	return true
}

// ServeHTTP implements the http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
