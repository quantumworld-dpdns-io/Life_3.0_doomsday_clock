use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ClockState {
    pub minutes_to_midnight: f32,
    pub sigma: f32,
    /// 1-12 per Tegmark's 12 scenarios; 0 = unknown
    pub dominant_scenario: u8,
    pub scenario_confidence: f32,
    /// length 13, index = scenario ID (index 0 unused)
    pub scenario_weights: Vec<f32>,
    pub computed_at: i64,
}

impl ClockState {
    pub fn unknown() -> Self {
        Self {
            minutes_to_midnight: 45.0,
            sigma: 0.0,
            dominant_scenario: 0,
            scenario_confidence: 0.0,
            scenario_weights: vec![0.0; 13],
            computed_at: now_unix(),
        }
    }
}

pub fn now_unix() -> i64 {
    std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap_or_default()
        .as_secs() as i64
}
