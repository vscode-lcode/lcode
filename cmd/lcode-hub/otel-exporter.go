package main

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"sync"
	"text/template"

	"github.com/lainio/err2"
	. "github.com/lainio/err2/try"
	"github.com/vscode-lcode/lcode/v2/util/err0"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

func newTracerProvider(logLevel LogLevel) (tp *sdktrace.TracerProvider) {

	resource := To1(resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("lcode-hub"),
			semconv.ServiceVersionKey.String(VERSION),
		),
	))

	exporter := NewConsoleLogExporter(logLevel)

	tp = sdktrace.NewTracerProvider(
		sdktrace.WithResource(resource),
		sdktrace.WithSpanProcessor(exporter),
	)

	otel.SetTracerProvider(tp)

	return
}

type ConsoleLogExporter struct {
	Level LogLevel
	wg    *sync.WaitGroup
}

var _ sdktrace.SpanProcessor = (*ConsoleLogExporter)(nil)

func NewConsoleLogExporter(level LogLevel) *ConsoleLogExporter {
	return &ConsoleLogExporter{
		Level: level,
		wg:    &sync.WaitGroup{},
	}
}
func (log *ConsoleLogExporter) Shutdown(ctx context.Context) error {
	log.wg.Wait()
	return nil
}

func (log *ConsoleLogExporter) OnStart(ctx context.Context, span sdktrace.ReadWriteSpan) {
	if log.Level == NoneLogLevel {
		return
	}
	err0.ApplyStatusWithCtx(ctx, span)
	if code := span.Status().Code; code != codes.Unset {
		log.wg.Add(1)
		go func() {
			defer log.wg.Done()
			log.PrintlnSpan(span)
		}()
	} else {
		return
	}
}

func (log *ConsoleLogExporter) OnEnd(span sdktrace.ReadOnlySpan) {
	log.wg.Add(1)
	go func() {
		defer log.wg.Done()
		log.PrintlnSpan(span)
	}()
}

func (log *ConsoleLogExporter) ForceFlush(ctx context.Context) error { return nil }

var tpl = To1(template.New("").Parse("[{{.scope}}] {{.name}} {{.attrs}} {{.status}}"))

type LogLevel int

const (
	NoneLogLevel  LogLevel = 0
	ErrorLogLevel LogLevel = 1 << (iota - 1)
	InfoLogLevel
	DebugLogLevel
)

func (log ConsoleLogExporter) shouldLog(span sdktrace.ReadOnlySpan) bool {
	if _, onStart := span.(sdktrace.ReadWriteSpan); !onStart { // onEnd
		for _, ev := range span.Events() {
			if ev.Name == err0.TheLogHasBeenOutput {
				return false
			}
		}
	}
	switch s := span.Status(); s.Code {
	case codes.Unset:
		return log.Level&DebugLogLevel != 0
	case codes.Ok:
		return log.Level&InfoLogLevel != 0
	case codes.Error:
		return log.Level&ErrorLogLevel != 0
	}
	return false
}

func (log ConsoleLogExporter) PrintlnSpan(span sdktrace.ReadOnlySpan) (err error) {
	defer err2.Handle(&err)
	if !log.shouldLog(span) {
		return
	}

	var assigns = map[string]string{}
	output := os.Stderr
	switch s := span.Status(); s.Code {
	case codes.Unset: //debug
		assigns["status"] = "desc: " + s.Description
	case codes.Ok: //info
		output = os.Stdout
		assigns["status"] = "msg: " + s.Description
	case codes.Error: //error
		assigns["status"] = "err: " + s.Description
	}

	assigns["name"] = "span: " + span.Name()

	scope := span.InstrumentationScope()
	scope.Name = strings.TrimPrefix(scope.Name, "github.com/vscode-lcode/lcode/v2")
	assigns["scope"] = scope.Name

	var attrs = map[string]any{}
	for _, kv := range span.Attributes() {
		k := string(kv.Key)
		attrs[k] = kv.Value.AsInterface()
	}
	b := To1(json.Marshal(attrs))
	assigns["attrs"] = "attrs: " + string(b)

	_, onStart := span.(sdktrace.ReadWriteSpan)
	var endStr = "\n"
	if onEnd := !onStart; onEnd {
		endStr = " end\n"
	}

	To(tpl.Execute(output, assigns))
	io.WriteString(output, endStr)
	return
}
