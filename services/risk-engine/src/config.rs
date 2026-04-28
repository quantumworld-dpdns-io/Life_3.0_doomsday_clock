use serde::Deserialize;
use std::collections::HashMap;

#[derive(Debug, Deserialize, Clone)]
pub struct ScenarioWeightsConfig {
    pub scenarios: HashMap<String, f32>,
    pub clock: ClockConfig,
}

#[derive(Debug, Deserialize, Clone)]
pub struct ClockConfig {
    pub base_minutes_to_midnight: f32,
    pub monte_carlo_iterations: u32,
    pub entropy_bits_per_step: u32,
    pub floor_minutes: f32,
}

#[derive(Debug, Clone)]
pub struct AppConfig {
    pub weights_cfg: ScenarioWeightsConfig,
    pub intelligence_addr: String,
    pub quantum_addr: String,
    pub grpc_port: u16,
    pub http_port: u16,
}

impl AppConfig {
    pub fn load() -> anyhow::Result<Self> {
        let weights_cfg: ScenarioWeightsConfig = {
            let content = std::fs::read_to_string("config/scenario_weights.toml")?;
            toml::from_str(&content)?
        };
        Ok(Self {
            weights_cfg,
            intelligence_addr: std::env::var("INTELLIGENCE_ADDR")
                .unwrap_or_else(|_| "http://localhost:8001".to_string()),
            quantum_addr: std::env::var("QUANTUM_ADDR")
                .unwrap_or_else(|_| "http://localhost:8002".to_string()),
            grpc_port: std::env::var("GRPC_PORT")
                .ok()
                .and_then(|s| s.parse().ok())
                .unwrap_or(50053),
            http_port: std::env::var("HTTP_PORT")
                .ok()
                .and_then(|s| s.parse().ok())
                .unwrap_or(8003),
        })
    }

    #[cfg(test)]
    pub fn for_test() -> anyhow::Result<Self> {
        let content = include_str!("../config/scenario_weights.toml");
        let weights_cfg = toml::from_str(content)?;
        Ok(Self {
            weights_cfg,
            intelligence_addr: "http://localhost:8001".to_string(),
            quantum_addr: "http://localhost:8002".to_string(),
            grpc_port: 50053,
            http_port: 8003,
        })
    }
}
