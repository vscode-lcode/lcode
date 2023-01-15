package err0

import (
	"context"

	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type contextKey string

var (
	contextKeySpanCode          = contextKey("span code")
	ContextKeySpanStatus        = contextKey("span status")
	ContextKeySpanKeepEndOutput = contextKey("keep span end output")
)

func WithStatus(ctx context.Context, code codes.Code, status string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx = context.WithValue(ctx, contextKeySpanCode, code)
	ctx = context.WithValue(ctx, ContextKeySpanStatus, status)
	return ctx
}

const TheLogHasBeenOutput = "the log has been output"

func KeepEndOutput(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx = context.WithValue(ctx, ContextKeySpanKeepEndOutput, true)
	return ctx
}

func ApplyStatusWithCtx(ctx context.Context, span sdktrace.ReadWriteSpan) {
	code, ok := ctx.Value(contextKeySpanCode).(codes.Code)
	if !ok {
		return
	}
	status, ok := ctx.Value(ContextKeySpanStatus).(string)
	if !ok {
		return
	}
	span.SetStatus(code, status)
	_, ok = ctx.Value(ContextKeySpanKeepEndOutput).(bool)
	if disableEndOutput := !ok; disableEndOutput {
		span.AddEvent(TheLogHasBeenOutput)
	}
}
