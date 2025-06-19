package middleware

import (
	"context"
	"net/http"
        "github.com/wdrg22/blog-aggregator/internal/database"
)

// Key used to store and retrieve user from request context
// Single source of truth for all other packages
type contextKey string
const UserContextKey = contextKey("user")

// Checks for a session cookie. If present and valid, adds that user to the request context. Never blocks a request
func PopulateUserContext(next http.Handler, db *database.Queries) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_user")
		if err != nil {
			// No cookie found. Proceed to next handler without a user in the context
			next.ServeHTTP(w, r)
			return
		}

		userName := cookie.Value
		user, err := db.GetUser(r.Context(), userName)
		if err != nil {
			// Cookie found but user is invalid or not in DB
			// Proceed to next handler without user in context
			next.ServeHTTP(w, r)
			return
		}

		// Valid user found. Add them to request context
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		// Proceed to next handler with enriched context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}


// Protects routes. Checks if a user had been added to context (by PopulateUserContext)
// If no user, redirects to login page
func RequireAuth(next http.HandlerFunc) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if user in context  
		user := r.Context().Value(UserContextKey)

		// No user in context, redirect to login
		if user == nil {
			http.Redirect(w,r,"/login", http.StatusFound)
			return
		}

		// User authenticated, call protected handler
		next(w,r)
	}
}
