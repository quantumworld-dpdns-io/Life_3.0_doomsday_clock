package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/life3/api-gateway/internal/auth"
	gatewaygraphql "github.com/life3/api-gateway/internal/graphql"
	"github.com/life3/api-gateway/internal/grpc_client"
)

func main() {
	cfg := loadConfig()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	clients, err := grpc_client.New(ctx, grpc_client.Config{
		RiskEngineAddr:        cfg.RiskEngineAddr,
		IntelligenceServerAddr: cfg.IntelligenceAddr,
		Timeout:               3 * time.Second,
	})
	if err != nil {
		log.Fatalf("initialize gRPC clients: %v", err)
	}
	defer clients.Close()

	authenticator := auth.New(auth.Config{
		APIKey:    cfg.APIKey,
		JWTSecret: cfg.JWTSecret,
		Issuer:    "life3-api-gateway",
		TTL:       24 * time.Hour,
	})

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.Handle("POST /auth/login", authenticator.LoginHandler())
	mux.Handle("/graphql", authenticator.Middleware(gatewaygraphql.NewHandler(clients)))

	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("api-gateway listening on %s", cfg.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("api-gateway failed: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("api-gateway shutdown failed: %v", err)
	}
}

type config struct {
	Addr             string
	APIKey           string
	JWTSecret        string
	RiskEngineAddr   string
	IntelligenceAddr string
}

func loadConfig() config {
	return config{
		Addr:             env("API_GATEWAY_ADDR", ":4000"),
		APIKey:           env("API_GATEWAY_API_KEY", env("GATEWAY_API_KEY", "dev-api-key")),
		JWTSecret:        env("API_GATEWAY_JWT_SECRET", env("JWT_SECRET", "dev-secret-change-me")),
		RiskEngineAddr:   env("RISK_ENGINE_GRPC_ADDR", "localhost:50051"),
		IntelligenceAddr: env("INTELLIGENCE_GRPC_ADDR", "localhost:50052"),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok","service":"api-gateway"}`))
}

