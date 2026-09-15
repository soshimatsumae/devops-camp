package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/smatsumae/devops-camp/internal/httpx"
)

type contextKey string

const userIDKey contextKey = "userID"

// RequireAuth returns a decorator that validates the Authorization: Bearer <token>
// header and injects the authenticated user id into the request context.
func RequireAuth(secret []byte) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			token, ok := strings.CutPrefix(header, "Bearer ")
			if !ok || token == "" {
				httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing or malformed Authorization header")
				return
			}

			userID, err := ParseToken(secret, token)
			if err != nil {
				httpx.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next(w, r.WithContext(ctx))
		}
	}
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDKey).(int64)
	return userID, ok
}

// ContextWithUserID returns a copy of ctx carrying userID as the authenticated
// user, as RequireAuth would after validating a token. Exported for tests that
// call handlers directly without going through the RequireAuth middleware.
func ContextWithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}
