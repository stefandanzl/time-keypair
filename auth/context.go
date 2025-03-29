package auth

import (
	"context"
)

// contextKey is a private type for context keys
type contextKey int

const (
	// userKey is the context key for the user
	userKey contextKey = iota
)

// ContextWithUser returns a new context with the user value
func ContextWithUser(ctx context.Context, user string) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// UserFromContext returns the user from the context
func UserFromContext(ctx context.Context) (string, bool) {
	user, ok := ctx.Value(userKey).(string)
	return user, ok
}
