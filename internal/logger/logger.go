package logger

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"

	"github.com/google/uuid"
)

type HandlerMiddlware struct {
	next slog.Handler
}

func NewHandlerMiddleware(next slog.Handler) *HandlerMiddlware {
	return &HandlerMiddlware{next: next}
}

func (h *HandlerMiddlware) Enabled(ctx context.Context, rec slog.Level) bool {
	return h.next.Enabled(ctx, rec)
}

type keyType int

const key = keyType(0)

type logCtx struct {
	RequestID uuid.UUID
	JwtId     uuid.UUID
	Id        uuid.UUID
	Name      string
	Email     string
	// ListID uuid.UUID
}

func (h *HandlerMiddlware) Handle(ctx context.Context, rec slog.Record) error {
	if c, ok := ctx.Value(key).(logCtx); ok {
		// if c.ListID != uuid.Nil {
		// 	rec.Add("listID", c.ListID)
		// }
		if c.RequestID != uuid.Nil {
			rec.Add("requestID", c.RequestID)
		}
		if c.JwtId != uuid.Nil {
			rec.Add("jwtID", c.JwtId)
		}
		if c.Id != uuid.Nil {
			rec.Add("id", c.Id)
		}
		if c.Email != "" {
			rec.Add("email", c.Email)
		}
		if c.Name != "" {
			rec.Add("name", c.Name)
		}
	}
	return h.next.Handle(ctx, rec)
}

func (h *HandlerMiddlware) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &HandlerMiddlware{next: h.next.WithAttrs(attrs)}
}

func (h *HandlerMiddlware) WithGroup(name string) slog.Handler {
	return &HandlerMiddlware{next: h.next.WithGroup(name)}
}

// func WithLogListID(ctx context.Context, listID uuid.UUID) context.Context {
// 	if c, ok := ctx.Value(key).(logCtx); ok {
// 		c.ListID = listID
// 		return context.WithValue(ctx, key, c)
// 	}
// 	return context.WithValue(ctx, key, logCtx{ListID: listID})
// }

func WithLogRequestID(ctx context.Context, RequestID uuid.UUID) context.Context {
	if c, ok := ctx.Value(key).(logCtx); ok {
		c.RequestID = RequestID
		return context.WithValue(ctx, key, c)
	}
	return context.WithValue(ctx, key, logCtx{RequestID: RequestID})
}

func WithLogId(ctx context.Context, id uuid.UUID) context.Context {
	if c, ok := ctx.Value(key).(logCtx); ok {
		c.Id = id
		return context.WithValue(ctx, key, c)
	}
	return context.WithValue(ctx, key, logCtx{Id: id})
}

func WithLogJwtId(ctx context.Context, jwtId uuid.UUID) context.Context {
	if c, ok := ctx.Value(key).(logCtx); ok {
		c.JwtId = jwtId
		return context.WithValue(ctx, key, c)
	}
	return context.WithValue(ctx, key, logCtx{JwtId: jwtId})
}

func WithLogEmail(ctx context.Context, email string) context.Context {
	email = strings.TrimSpace(email)

	if len(email) > 5 {
		email = email[:5] + strings.Repeat("*", len(email)-5)
	}
	if c, ok := ctx.Value(key).(logCtx); ok {
		c.Email = email
		return context.WithValue(ctx, key, c)
	}
	return context.WithValue(ctx, key, logCtx{Email: email})
}

func WithLogName(ctx context.Context, name string) context.Context {
	name = strings.TrimSpace(name)
	if c, ok := ctx.Value(key).(logCtx); ok {
		c.Name = name
		return context.WithValue(ctx, key, c)
	}
	return context.WithValue(ctx, key, logCtx{Name: name})
}

type errorWithLogCtx struct {
	next error
	ctx  logCtx
}

func (e *errorWithLogCtx) Error() string {
	return e.next.Error()
}

func (e *errorWithLogCtx) Unwrap() error {
	return e.next
}

func WrapError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	c := logCtx{}
	if x, ok := ctx.Value(key).(logCtx); ok {
		c = x
	}
	return &errorWithLogCtx{
		next: err,
		ctx:  c,
	}
}

func ErrorCtx(ctx context.Context, err error) context.Context {
	var e *errorWithLogCtx
	if errors.As(err, &e) {
		return context.WithValue(ctx, key, e.ctx)
	}
	return ctx
}

func InitLogging(levelString string) {
	level := slog.LevelInfo
	switch strings.ToUpper(strings.TrimSpace(levelString)) {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "ERROR":
		level = slog.LevelError
	case "WARN":
		level = slog.LevelWarn
	default:
		level = slog.LevelInfo
	}

	handler := slog.Handler(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level, AddSource: false}))
	handler = NewHandlerMiddleware(handler)
	slog.SetDefault(slog.New(handler))
}
