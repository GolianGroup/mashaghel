package app

import (
	"context"
	"log"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/credentials"
)

func (a *application) InitTracerProvider() oteltrace.Tracer {
	return otel.Tracer(a.config.Tracer.ServiceName)
}

func (a *application) InitTracer() func(context.Context) error {
	var secureOption otlptracegrpc.Option

	if strings.ToLower(a.config.Tracer.Insecure) == "false" {
		secureOption = otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	} else {
		secureOption = otlptracegrpc.WithInsecure()
	}

	exporter, err := otlptrace.New(
		a.ctx,
		otlptracegrpc.NewClient(
			secureOption,
			otlptracegrpc.WithEndpoint(a.config.Tracer.CollectorUrl),
		),
	)
	if err != nil {
		log.Fatalf("Failed to create exporter: %v", err)
	}

	resources, err := resource.New(
		a.ctx,
		resource.WithAttributes(
			attribute.String("service.name", a.config.Tracer.ServiceName),
			attribute.String("library.language", "go"),
		),
	)
	if err != nil {
		log.Fatalf("Could not set resources: %v", err)
	}

	otel.SetTracerProvider(
		trace.NewTracerProvider(
			trace.WithSampler(trace.AlwaysSample()),
			trace.WithBatcher(exporter),
			trace.WithResource(resources),
		))

	return exporter.Shutdown
}
