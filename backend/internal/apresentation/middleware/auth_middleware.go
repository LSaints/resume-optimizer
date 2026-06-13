package middleware

import (
	"context"
	"net/http"
	"strings"

	"backend/internal/application/services"
)

type contextKey string

const UserIDKey contextKey = "userID"

type AuthMiddleware struct {
	AuthServices *services.AuthServices
}

func NewAuthMiddleware(
	authServices *services.AuthServices,
) *AuthMiddleware {
	return &AuthMiddleware{
		AuthServices: authServices,
	}
}

func (m *AuthMiddleware) Middleware(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			http.Error(
				w,
				"token ausente",
				http.StatusUnauthorized,
			)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(
				w,
				"formato de token inválido",
				http.StatusUnauthorized,
			)
			return
		}

		tokenString := parts[1]

		if tokenString == "" {
			http.Error(
				w,
				"token ausente",
				http.StatusUnauthorized,
			)
			return
		}

		claims, err := m.AuthServices.ValidateToken(tokenString)
		if err != nil {
			http.Error(
				w,
				err.Error(),
				http.StatusUnauthorized,
			)
			return
		}

		userID, ok := claims["userID"].(string)
		if !ok {
			http.Error(
				w,
				"token inválido",
				http.StatusUnauthorized,
			)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
