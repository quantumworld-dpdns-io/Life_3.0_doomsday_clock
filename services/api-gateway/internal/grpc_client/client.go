package grpc_client

import (
	"context"
	"io"
	"log"
	"time"

	life3proto "github.com/life3/api-gateway/internal/proto/github.com/life3/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Config struct {
	RiskEngineAddr         string
	IntelligenceServerAddr string
	Timeout                time.Duration
}

type Client struct {
	riskConn         *grpc.ClientConn
	intelligenceConn *grpc.ClientConn
	risk             life3proto.RiskEngineClient
	intelligence     life3proto.IntelligenceServiceClient
	timeout          time.Duration
}

func New(ctx context.Context, cfg Config) (*Client, error) {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 3 * time.Second
	}

	c := &Client{timeout: timeout}

	if cfg.RiskEngineAddr != "" {
		conn, err := grpc.DialContext(ctx, cfg.RiskEngineAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}
		c.riskConn = conn
		c.risk = life3proto.NewRiskEngineClient(conn)
	}

	if cfg.IntelligenceServerAddr != "" {
		conn, err := grpc.DialContext(ctx, cfg.IntelligenceServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			_ = c.Close()
			return nil, err
		}
		c.intelligenceConn = conn
		c.intelligence = life3proto.NewIntelligenceServiceClient(conn)
	}

	return c, nil
}

func (c *Client) Close() error {
	if c == nil {
		return nil
	}
	if c.riskConn != nil {
		_ = c.riskConn.Close()
	}
	if c.intelligenceConn != nil {
		_ = c.intelligenceConn.Close()
	}
	return nil
}

func (c *Client) ClockState(ctx context.Context) (ClockState, error) {
	if c == nil || c.risk == nil {
		return fallbackClockState(), nil
	}

	callCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	state, err := c.risk.GetClockState(callCtx, &emptypb.Empty{})
	if err != nil {
		log.Printf("risk-engine GetClockState failed, using fallback clock state: %v", err)
		return fallbackClockState(), nil
	}
	return clockStateFromProto(state), nil
}

func (c *Client) RecentSignals(ctx context.Context, limit int) ([]ScenarioSignal, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if c == nil || c.intelligence == nil {
		return []ScenarioSignal{}, nil
	}

	callCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.intelligence.GetLatestSignals(callCtx, &life3proto.SignalRequest{Limit: int32(limit)})
	if err != nil {
		log.Printf("intelligence GetLatestSignals failed, using empty signals: %v", err)
		return []ScenarioSignal{}, nil
	}

	signals := make([]ScenarioSignal, 0, len(resp.GetSignals()))
	for _, signal := range resp.GetSignals() {
		signals = append(signals, signalFromProto(signal))
	}
	return signals, nil
}

func (c *Client) StreamClockStates(ctx context.Context) (<-chan ClockState, <-chan error) {
	states := make(chan ClockState)
	errs := make(chan error, 1)

	go func() {
		defer close(states)
		defer close(errs)

		if c == nil || c.risk == nil {
			state := fallbackClockState()
			select {
			case states <- state:
			case <-ctx.Done():
				errs <- ctx.Err()
			}
			return
		}

		callCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		stream, err := c.risk.StreamClockState(callCtx, &emptypb.Empty{})
		if err != nil {
			errs <- err
			return
		}
		for {
			msg, err := stream.Recv()
			if err != nil {
				if err != io.EOF && ctx.Err() == nil {
					errs <- err
				}
				return
			}
			select {
			case states <- clockStateFromProto(msg):
			case <-ctx.Done():
				errs <- ctx.Err()
				return
			}
		}
	}()

	return states, errs
}

func clockStateFromProto(state *life3proto.ClockState) ClockState {
	if state == nil {
		return fallbackClockState()
	}
	return ClockState{
		MinutesToMidnight: float64(state.GetMinutesToMidnight()),
		DominantScenario:  scenarioName(state.GetDominantScenario()),
		ScenarioConfidence: float64(state.GetScenarioConfidence()),
		ScenarioWeights:   normalizeWeights(state.GetScenarioWeights()),
		Sigma:             float64(state.GetSigma()),
		ComputedAt:        time.Unix(state.GetComputedAt(), 0).UTC(),
	}
}

func signalFromProto(signal *life3proto.ScenarioSignal) ScenarioSignal {
	if signal == nil {
		return ScenarioSignal{}
	}
	return ScenarioSignal{
		Scenario:   scenarioName(signal.GetScenario()),
		Confidence: float64(signal.GetConfidence()),
		SourceURL:  signal.GetSourceUrl(),
		Timestamp:  time.Unix(signal.GetTimestamp(), 0).UTC(),
	}
}

func normalizeWeights(weights []float32) []float64 {
	out := make([]float64, 0, 12)
	start := 0
	if len(weights) == 13 {
		start = 1
	}
	for i := start; i < len(weights) && len(out) < 12; i++ {
		out = append(out, float64(weights[i]))
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

func scenarioName(id life3proto.ScenarioID) string {
	switch id {
	case life3proto.ScenarioID_LIBERTARIAN_UTOPIA:
		return "LIBERTARIAN_UTOPIA"
	case life3proto.ScenarioID_BENEVOLENT_DICTATOR:
		return "BENEVOLENT_DICTATOR"
	case life3proto.ScenarioID_EGALITARIAN_UTOPIA:
		return "EGALITARIAN_UTOPIA"
	case life3proto.ScenarioID_GATEKEEPER:
		return "GATEKEEPER"
	case life3proto.ScenarioID_PROTECTOR_GOD:
		return "PROTECTOR_GOD"
	case life3proto.ScenarioID_ENSLAVED_GOD:
		return "ENSLAVED_GOD"
	case life3proto.ScenarioID_CONQUERORS:
		return "CONQUERORS"
	case life3proto.ScenarioID_DESCENDANTS:
		return "DESCENDANTS"
	case life3proto.ScenarioID_ZOOKEEPER:
		return "ZOOKEEPER"
	case life3proto.ScenarioID_NINETEEN_EIGHTY_FOUR:
		return "NINETEEN_EIGHTY_FOUR"
	case life3proto.ScenarioID_REVERT:
		return "REVERT"
	case life3proto.ScenarioID_SELF_DESTRUCTION:
		return "SELF_DESTRUCTION"
	default:
		return "UNKNOWN"
	}
}

