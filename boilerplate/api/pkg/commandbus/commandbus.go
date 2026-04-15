package commandbus

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// resultIDKey adalah context key untuk menyimpan ID entity yang baru dibuat.
type resultIDKey struct{}

// ResultIDHolder menyimpan ID entity yang baru dibuat.
// Command handler mengisi ID; HTTP handler membaca setelah dispatch.
type ResultIDHolder struct {
	ID uuid.UUID
}

// WithResultID mengembalikan context baru dengan ResultIDHolder dan holder-nya.
func WithResultID(ctx context.Context) (context.Context, *ResultIDHolder) {
	h := &ResultIDHolder{}
	return context.WithValue(ctx, resultIDKey{}, h), h
}

// SetCreatedID menyimpan ID entity yang baru dibuat ke dalam context.
func SetCreatedID(ctx context.Context, id uuid.UUID) {
	if h, ok := ctx.Value(resultIDKey{}).(*ResultIDHolder); ok && h != nil {
		h.ID = id
	}
}

// Command adalah interface dasar untuk semua command.
type Command interface {
	CommandName() string
}

// CommandHandler menangani command dengan tipe tertentu.
type CommandHandler[C Command] interface {
	Handle(ctx context.Context, cmd C) error
}

// Middleware membungkus eksekusi command.
type Middleware func(ctx context.Context, cmd Command, next func(ctx context.Context, cmd Command) error) error

// CommandBus mendispatch command ke handler yang terdaftar.
type CommandBus struct {
	handlers   map[string]func(ctx context.Context, cmd Command) error
	middleware []Middleware
	tracer     trace.Tracer
}

func New() *CommandBus {
	return &CommandBus{
		handlers: make(map[string]func(ctx context.Context, cmd Command) error),
		tracer:   otel.Tracer("commandbus"),
	}
}

// Use menambahkan middleware ke command bus.
func (b *CommandBus) Use(m Middleware) {
	b.middleware = append(b.middleware, m)
}

// Register mendaftarkan handler untuk tipe command tertentu.
func Register[C Command](bus *CommandBus, handler CommandHandler[C]) {
	var zero C
	name := zero.CommandName()
	bus.handlers[name] = func(ctx context.Context, cmd Command) error {
		typed, ok := cmd.(C)
		if !ok {
			return fmt.Errorf("commandbus: type mismatch for command %s", name)
		}
		return handler.Handle(ctx, typed)
	}
}

// Dispatch mengirimkan command ke handler yang terdaftar.
func (b *CommandBus) Dispatch(ctx context.Context, cmd Command) error {
	ctx, span := b.tracer.Start(ctx, fmt.Sprintf("command/%s", cmd.CommandName()),
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attribute.String("command.name", cmd.CommandName())),
	)
	defer span.End()

	handler, ok := b.handlers[cmd.CommandName()]
	if !ok {
		return fmt.Errorf("commandbus: no handler registered for command %s", cmd.CommandName())
	}

	chain := handler
	for i := len(b.middleware) - 1; i >= 0; i-- {
		m := b.middleware[i]
		next := chain
		chain = func(ctx context.Context, cmd Command) error {
			return m(ctx, cmd, next)
		}
	}

	start := time.Now()
	err := chain(ctx, cmd)
	duration := time.Since(start)

	span.SetAttributes(
		attribute.Float64("command.duration_ms", float64(duration.Milliseconds())),
		attribute.Bool("command.success", err == nil),
	)
	if err != nil {
		span.RecordError(err)
	}

	return err
}
