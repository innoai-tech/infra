package middleware

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/go-courier/logr"
	"github.com/innoai-tech/infra/pkg/http/middleware/metrichttp"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel/propagation"
)

func NewLogRoundTripper() func(roundTripper http.RoundTripper) http.RoundTripper {
	return func(roundTripper http.RoundTripper) http.RoundTripper {
		return &LogRoundTripper{
			nextRoundTripper: roundTripper,
		}
	}
}

type LogRoundTripper struct {
	nextRoundTripper http.RoundTripper
}

func (rt *LogRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	startedAt := time.Now()

	ctx := req.Context()

	// inject b3 form context
	b3.New().Inject(ctx, propagation.HeaderCarrier(req.Header))

	ctx, log := logr.Start(ctx, "Request")
	defer log.End()

	resp, err := rt.nextRoundTripper.RoundTrip(req.WithContext(ctx))

	cost := time.Since(startedAt)

	l := log.WithValues(
		slog.String("http.method", req.Method),
		slog.String("http.url", omitAuthorization(req.URL)),
		slog.String("http.client.duration", cost.String()),
	)

	if resp != nil {
		p, _ := strconv.ParseInt(req.URL.Port(), 10, 64)

		attrs := []attribute.KeyValue{
			attribute.String("http.request.method", req.Method),
			attribute.Int("http.response.status_code", resp.StatusCode),
			attribute.String("server.address", req.URL.Hostname()),
			attribute.Int("server.port", int(p)),
		}

		metrichttp.ClientDuration.Record(ctx, cost.Seconds(), metric.WithAttributes(attrs...))
		metrichttp.ClientRequestSize.Record(ctx, req.ContentLength, metric.WithAttributes(attrs...))
		metrichttp.ClientResponseSize.Record(ctx, resp.ContentLength, metric.WithAttributes(attrs...))

		l = l.WithValues(
			slog.String("http.method", req.Method),
			slog.Int("http.status_code", resp.StatusCode),
		)
	}

	if req.ContentLength > 0 {
		l = l.WithValues(
			slog.String("http.content-type", req.Header.Get("Content-Type")),
			slog.Int("http.response_content_length", int(req.ContentLength)),
		)
	}

	if err == nil {
		if resp.StatusCode > http.StatusBadRequest {
			l.Warn(errors.New("http request failed"))
		} else {
			l.Info("success")
		}
	} else {
		l.Warn(fmt.Errorf("http request failed: %w", err))
	}

	return resp, err
}
