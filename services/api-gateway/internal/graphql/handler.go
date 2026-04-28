package graphql

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/life3/api-gateway/internal/grpc_client"
)

type Service interface {
	ClockState(context.Context) (grpc_client.ClockState, error)
	RecentSignals(context.Context, int) ([]grpc_client.ScenarioSignal, error)
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

type requestBody struct {
	Query         string         `json:"query"`
	Variables     map[string]any `json:"variables"`
	OperationName string         `json:"operationName"`
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	req, err := decodeGraphQLRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if isClockSubscriptionPlaceholder(req.Query) {
		_ = json.NewEncoder(w).Encode(h.subscriptionPlaceholder(r.Context()))
		return
	}
	_ = json.NewEncoder(w).Encode(h.execute(r.Context(), req))
}

func decodeGraphQLRequest(r *http.Request) (requestBody, error) {
	if r.Method == http.MethodGet {
		return requestBody{
			Query:         r.URL.Query().Get("query"),
			OperationName: r.URL.Query().Get("operationName"),
		}, nil
	}

	var req requestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return requestBody{}, err
	}
	return req, nil
}

func isClockSubscriptionPlaceholder(query string) bool {
	normalized := strings.ToLower(query)
	return strings.Contains(normalized, "subscription") && strings.Contains(normalized, "clockstatestream")
}

func (h *Handler) subscriptionPlaceholder(ctx context.Context) map[string]any {
	state, err := h.service.ClockState(ctx)
	if err != nil {
		return graphQLError(err)
	}
	return map[string]any{"data": map[string]any{"clockStateStream": marshalClockState(state)}}
}

func (h *Handler) execute(ctx context.Context, req requestBody) map[string]any {
	data := map[string]any{}
	query := strings.ToLower(req.Query)

	if strings.Contains(query, "clockstate") {
		state, err := h.service.ClockState(ctx)
		if err != nil {
			return graphQLError(err)
		}
		data["clockState"] = marshalClockState(state)
	}
	if strings.Contains(query, "recentsignals") {
		signals, err := h.service.RecentSignals(ctx, limitFromRequest(req))
		if err != nil {
			return graphQLError(err)
		}
		out := make([]map[string]any, 0, len(signals))
		for _, signal := range signals {
			out = append(out, marshalSignal(signal))
		}
		data["recentSignals"] = out
	}
	if len(data) == 0 {
		return graphQLErrorMessage("unsupported query")
	}
	return map[string]any{"data": data}
}

func limitFromRequest(req requestBody) int {
	if raw, ok := req.Variables["limit"]; ok {
		switch value := raw.(type) {
		case float64:
			return CoerceLimit(strconv.Itoa(int(value)))
		case string:
			return CoerceLimit(value)
		}
	}

	query := strings.ToLower(req.Query)
	idx := strings.Index(query, "recentsignals")
	if idx < 0 {
		return 20
	}
	fragment := query[idx:]
	limitIdx := strings.Index(fragment, "limit")
	if limitIdx < 0 {
		return 20
	}
	fragment = fragment[limitIdx+len("limit"):]
	colon := strings.Index(fragment, ":")
	if colon < 0 {
		return 20
	}
	fragment = strings.TrimSpace(fragment[colon+1:])
	end := 0
	for end < len(fragment) && fragment[end] >= '0' && fragment[end] <= '9' {
		end++
	}
	return CoerceLimit(fragment[:end])
}

func marshalClockState(state grpc_client.ClockState) map[string]any {
	return map[string]any{
		"minutesToMidnight":  state.MinutesToMidnight,
		"dominantScenario":   state.DominantScenario,
		"scenarioConfidence": state.ScenarioConfidence,
		"scenarioWeights":    state.ScenarioWeights,
		"sigma":              state.Sigma,
		"computedAt":         formatTime(state.ComputedAt),
	}
}

func marshalSignal(signal grpc_client.ScenarioSignal) map[string]any {
	return map[string]any{
		"scenario":   signal.Scenario,
		"confidence": signal.Confidence,
		"sourceUrl":  signal.SourceURL,
		"timestamp":  formatTime(signal.Timestamp),
	}
}

func graphQLError(err error) map[string]any {
	return graphQLErrorMessage(err.Error())
}

func graphQLErrorMessage(message string) map[string]any {
	return map[string]any{"errors": []map[string]string{{"message": message}}}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return time.Unix(0, 0).UTC().Format(time.RFC3339)
	}
	return t.UTC().Format(time.RFC3339)
}

func CoerceLimit(raw string) int {
	limit, err := strconv.Atoi(raw)
	if err != nil || limit <= 0 {
		return 20
	}
	if limit > 100 {
		return 100
	}
	return limit
}
