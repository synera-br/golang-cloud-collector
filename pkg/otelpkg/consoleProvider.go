package otelpkg

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
)

func newConsoleExporter() (*OtelPkgInstrument, error) {

	_, err := stdouttrace.New()
	if err != nil {
		return nil, err
	}

	metric, err := stdoutmetric.New()
	if err != nil {
		return nil, err
	}

	l, err := stdoutlog.New()
	if err != nil {
		return nil, err
	}

	return &OtelPkgInstrument{
		Metric: metric,
		Log:    l,
	}, nil
}

func newOtlpExporter(ctx context.Context) (*OtelPkgInstrument, error) {

	_, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, err
	}

	metric, err := otlpmetrichttp.New(ctx)
	if err != nil {
		return nil, err
	}

	l, err := otlploghttp.New(ctx)
	if err != nil {
		return nil, err
	}

	return &OtelPkgInstrument{
		Metric: metric,
		Log:    l,
	}, nil
}
