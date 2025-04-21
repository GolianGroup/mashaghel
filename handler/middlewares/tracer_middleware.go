package middlewares

import (
	"errors"
	"runtime"
	"strconv"

	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// TracingMiddleware creates a middleware to trace incoming HTTP requests.
func TracingMiddleware(tracer trace.Tracer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Start a new span
		_, span := tracer.Start(c.Context(), c.Method()+" "+c.OriginalURL())
		defer span.End()

		// Extract TraceID and SpanID
		spanContext := span.SpanContext()
		span.SetAttributes(
			attribute.String("trace.id", spanContext.TraceID().String()),
			attribute.String("span.id", spanContext.SpanID().String()),
		)

		// Trace request method
		span.SetAttributes(attribute.String("http.method", c.Method()))

		// Trace endpoint name
		span.SetAttributes(attribute.String("http.url", c.OriginalURL()))

		// Trace payload size (using Content-Length)
		payloadSize, err := strconv.Atoi(c.Get("Content-Length"))
		if err == nil {
			span.SetAttributes(attribute.Int("http.request.size", payloadSize))
		}

		// Call next handler
		err = c.Next()

		// Trace response status code
		statusCode := c.Response().StatusCode()
		span.SetAttributes(attribute.Int("http.status_code", statusCode))

		// Check if rate-limited
		if statusCode == fiber.StatusTooManyRequests {
			span.SetAttributes(attribute.Bool("http.rate_limited", true))
		}

		// Trace error type if there is an error
		if err != nil {
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				span.SetAttributes(
					attribute.String("error.type", "fiber.Error"),
					attribute.String("error.message", fiberErr.Message),
					attribute.Int("error.status_code", fiberErr.Code),
				)
			} else {
				errorType := fmt.Sprintf("%T", err) // Get the type of the error
				span.SetAttributes(attribute.String("error.type", errorType))
				span.SetAttributes(attribute.String("error.message", err.Error()))
			}
		}

		// Trace system metrics: CPU/memory usage
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		span.SetAttributes(
			attribute.Int64("system.memory.used_bytes", int64(memStats.Alloc)),
			attribute.Int64("system.memory.total_bytes", int64(memStats.Sys)),
			attribute.Int64("system.memory.gc_count", int64(memStats.NumGC)),
		)

		// Trace number of goroutines (threads)
		span.SetAttributes(attribute.Int("system.threads.goroutines", runtime.NumGoroutine()))

		return err
	}
}
