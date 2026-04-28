package graphql

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/life3/api-gateway/internal/grpc_client"
)

type fakeService struct{}

func (fakeService) ClockState(context.Context) (grpc_client.ClockState, error) {
	return grpc_client.ClockState{
		MinutesToMidnight:  1.5,
		DominantScenario:   "SELF_DESTRUCTION",
		ScenarioConfidence: 0.92,
		ScenarioWeights:    []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 0.1, 0.2, 0.3},
		Sigma:              0.04,
		ComputedAt:         time.Unix(1710000000, 0).UTC(),
	}, nil
}

func (fakeService) RecentSignals(context.Context, int) ([]grpc_client.ScenarioSignal, error) {
	return []grpc_client.ScenarioSignal{
		{
			Scenario:   "GATEKEEPER",
			Confidence: 0.7,
			SourceURL:  "https://example.test/signal",
			Timestamp:  time.Unix(1710000001, 0).UTC(),
		},
	}, nil
}

func TestGraphQLClockStateAndSignals(t *testing.T) {
	body := `{"query":"query { clockState { minutesToMidnight dominantScenario computedAt scenarioWeights } recentSignals(limit: 1) { scenario confidence sourceUrl timestamp } }"}`
	req := httptest.NewRequest(http.MethodPost, "/graphql", strings.NewReader(body))
	rec := httptest.NewRecorder()

	NewHandler(fakeService{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	response := rec.Body.String()
	for _, want := range []string{"SELF_DESTRUCTION", "GATEKEEPER", "https://example.test/signal"} {
		if !strings.Contains(response, want) {
			t.Fatalf("expected response to contain %q, got %s", want, response)
		}
	}
}

func TestSubscriptionPlaceholder(t *testing.T) {
	body := `{"query":"subscription { clockStateStream { minutesToMidnight dominantScenario } }"}`
	req := httptest.NewRequest(http.MethodPost, "/graphql", strings.NewReader(body))
	rec := httptest.NewRecorder()

	NewHandler(fakeService{}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "clockStateStream") {
		t.Fatalf("expected subscription placeholder payload, got %s", rec.Body.String())
	}
}

