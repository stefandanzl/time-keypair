package auth

import (
	"data-cron-server/config"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthenticateSuperAdmin(t *testing.T) {
	cfg := config.NewConfig()
	superAdminKey := "test_super_admin_key"
	auth := NewAuthenticator(cfg, superAdminKey)
	
	// Valid key
	if err := auth.AuthenticateSuperAdmin(superAdminKey); err != nil {
		t.Errorf("AuthenticateSuperAdmin() returned error for valid key: %v", err)
	}
	
	// Invalid key
	if err := auth.AuthenticateSuperAdmin("invalid_key"); err != ErrInvalidSuperAdmin {
		t.Errorf("AuthenticateSuperAdmin() did not return expected error for invalid key: %v", err)
	}
}

func TestAuthenticateUser(t *testing.T) {
	cfg := config.NewConfig()
	user := "testuser"
	cfg.CreateUser(user)
	
	auth := NewAuthenticator(cfg, "super_admin_key")
	
	// Valid user and key
	if err := auth.AuthenticateUser(user, user); err != nil {
		t.Errorf("AuthenticateUser() returned error for valid user and key: %v", err)
	}
	
	// User not found
	if err := auth.AuthenticateUser("nonexistent", "nonexistent"); err != ErrUserNotFound {
		t.Errorf("AuthenticateUser() did not return expected error for nonexistent user: %v", err)
	}
	
	// Invalid key
	if err := auth.AuthenticateUser(user, "invalid_key"); err != ErrInvalidKey {
		t.Errorf("AuthenticateUser() did not return expected error for invalid key: %v", err)
	}
}

func TestRequireSuperAdmin(t *testing.T) {
	cfg := config.NewConfig()
	superAdminKey := "test_super_admin_key"
	auth := NewAuthenticator(cfg, superAdminKey)
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	
	middleware := auth.RequireSuperAdmin(handler)
	
	// Valid key
	req := httptest.NewRequest("GET", "/admin/"+superAdminKey+"/users", nil)
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)
	
	if rr.Code != http.StatusOK {
		t.Errorf("RequireSuperAdmin() returned status %d for valid key, expected %d", rr.Code, http.StatusOK)
	}
	
	// Invalid key
	req = httptest.NewRequest("GET", "/admin/invalid_key/users", nil)
	rr = httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)
	
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("RequireSuperAdmin() returned status %d for invalid key, expected %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestRequireUser(t *testing.T) {
	cfg := config.NewConfig()
	user := "testuser"
	cfg.CreateUser(user)
	
	auth := NewAuthenticator(cfg, "super_admin_key")
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if user is in context
		ctxUser, ok := UserFromContext(r.Context())
		if !ok || ctxUser != user {
			t.Errorf("RequireUser() did not set user in context correctly: got %s, expected %s", ctxUser, user)
		}
		w.WriteHeader(http.StatusOK)
	})
	
	middleware := auth.RequireUser(handler)
	
	// Valid user
	req := httptest.NewRequest("GET", "/data/"+user+"/keys", nil)
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)
	
	if rr.Code != http.StatusOK {
		t.Errorf("RequireUser() returned status %d for valid user, expected %d", rr.Code, http.StatusOK)
	}
	
	// Invalid user
	req = httptest.NewRequest("GET", "/data/invalid/keys", nil)
	rr = httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)
	
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("RequireUser() returned status %d for invalid user, expected %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestUserContext(t *testing.T) {
	user := "testuser"
	
	// Set user in context
	ctx := context.Background()
	ctx = ContextWithUser(ctx, user)
	
	// Get user from context
	ctxUser, ok := UserFromContext(ctx)
	if !ok {
		t.Error("UserFromContext() returned false for valid user")
	}
	if ctxUser != user {
		t.Errorf("UserFromContext() returned %s, expected %s", ctxUser, user)
	}
	
	// Try with empty context
	emptyCtx := context.Background()
	ctxUser, ok = UserFromContext(emptyCtx)
	if ok {
		t.Error("UserFromContext() returned true for empty context")
	}
	if ctxUser != "" {
		t.Errorf("UserFromContext() returned %s, expected empty string", ctxUser)
	}
}
