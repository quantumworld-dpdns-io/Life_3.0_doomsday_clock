package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Config struct {
	APIKey    string
	JWTSecret string
	Issuer    string
	TTL       time.Duration
}

type Authenticator struct {
	apiKey    string
	jwtSecret []byte
	issuer    string
	ttl       time.Duration
}

func New(cfg Config) *Authenticator {
	ttl := cfg.TTL
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	issuer := cfg.Issuer
	if issuer == "" {
		issuer = "life3-api-gateway"
	}
	return &Authenticator{
		apiKey:    cfg.APIKey,
		jwtSecret: []byte(cfg.JWTSecret),
		issuer:    issuer,
		ttl:       ttl,
	}
}

type loginRequest struct {
	APIKey string `json:"apiKey"`
}

type loginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func (a *Authenticator) LoginHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req loginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		if req.APIKey == "" || req.APIKey != a.apiKey {
			http.Error(w, "invalid API key", http.StatusUnauthorized)
			return
		}

		token, expiresAt, err := a.IssueToken("api-client")
		if err != nil {
			http.Error(w, "failed to issue token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(loginResponse{Token: token, ExpiresAt: expiresAt})
	})
}

func (a *Authenticator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := a.ValidateRequest(r); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *Authenticator) IssueToken(subject string) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(a.ttl)
	claims := jwt.RegisteredClaims{
		Issuer:    a.issuer,
		Subject:   subject,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(a.jwtSecret)
	return signed, expiresAt, err
}

func (a *Authenticator) ValidateRequest(r *http.Request) error {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return errors.New("missing bearer token")
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(authHeader, prefix) {
		return errors.New("invalid authorization scheme")
	}
	return a.ValidateToken(strings.TrimSpace(strings.TrimPrefix(authHeader, prefix)))
}

func (a *Authenticator) ValidateToken(raw string) error {
	if raw == "" {
		return errors.New("missing bearer token")
	}

	token, err := jwt.ParseWithClaims(raw, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return a.jwtSecret, nil
	}, jwt.WithIssuer(a.issuer))
	if err != nil {
		return err
	}
	if !token.Valid {
		return errors.New("invalid bearer token")
	}
	return nil
}

