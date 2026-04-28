mod clients;
mod clock;
mod config;
mod grpc_server;
mod monte_carlo;

use tracing_subscriber::EnvFilter;

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    tracing_subscriber::fmt()
        .with_env_filter(
            EnvFilter::from_default_env()
                .add_directive("risk_engine=info".parse()?)
                .add_directive("warn".parse()?),
        )
        .init();

    let cfg = config::AppConfig::load()?;

    let http_cfg = cfg.clone();
    tokio::spawn(async move {
        run_http_server(http_cfg).await;
    });

    grpc_server::serve(cfg).await?;
    Ok(())
}

async fn run_http_server(cfg: config::AppConfig) {
    use axum::{routing::get, Json, Router};
    let port = cfg.http_port;
    let clock_cfg = cfg.clone();
    let app = Router::new()
        .route("/health", get(|| async { r#"{"status":"ok"}"# }))
        .route(
            "/clock",
            get(move || {
                let cfg = clock_cfg.clone();
                async move {
                    let intelligence = clients::IntelligenceClient::new(cfg.intelligence_addr.clone());
                    let quantum = clients::QuantumClient::new(cfg.quantum_addr.clone());
                    let state = grpc_server::compute_state(&intelligence, &quantum, &cfg)
                        .await
                        .unwrap_or_else(|_| clock::ClockState::unknown());
                    Json(state)
                }
            }),
        );
    let addr: std::net::SocketAddr = format!("0.0.0.0:{port}").parse().unwrap();
    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    tracing::info!("HTTP server listening on {}", addr);
    axum::serve(listener, app).await.unwrap();
}
