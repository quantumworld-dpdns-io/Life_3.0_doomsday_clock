package graphql

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/life3/api-gateway/internal/grpc_client"
)

type Service interface {
	ClockState(context.Context) (grpc_client.ClockState, error)
	RecentSignals(context.Context, int) ([]grpc_client.ScenarioSignal, error)
}

type Handler struct {
	service Service
	schema  graphql.Schema
}

func NewHandler(service Service) *Handler {
	h := &Handler{service: service}
	h.schema = h.mustBuildSchema()
	return h
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

	result := graphql.Do(graphql.Params{
		Schema:         h.schema,
		RequestString:  req.Query,
		VariableValues: req.Variables,
		OperationName:  req.OperationName,
		Context:        r.Context(),
	})
	_ = json.NewEncoder(w).Encode(result)
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
		return map[string]any{"errors": []map[string]string{{"message": err.Error()}}}
	}
	return map[string]any{"data": map[string]any{"clockStateStream": marshalClockState(state)}}
}

func (h *Handler) mustBuildSchema() graphql.Schema {
	scenarioEnum := graphql.NewEnum(graphql.EnumConfig{
		Name: "Scenario",
		Values: graphql.EnumValueConfigMap{
			"UNKNOWN":              &graphql.EnumValueConfig{Value: "UNKNOWN"},
			"LIBERTARIAN_UTOPIA":   &graphql.EnumValueConfig{Value: "LIBERTARIAN_UTOPIA"},
			"BENEVOLENT_DICTATOR":  &graphql.EnumValueConfig{Value: "BENEVOLENT_DICTATOR"},
			"EGALITARIAN_UTOPIA":   &graphql.EnumValueConfig{Value: "EGALITARIAN_UTOPIA"},
			"GATEKEEPER":           &graphql.EnumValueConfig{Value: "GATEKEEPER"},
			"PROTECTOR_GOD":        &graphql.EnumValueConfig{Value: "PROTECTOR_GOD"},
			"ENSLAVED_GOD":         &graphql.EnumValueConfig{Value: "ENSLAVED_GOD"},
			"CONQUERORS":           &graphql.EnumValueConfig{Value: "CONQUERORS"},
			"DESCENDANTS":          &graphql.EnumValueConfig{Value: "DESCENDANTS"},
			"ZOOKEEPER":            &graphql.EnumValueConfig{Value: "ZOOKEEPER"},
			"NINETEEN_EIGHTY_FOUR": &graphql.EnumValueConfig{Value: "NINETEEN_EIGHTY_FOUR"},
			"REVERT":               &graphql.EnumValueConfig{Value: "REVERT"},
			"SELF_DESTRUCTION":     &graphql.EnumValueConfig{Value: "SELF_DESTRUCTION"},
		},
	})

	clockStateType := graphql.NewObject(graphql.ObjectConfig{
		Name: "ClockState",
		Fields: graphql.Fields{
			"minutesToMidnight":  &graphql.Field{Type: graphql.NewNonNull(graphql.Float)},
			"dominantScenario":   &graphql.Field{Type: graphql.NewNonNull(scenarioEnum)},
			"scenarioConfidence": &graphql.Field{Type: graphql.Float},
			"scenarioWeights":    &graphql.Field{Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(graphql.Float)))},
			"sigma":              &graphql.Field{Type: graphql.Float},
			"computedAt":         &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		},
	})

	signalType := graphql.NewObject(graphql.ObjectConfig{
		Name: "ScenarioSignal",
		Fields: graphql.Fields{
			"scenario":   &graphql.Field{Type: graphql.NewNonNull(scenarioEnum)},
			"confidence": &graphql.Field{Type: graphql.NewNonNull(graphql.Float)},
			"sourceUrl":  &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
			"timestamp":  &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		},
	})

	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"clockState": &graphql.Field{
				Type: graphql.NewNonNull(clockStateType),
				Resolve: func(p graphql.ResolveParams) (any, error) {
					state, err := h.service.ClockState(p.Context)
					if err != nil {
						return nil, err
					}
					return marshalClockState(state), nil
				},
			},
			"recentSignals": &graphql.Field{
				Type: graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(signalType))),
				Args: graphql.FieldConfigArgument{
					"limit": &graphql.ArgumentConfig{Type: graphql.Int, DefaultValue: 20},
				},
				Resolve: func(p graphql.ResolveParams) (any, error) {
					limit := 20
					if value, ok := p.Args["limit"].(int); ok {
						limit = value
					}
					signals, err := h.service.RecentSignals(p.Context, limit)
					if err != nil {
						return nil, err
					}
					out := make([]map[string]any, 0, len(signals))
					for _, signal := range signals {
						out = append(out, marshalSignal(signal))
					}
					return out, nil
				},
			},
		},
	})

	schema, err := graphql.NewSchema(graphql.SchemaConfig{Query: queryType})
	if err != nil {
		panic(err)
	}
	return schema
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
