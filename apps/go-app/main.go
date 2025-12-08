package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

var (
	httpReqs = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "go_app",
			Name:      "http_requests_total",
			Help:      "Total HTTP requests",
		},
		[]string{"method", "path", "status"},
	)
)

func main() {
	ctx := context.Background()
	logger := log.New(os.Stdout, "", log.LstdFlags)

	otelShutdown := setupTracing(ctx)
	defer func() {
		if otelShutdown != nil {
			_ = otelShutdown(context.Background())
		}
	}()

	prometheus.MustRegister(httpReqs)

	mux := http.NewServeMux()
	mux.HandleFunc("/", traceHandler("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "hey i am go app")
	}))
	health := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}
	mux.HandleFunc("/health", traceHandler("/health", health))
	mux.HandleFunc("/health/", traceHandler("/health", health))

	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/metrics/", promhttp.Handler())

	addr := ":3000"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}
	logger.Printf("go app listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		logger.Fatalf("server failed: %v", err)
	}
}

func traceHandler(name string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	spanCtx, span := otel.Tracer("go-app").Start(r.Context(), r.Method+" "+name)
	start := time.Now()
	h(w, r.WithContext(spanCtx))
	status := fmt.Sprintf("%d", http.StatusOK)
	httpReqs.WithLabelValues(r.Method, name, status).Inc()
	log.Printf("handled %s %s status=%s", r.Method, name, status)
	span.SetAttributes(
		semconv.HTTPMethodKey.String(r.Method),
		semconv.HTTPTargetKey.String(name),
		semconv.HTTPStatusCodeKey.Int(http.StatusOK),
		semconv.ServerAddressKey.String(r.Host),
	)
	span.SetAttributes(semconv.ExceptionMessageKey.String(fmt.Sprintf("duration_ms:%d", time.Since(start).Milliseconds())))
	span.End()
	}
}

func setupTracing(ctx context.Context) func(context.Context) error {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:4318/v1/traces"
	}
	exp, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpoint(endpoint), otlptracehttp.WithInsecure())
	if err != nil {
		log.Printf("tracing disabled (exporter init failed): %v", err)
		return nil
	}
	res, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("go-app"),
		),
	)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	return tp.Shutdown
}
