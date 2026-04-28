FROM golang:1.22-alpine AS builder

WORKDIR /src
COPY services/api-gateway/go.mod services/api-gateway/go.sum ./
RUN go mod download

COPY services/api-gateway/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api-gateway ./cmd/gateway

FROM alpine:3.20
RUN adduser -D -H api
USER api
COPY --from=builder /out/api-gateway /usr/local/bin/api-gateway
EXPOSE 4000
ENTRYPOINT ["api-gateway"]
