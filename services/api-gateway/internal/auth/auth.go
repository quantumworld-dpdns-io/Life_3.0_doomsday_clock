package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
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
	claims := map[string]any{
		"iss": a.issuer,
		"sub": subject,
		"iat": now.Unix(),
		"exp": expiresAt.Unix(),
	}
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", time.Time{}, err
	}
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", time.Time{}, err
	}
	signingInput := base64.RawURLEncoding.EncodeToString(headerJSON) + "." + base64.RawURLEncoding.EncodeToString(claimsJSON)
	return signingInput + "." + a.sign(signingInput), expiresAt, nil
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

	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return errors.New("invalid bearer token")
	}
	signingInput := parts[0] + "." + parts[1]
	if !hmac.Equal([]byte(parts[2]), []byte(a.sign(signingInput))) {
		return errors.New("invalid bearer token signature")
	}
	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return errors.New("invalid bearer token claims")
	}
	var claims struct {
		Issuer  string `json:"iss"`
		Expires int64  `json:"exp"`
	}
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return errors.New("invalid bearer token claims")
	}
	if claims.Issuer != a.issuer {
		return errors.New("invalid bearer token issuer")
	}
	if claims.Expires <= time.Now().UTC().Unix() {
		return errors.New("expired bearer token")
	}
	return nil
}

func (a *Authenticator) sign(signingInput string) string {
	mac := hmac.New(sha256.New, a.jwtSecret)
	_, _ = mac.Write([]byte(signingInput))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
