# Go App

Lightweight Go HTTP service with Prometheus metrics, basic logging, and OTLP tracing.

## Run (with shared collector)
From `apps/`:
```sh
podman compose -f compose.yaml -f go-app/compose.yaml up -d --build
```

## Endpoints
- `http://localhost:3002/` – hello text
- `/health` – health check JSON
- `/metrics` – Prometheus metrics
