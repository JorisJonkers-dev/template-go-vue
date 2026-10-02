// Package httpx holds the HTTP pieces every inbound adapter shares: RFC 9457 problems, browser
// hardening headers, and the local stand-in for the platform's forward-auth.
package httpx

import (
	"encoding/json"
	"net/http"
)

// IdentityHeader carries the caller's subject. The platform's forward-auth sets it at the edge.
const IdentityHeader = "X-User-Id"

const contentSecurityPolicy = "default-src 'self'; img-src 'self' data:; style-src 'self'; " +
	"connect-src 'self'; object-src 'none'; base-uri 'self'; frame-ancestors 'none'"

type problem struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail,omitempty"`
}

// WriteProblem writes an RFC 9457 problem. The detail stays generic; causes belong in the logs.
func WriteProblem(w http.ResponseWriter, status int, title, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(problem{Type: "about:blank", Title: title, Status: status, Detail: detail})
}

// SecurityHeaders sets the browser hardening headers on every response.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", contentSecurityPolicy)
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

// DevIdentity fills in subject when a request carries no identity. Wire it only for local runs,
// where no forward-auth sits in front of the service.
func DevIdentity(subject string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(IdentityHeader) == "" {
			r.Header.Set(IdentityHeader, subject)
		}
		next.ServeHTTP(w, r)
	})
}
