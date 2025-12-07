# Node.js App

Minimal Express service with Prometheus metrics, Pino logging, and OTLP tracing.

## Run (with shared collector)
From `apps/`:
```sh
podman compose -f compose.yaml -f nodejs-app/compose.yaml up -d --build
```

## Endpoints
- `http://localhost:3001/` – hello text
- `/health` – health check JSON
- `/metrics` – Prometheus metrics
