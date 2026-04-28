use std::time::Duration;
use tokio::time::interval;

use crate::clients::{IntelligenceClient, QuantumClient};
use crate::clock::ClockState;
use crate::config::AppConfig;
use crate::monte_carlo;

/// Compute a fresh ClockState by pulling from upstream services.
pub async fn compute_state(
    intelligence: &IntelligenceClient,
    quantum: &QuantumClient,
    cfg: &AppConfig,
) -> anyhow::Result<ClockState> {
    let weights = intelligence.get_aggregate().await?;
    let entropy_delta = quantum.get_entropy_delta().await?;
    Ok(monte_carlo::simulate(&weights, entropy_delta, cfg))
}

/// Main gRPC/compute loop.
/// Tonic `RiskEngine` service will be wired here once protos are compiled.
/// For now, logs computed state on a 60-second tick.
pub async fn serve(cfg: AppConfig) -> anyhow::Result<()> {
    let grpc_addr: std::net::SocketAddr = format!("0.0.0.0:{}", cfg.grpc_port).parse()?;
    tracing::info!("Risk Engine compute loop starting (gRPC will bind {})", grpc_addr);

    let intelligence = IntelligenceClient::new(cfg.intelligence_addr.clone());
    let quantum = QuantumClient::new(cfg.quantum_addr.clone());

    let mut tick = interval(Duration::from_secs(60));
    loop {
        tick.tick().await;
        match compute_state(&intelligence, &quantum, &cfg).await {
            Ok(state) => tracing::info!(
                minutes_to_midnight = state.minutes_to_midnight,
                sigma = state.sigma,
                dominant_scenario = state.dominant_scenario,
                "ClockState computed"
            ),
            Err(e) => tracing::error!("Clock computation failed: {}", e),
        }
    }
}
