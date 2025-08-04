package middleware

import (
	"net/http"

	"github.com/fhsmendes/rate-limiter/configs"
	"github.com/go-chi/jwtauth"
)

func ValidTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("API_KEY")
		if token == "" {
			http.Error(w, "API_KEY header required", http.StatusUnauthorized)
			return
		}

		tokenAuth := jwtauth.New("HS256", []byte(configs.Envs.Jwt_Secret), nil)

		// Parse and verify the token
		jwtToken, err := tokenAuth.Decode(token)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Add token to context
		ctx := jwtauth.NewContext(r.Context(), jwtToken, nil)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
