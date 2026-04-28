use reqwest::Client;
use serde::Deserialize;

#[derive(Deserialize)]
pub struct AggregateResponse {
    pub weights: Vec<f32>,
    pub window_days: i32,
}

#[derive(Deserialize)]
pub struct EntropyResponse {
    pub bits: Vec<bool>,
    pub delta: f64,
    pub num_bits: u32,
}

pub struct IntelligenceClient {
    client: Client,
    base_url: String,
}

impl IntelligenceClient {
    pub fn new(base_url: String) -> Self {
        Self { client: Client::new(), base_url }
    }

    pub async fn get_aggregate(&self) -> anyhow::Result<Vec<f32>> {
        let url = format!("{}/signals/aggregate?window_days=7", self.base_url);
        match self.client.get(&url).send().await {
            Ok(resp) => {
                let data: AggregateResponse = resp.json().await?;
                Ok(data.weights)
            }
            Err(e) => {
                tracing::warn!("Intelligence service unavailable ({}); using zero weights.", e);
                Ok(vec![0.0f32; 13])
            }
        }
    }
}

pub struct QuantumClient {
    client: Client,
    base_url: String,
}

impl QuantumClient {
    pub fn new(base_url: String) -> Self {
        Self { client: Client::new(), base_url }
    }

    pub async fn get_entropy_delta(&self) -> anyhow::Result<f64> {
        let url = format!("{}/entropy/16", self.base_url);
        match self.client.get(&url).send().await {
            Ok(resp) => {
                let data: EntropyResponse = resp.json().await?;
                Ok(data.delta)
            }
            Err(e) => {
                tracing::warn!("Quantum-sim unavailable ({}); falling back to pseudo-random.", e);
                use rand::Rng;
                Ok(rand::thread_rng().gen_range(-1.0_f64..1.0_f64))
            }
        }
    }
}
