package grpc_client

import "time"

type ClockState struct {
	MinutesToMidnight float64
	DominantScenario  string
	ScenarioConfidence float64
	ScenarioWeights   []float64
	Sigma             float64
	ComputedAt        time.Time
}

type ScenarioSignal struct {
	Scenario   string
	Confidence float64
	SourceURL  string
	Timestamp  time.Time
}

