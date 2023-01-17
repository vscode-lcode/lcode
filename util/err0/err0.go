package err0

import (
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func Record(err *error, span trace.Span) {
	if *err != nil {
		span.SetStatus(codes.Error, (*err).Error())
		span.RecordError(*err)
	}
}
