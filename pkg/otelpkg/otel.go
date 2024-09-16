package otelpkg

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

type OtelPkgInterface interface{}

type OtelExtraParameters struct {
	AppName string
	License string
}

type OtelPkgInstrument struct {
	Span       trace.Span
	Metric     metric.Exporter
	Log        log.Exporter
	TracerSdk  *sdktrace.TracerProvider
	Tracer     trace.Tracer
	Attr       attribute.KeyValue
	Parameters OtelExtraParameters
}

func NewOtel(ctx context.Context, pathConfigFile, nameFileConfig, nameFileExtention string) (*OtelPkgInstrument, error) {

	var err error
	otl := &OtelPkgInstrument{}

	otlCfg, err := parseConfig(nameFileConfig, nameFileConfig, nameFileConfig)
	if err != nil {
		return nil, err
	}

	if otlCfg.Provider == "otlp" {
		otl, err = newOtlpExporter(ctx)
		if err != nil {
			return nil, err
		}
	} else {
		otl, err = newConsoleExporter()
		if err != nil {
			return nil, err
		}
	}

	tp := newTraceProvider(otlCfg)
	otel.SetTracerProvider(tp)

	otl.TracerSdk = tp
	otl.Tracer = tp.Tracer("start")

	otl.Parameters.AppName = otlCfg.Name
	otl.Parameters.License = otlCfg.Headers["api-key"]

	return otl, err
}

func newTraceProvider(otlCfg *otelConfig) *sdktrace.TracerProvider {

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracehttp.NewClient(
			otlptracehttp.WithEndpoint(otlCfg.Endpoint),
			otlptracehttp.WithHeaders(otlCfg.Headers),
			// otlptracehttp.WithInsecure(),
		),
	)
	if err != nil {
		fmt.Println("Error is....", err)
	}

	tracerprovider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(
			exporter,
			sdktrace.WithMaxExportBatchSize(sdktrace.DefaultMaxExportBatchSize),
			sdktrace.WithBatchTimeout(sdktrace.DefaultScheduleDelay*time.Millisecond),
			sdktrace.WithMaxExportBatchSize(sdktrace.DefaultMaxExportBatchSize),
		),
		sdktrace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(otlCfg.Name),
			),
		),
	)

	otel.SetTracerProvider(tracerprovider)
	return tracerprovider

}
