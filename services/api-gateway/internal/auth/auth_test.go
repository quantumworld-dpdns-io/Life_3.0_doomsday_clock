package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestLoginIssuesUsableToken(t *testing.T) {
	a := New(Config{
		APIKey:    "test-key",
		JWTSecret: "test-secret",
		Issuer:    "test-issuer",
		TTL:       time.Hour,
	})

	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{"apiKey":"test-key"}`))
	rec := httptest.NewRecorder()

	a.LoginHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"token"`) {
		t.Fatalf("expected token response, got %s", rec.Body.String())
	}
}

func TestMiddlewareRejectsMissingToken(t *testing.T) {
	a := New(Config{APIKey: "test-key", JWTSecret: "test-secret"})
	req := httptest.NewRequest(http.MethodPost, "/graphql", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()

	a.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestMiddlewareAcceptsIssuedToken(t *testing.T) {
	a := New(Config{APIKey: "test-key", JWTSecret: "test-secret", TTL: time.Hour})
	token, _, err := a.IssueToken("subject")
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/graphql", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	a.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d: %s", rec.Code, rec.Body.String())
	}
}
