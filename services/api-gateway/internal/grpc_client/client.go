package grpc_client

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Config struct {
	RiskEngineAddr         string
	IntelligenceServerAddr string
	Timeout                time.Duration
}

type Client struct {
	riskHTTP         string
	intelligenceHTTP string
	httpClient       *http.Client
}

func New(_ context.Context, cfg Config) (*Client, error) {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	return &Client{
		riskHTTP:         normalizeHTTPAddr(cfg.RiskEngineAddr, "http://localhost:8003"),
		intelligenceHTTP: normalizeHTTPAddr(cfg.IntelligenceServerAddr, "http://localhost:8001"),
		httpClient:       &http.Client{Timeout: timeout},
	}, nil
}

func (c *Client) Close() error {
	return nil
}

func (c *Client) ClockState(ctx context.Context) (ClockState, error) {
	if c == nil || c.riskHTTP == "" {
		return fallbackClockState(), nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.riskHTTP+"/clock", nil)
	if err != nil {
		return fallbackClockState(), nil
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("risk-engine clock fetch failed, using fallback clock state: %v", err)
		return fallbackClockState(), nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fallbackClockState(), nil
	}

	var raw struct {
		MinutesToMidnight  float64   `json:"minutes_to_midnight"`
		DominantScenario   string    `json:"dominant_scenario"`
		ScenarioConfidence float64   `json:"scenario_confidence"`
		ScenarioWeights    []float64 `json:"scenario_weights"`
		Sigma              float64   `json:"sigma"`
		ComputedAt         int64     `json:"computed_at"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return fallbackClockState(), nil
	}
	return ClockState{
		MinutesToMidnight:  raw.MinutesToMidnight,
		DominantScenario:   coalesceScenario(raw.DominantScenario),
		ScenarioConfidence: raw.ScenarioConfidence,
		ScenarioWeights:    normalizeWeights(raw.ScenarioWeights),
		Sigma:              raw.Sigma,
		ComputedAt:         time.Unix(raw.ComputedAt, 0).UTC(),
	}, nil
}

func (c *Client) RecentSignals(ctx context.Context, limit int) ([]ScenarioSignal, error) {
	limit = clampLimit(limit)
	if c == nil || c.intelligenceHTTP == "" {
		return []ScenarioSignal{}, nil
	}

	endpoint := c.intelligenceHTTP + "/signals/latest?limit=" + url.QueryEscape(strconv.Itoa(limit))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return []ScenarioSignal{}, nil
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("intelligence signals fetch failed, using empty signals: %v", err)
		return []ScenarioSignal{}, nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return []ScenarioSignal{}, nil
	}

	var raw []struct {
		Scenario   string  `json:"scenario"`
		Confidence float64 `json:"confidence"`
		SourceURL  string  `json:"source_url"`
		URL        string  `json:"url"`
		Timestamp  int64   `json:"timestamp"`
		CreatedAt  string  `json:"created_at"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return []ScenarioSignal{}, nil
	}

	signals := make([]ScenarioSignal, 0, len(raw))
	for _, item := range raw {
		timestamp := time.Unix(item.Timestamp, 0).UTC()
		if item.Timestamp == 0 && item.CreatedAt != "" {
			if parsed, err := time.Parse(time.RFC3339, item.CreatedAt); err == nil {
				timestamp = parsed.UTC()
			}
		}
		sourceURL := item.SourceURL
		if sourceURL == "" {
			sourceURL = item.URL
		}
		signals = append(signals, ScenarioSignal{
			Scenario:   coalesceScenario(item.Scenario),
			Confidence: item.Confidence,
			SourceURL:  sourceURL,
			Timestamp:  timestamp,
		})
	}
	return signals, nil
}

func (c *Client) StreamClockStates(ctx context.Context) (<-chan ClockState, <-chan error) {
	states := make(chan ClockState, 1)
	errs := make(chan error, 1)
	go func() {
		defer close(states)
		defer close(errs)
		state, err := c.ClockState(ctx)
		if err != nil {
			errs <- err
			return
		}
		select {
		case states <- state:
		case <-ctx.Done():
			errs <- ctx.Err()
		}
	}()
	return states, errs
}

func normalizeHTTPAddr(addr string, fallback string) string {
	if addr == "" {
		return fallback
	}
	if parsed, err := url.Parse(addr); err == nil && parsed.Scheme != "" {
		return addr
	}
	return "http://" + addr
}

func normalizeWeights(weights []float64) []float64 {
	out := make([]float64, 0, 12)
	start := 0
	if len(weights) == 13 {
		start = 1
	}
	for i := start; i < len(weights) && len(out) < 12; i++ {
		out = append(out, weights[i])
	}
	for len(out) < 12 {
		out = append(out, 0)
	}
	return out
}

func fallbackClockState() ClockState {
	return ClockState{
		MinutesToMidnight: 60,
		DominantScenario:  "UNKNOWN",
		ScenarioWeights:   make([]float64, 12),
		ComputedAt:        time.Now().UTC(),
	}
}

func coalesceScenario(value string) string {
	if value == "" || value == "0" {
		return "UNKNOWN"
	}
	return value
}

func clampLimit(limit int) int {
	if limit <= 0 {
		return 20
	}
	if limit > 100 {
		return 100
	}
	return limit
}
