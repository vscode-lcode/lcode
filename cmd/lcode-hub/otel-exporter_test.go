package main

import (
	"context"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func TestExporter(t *testing.T) {
	tp := newTracerProvider(0b111)
	defer tp.Shutdown(context.Background())

	tracer := otel.Tracer("test")

	_, span := tracer.Start(context.Background(), "err")
	span.SetStatus(codes.Error, "some err string")
	span.End()

	_, span = tracer.Start(context.Background(), "debug")
	span.SetAttributes(
		attribute.Bool("debug", true),
		attribute.String("some debug", "some value"),
	)
	span.End()

	_, span = tracer.Start(context.Background(), "info")
	span.SetStatus(codes.Ok, "some err string")
	span.End()

	time.Sleep(1 * time.Second)
}
