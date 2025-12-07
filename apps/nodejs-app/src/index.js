import express from 'express';
import promClient from 'prom-client';
import { context, trace } from '@opentelemetry/api';
import { NodeTracerProvider } from '@opentelemetry/sdk-trace-node';
import { SimpleSpanProcessor } from '@opentelemetry/sdk-trace-base';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-http';
import pino from 'pino';

const app = express();
const port = process.env.PORT || 3000;
const logger = pino({ level: process.env.LOG_LEVEL || 'info' });

// --- Tracing setup (no-op if exporter unreachable)
const tracerProvider = new NodeTracerProvider();
const otlpEndpoint =
  process.env.OTEL_EXPORTER_OTLP_ENDPOINT || 'http://localhost:4318/v1/traces';
tracerProvider.addSpanProcessor(
  new SimpleSpanProcessor(
    new OTLPTraceExporter({
      url: otlpEndpoint,
      timeoutMillis: 2000,
    }),
  ),
);
tracerProvider.register();
const tracer = trace.getTracer('nodejs-app');

// --- Metrics setup
promClient.collectDefaultMetrics({ prefix: 'nodejs_app_' });
const httpRequests = new promClient.Counter({
  name: 'nodejs_app_http_requests_total',
  help: 'Total HTTP requests',
  labelNames: ['method', 'path', 'status'],
});

// simple middleware to wrap request in a span and record metrics
app.use((req, res, next) => {
  const span = tracer.startSpan(`${req.method} ${req.path}`);
  res.on('finish', () => {
    httpRequests
      .labels(req.method, req.path, String(res.statusCode))
      .inc();
    span.setAttribute('http.status_code', res.statusCode);
    logger.info(
      { path: req.path, status: res.statusCode, method: req.method },
      'request handled',
    );
    span.end();
  });
  context.with(trace.setSpan(context.active(), span), next);
});

app.get('/', (_req, res) => {
  res.send('hey i am nodejs app');
});

app.get('/health', (_req, res) => {
  res.status(200).json({ status: 'ok' });
});

app.get('/metrics', async (_req, res) => {
  res.set('Content-Type', promClient.register.contentType);
  res.end(await promClient.register.metrics());
});

app.listen(port, () => {
  // eslint-disable-next-line no-console
  console.log(`nodejs app listening on port ${port}`);
});
