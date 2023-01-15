package err0

import (
	"github.com/lainio/err2"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func Record(err *error, span trace.Span) {
	err2.Handle(err, func() {
		span.SetStatus(codes.Error, (*err).Error())
		span.RecordError(*err)
	})
}
