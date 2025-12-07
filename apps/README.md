# Local Apps

Simple Go and Node.js demo services with Prometheus metrics and OTLP tracing via the shared OpenTelemetry Collector.

## Run everything
From this folder:
```sh
podman compose -f compose.yaml -f nodejs-app/compose.yaml -f go-app/compose.yaml up -d --build
```

## Stop everything
```sh
podman compose -f compose.yaml -f nodejs-app/compose.yaml -f go-app/compose.yaml down
```

## Endpoints
- Node: http://localhost:3001/ (root), /health, /metrics
- Go: http://localhost:3002/ (root), /health, /metrics
- Collector metrics: http://localhost:9464/metrics
