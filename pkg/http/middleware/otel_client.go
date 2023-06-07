package middleware

import (
	"fmt"
	"github.com/innoai-tech/infra/internal/otel"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"net/http"
	"time"

	"github.com/go-courier/logr"
	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel/propagation"
)

func NewLogRoundTripper(clientName string) func(roundTripper http.RoundTripper) http.RoundTripper {
	return func(roundTripper http.RoundTripper) http.RoundTripper {
		return &LogRoundTripper{
			clientName:       clientName,
			nextRoundTripper: roundTripper,
		}
	}
}

type LogRoundTripper struct {
	clientName       string
	nextRoundTripper http.RoundTripper
}

func (rt *LogRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	startedAt := time.Now()

	ctx := req.Context()

	mp := otel.MeterProviderFromContext(ctx)

	h, _ := mp.Meter(rt.clientName).Float64Histogram(
		"http.client.duration",
		metric.WithUnit("s"),
	)

	// inject b3 form context
	b3.New().Inject(ctx, propagation.HeaderCarrier(req.Header))

	ctx, log := logr.Start(ctx, "Request")
	defer log.End()

	resp, err := rt.nextRoundTripper.RoundTrip(req.WithContext(ctx))

	cost := time.Since(startedAt)
	l := log.WithValues(
		semconv.HTTPMethod(req.Method),
		semconv.HTTPURL(omitAuthorization(req.URL)),
		"http.client.duration", fmt.Sprintf("%0.3fms", float64(cost/time.Millisecond)),
	)

	h.Record(ctx, float64(cost)/float64(time.Second), metric.WithAttributes(
		semconv.HTTPMethod(req.Method),
		semconv.HTTPURL(omitAuthorization(req.URL)),
	))

	if resp != nil {
		l = l.WithValues(
			semconv.HTTPStatusCode(resp.StatusCode),
		)
	}

	if req.ContentLength > 0 {
		l = l.WithValues(
			"http.content-type", req.Header.Get("Content-Type"),
			semconv.HTTPResponseContentLength(int(req.ContentLength)),
		)
	}

	if err == nil {
		if resp.StatusCode > http.StatusBadRequest {
			l.Warn(errors.Wrap(err, "http request failed"))
		} else {
			l.Info("success")
		}
	} else {
		l.Warn(errors.Wrap(err, "http request failed"))
	}

	return resp, err
}
