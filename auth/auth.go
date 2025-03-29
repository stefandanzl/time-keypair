package auth

import (
	"data-cron-server/config"
	"errors"
	"net/http"
)

// Common errors
var (
	ErrInvalidKey        = errors.New("invalid authentication key")
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidSuperAdmin = errors.New("invalid super admin key")
)

// Authenticator provides authentication functionality
type Authenticator struct {
	config        *config.Config
	superAdminKey string
}

// NewAuthenticator creates a new authenticator
func NewAuthenticator(cfg *config.Config, superAdminKey string) *Authenticator {
	return &Authenticator{
		config:        cfg,
		superAdminKey: superAdminKey,
	}
}

// AuthenticateSuperAdmin authenticates a super admin request
func (a *Authenticator) AuthenticateSuperAdmin(key string) error {
	if key != a.superAdminKey {
		return ErrInvalidSuperAdmin
	}
	return nil
}

// AuthenticateUser authenticates a user request
func (a *Authenticator) AuthenticateUser(user, key string) error {
	if userData := a.config.GetUser(user); userData == nil {
		return ErrUserNotFound
	}
	
	// In this simplified authentication model, the key is the user's ID
	// In a real system, you would compare against a stored key
	if user != key {
		return ErrInvalidKey
	}
	
	return nil
}

// RequireSuperAdmin is a middleware that requires super admin authentication
func (a *Authenticator) RequireSuperAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := getKeyFromPath(r.URL.Path, 2) // /admin/{super_key}/...
		
		if err := a.AuthenticateSuperAdmin(key); err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// RequireUser is a middleware that requires user authentication
func (a *Authenticator) RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := getKeyFromPath(r.URL.Path, 1) // /{endpoint}/{user_key}/...
		
		// The user ID is the same as the key in this implementation
		user := key
		
		if err := a.AuthenticateUser(user, key); err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		
		// Store the user in the request context
		ctx := ContextWithUser(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getKeyFromPath extracts a key from a URL path at the given index
func getKeyFromPath(path string, index int) string {
	// Skip leading slash
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	
	// Split path
	parts := splitPath(path)
	
	// Check if index is valid
	if index < len(parts) {
		return parts[index]
	}
	
	return ""
}

// splitPath splits a URL path into parts
func splitPath(path string) []string {
	var parts []string
	var part string
	
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			if part != "" {
				parts = append(parts, part)
				part = ""
			}
		} else {
			part += string(path[i])
		}
	}
	
	if part != "" {
		parts = append(parts, part)
	}
	
	return parts
}
