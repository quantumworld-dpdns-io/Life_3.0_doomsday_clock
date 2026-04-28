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

    let health_port = cfg.http_port;
    tokio::spawn(async move {
        run_health_server(health_port).await;
    });

    grpc_server::serve(cfg).await?;
    Ok(())
}

async fn run_health_server(port: u16) {
    use axum::{routing::get, Router};
    let app = Router::new().route("/health", get(|| async { r#"{"status":"ok"}"# }));
    let addr: std::net::SocketAddr = format!("0.0.0.0:{port}").parse().unwrap();
    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    tracing::info!("Health server listening on {}", addr);
    axum::serve(listener, app).await.unwrap();
}
