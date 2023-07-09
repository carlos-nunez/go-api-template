package middleware

import (
	"net/http"
	"strings"

	"github.com/carlos-nunez/go-api-template/services"
)

// for use on route (using a http.HandlerFunc)
func Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		prefix := "Bearer "
		authHeader := r.Header.Get("Authorization")
		reqToken := strings.TrimPrefix(authHeader, prefix)

		if reqToken == "" {
			errMsg := "No token present!"
			http.Error(w, errMsg, http.StatusForbidden)
			return
		}

		_, err := services.ValidateToken(reqToken)

		if err != "" {
			errMsg := "Authentication error!"
			http.Error(w, errMsg, http.StatusForbidden)
			return
		}

		next(w, r)
	}
}
