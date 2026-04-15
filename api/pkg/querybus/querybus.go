package querybus

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Query adalah interface dasar untuk semua query.
type Query interface {
	QueryName() string
}

// QueryHandler menangani query dan mengembalikan result.
type QueryHandler[Q Query, R any] interface {
	Handle(ctx context.Context, query Q) (R, error)
}

type handlerFunc func(ctx context.Context, query Query) (any, error)

// QueryBus mendispatch query ke handler yang terdaftar.
type QueryBus struct {
	handlers map[string]handlerFunc
	tracer   trace.Tracer
}

func New() *QueryBus {
	return &QueryBus{
		handlers: make(map[string]handlerFunc),
		tracer:   otel.Tracer("querybus"),
	}
}

// Register mendaftarkan query handler.
func Register[Q Query, R any](bus *QueryBus, handler QueryHandler[Q, R]) {
	var zero Q
	name := zero.QueryName()
	bus.handlers[name] = func(ctx context.Context, query Query) (any, error) {
		typed, ok := query.(Q)
		if !ok {
			return nil, fmt.Errorf("querybus: type mismatch for query %s", name)
		}
		return handler.Handle(ctx, typed)
	}
}

// Dispatch mengirimkan query ke handler dan mengembalikan result.
func Dispatch[R any](ctx context.Context, bus *QueryBus, query Query) (R, error) {
	ctx, span := bus.tracer.Start(ctx, fmt.Sprintf("query/%s", query.QueryName()),
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("query.name", query.QueryName())),
	)
	defer span.End()

	handler, ok := bus.handlers[query.QueryName()]
	if !ok {
		var zero R
		return zero, fmt.Errorf("querybus: no handler registered for query %s", query.QueryName())
	}

	start := time.Now()
	result, err := handler(ctx, query)
	duration := time.Since(start)

	span.SetAttributes(
		attribute.Float64("query.duration_ms", float64(duration.Milliseconds())),
		attribute.Bool("query.success", err == nil),
	)

	if err != nil {
		span.RecordError(err)
		var zero R
		return zero, err
	}

	typed, ok := result.(R)
	if !ok {
		var zero R
		return zero, fmt.Errorf("querybus: handler returned unexpected type for query %s", query.QueryName())
	}

	return typed, nil
}
