package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

type eventKey struct{}

type wideEvent struct {
	mu     sync.Mutex
	fields []any
}

func (e *wideEvent) add(key string, val any) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.fields = append(e.fields, slog.Any(key, val))
}

// Enrich adds a field to the wide event stored in ctx.
// No-op if no event is present (e.g. in tests).
func Enrich(ctx context.Context, key string, val any) {
	if e, ok := ctx.Value(eventKey{}).(*wideEvent); ok {
		e.add(key, val)
	}
}

func wideEventMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			event := &wideEvent{}
			event.add("method", c.Request().Method)
			event.add("path", c.Request().URL.Path)
			event.add("request_id", requestID(c.Request()))

			ctx := context.WithValue(c.Request().Context(), eventKey{}, event)
			c.SetRequest(c.Request().WithContext(ctx))

			err := next(c)

			status := c.Response().Status
			if status == http.StatusSwitchingProtocols {
				return err // WS connection — hijacked, nothing to log as HTTP
			}
			if err != nil {
				event.add("error", err.Error())
			}
			event.add("status", status)
			event.add("duration_ms", time.Since(start).Milliseconds())

			if status >= 500 || err != nil {
				slog.Error("request", event.fields...)
			} else {
				slog.Info("request", event.fields...)
			}
			return err
		}
	}
}

func requestID(r *http.Request) string {
	if id := r.Header.Get("X-Request-Id"); id != "" {
		return id
	}
	return "-"
}
