package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/innoai-tech/infra/pkg/http/middleware/metrichttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"

	"github.com/go-courier/logr"
	"github.com/pkg/errors"
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
		semconv.HTTPMethod(req.Method),
		semconv.HTTPURL(omitAuthorization(req.URL)),
		attribute.Key("http.client.duration").String(fmt.Sprintf("%s", cost)),
	)

	if resp != nil {
		p, _ := strconv.ParseInt(req.URL.Port(), 10, 64)

		attrs := []attribute.KeyValue{
			attribute.Key("http.request.method").String(req.Method),
			attribute.Key("http.response.status_code").Int(resp.StatusCode),
			attribute.Key("server.address").String(req.URL.Hostname()),
			attribute.Key("server.port").Int(int(p)),
		}

		metrichttp.ClientDuration.Record(ctx, cost.Seconds(), metric.WithAttributes(attrs...))
		metrichttp.ClientRequestSize.Record(ctx, req.ContentLength, metric.WithAttributes(attrs...))
		metrichttp.ClientResponseSize.Record(ctx, resp.ContentLength, metric.WithAttributes(attrs...))

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
			l.Warn(errors.New("http request failed"))
		} else {
			l.Info("success")
		}
	} else {
		l.Warn(errors.Wrap(err, "http request failed"))
	}

	return resp, err
}
