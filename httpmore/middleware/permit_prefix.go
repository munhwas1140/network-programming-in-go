package middleware

import (
	"net/http"
)

func PermitPrefix(prefixes []string, next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if check(prefixes, r.URL.Path) {
				next.ServeHTTP(w, r)
			} else {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
			}
		},
	)
}

func check(prefixes []string, path string) bool {
	return false
}
