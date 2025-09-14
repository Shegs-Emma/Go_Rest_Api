package middlewares

import (
	"fmt"
	"net/http"
)

// Allowed Origins
var allowedOrigins = []string {
	"https://my-origin-url.com",
	"https://localhost:8000",
}

func Cors (next http.Handler) http.Handler {
	fmt.Println("Cors Middleware...")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Cors Middleware being returned...")
		origin := r.Header.Get("Origin")

		fmt.Println("origin:", origin)

		if isOriginAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			http.Error(w, "Not Allowed by CORS", http.StatusForbidden)
			return
		}

		w.Header().Set("Allow-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Allow-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE")
		w.Header().Set("Allow-Control-Allow-Credentials", "true")
		w.Header().Set("Allow-Control-Expose-Headers", "Authorization")
		w.Header().Set("Allow-Control-Max-Age", "3600")

		if r.Method == http.MethodOptions {
			return
		}

		next.ServeHTTP(w, r)
		fmt.Println("Cors Middleware ends...")
	})
}

func isOriginAllowed(origin string) bool {
	for _, allowedOrigin := range allowedOrigins {
		if origin == allowedOrigin {
			return true
		}
	}
	return false
}