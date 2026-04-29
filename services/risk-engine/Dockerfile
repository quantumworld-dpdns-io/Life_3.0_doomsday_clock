FROM rust:1.78-slim AS builder
WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    pkg-config libssl-dev protobuf-compiler \
    && rm -rf /var/lib/apt/lists/*

COPY Cargo.toml build.rs ./
COPY config/ ./config/
COPY src/ ./src/

RUN cargo build --release

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
    libssl3 ca-certificates \
    && rm -rf /var/lib/apt/lists/*

RUN adduser --disabled-password --gecos '' --uid 1001 appuser
WORKDIR /app
COPY --from=builder /app/target/release/risk-engine .
COPY config/ ./config/
RUN chown -R appuser /app
USER appuser

EXPOSE 8003 50053
CMD ["./risk-engine"]
