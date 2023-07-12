package middleware

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-courier/logr"
	"github.com/innoai-tech/infra/pkg/http/middleware/metrichttp"
	"github.com/octohelm/courier/pkg/courierhttp"
	"github.com/octohelm/courier/pkg/courierhttp/util"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
)

func MetricHandler(gatherer prometheus.Gatherer) func(handler http.Handler) http.Handler {
	h := promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})

	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if strings.HasPrefix(req.URL.Path, "/.sys/metrics") {
				h.ServeHTTP(rw, req)
				return
			}
			handler.ServeHTTP(rw, req)
		})
	}
}

func httpBasicAttrs(req *http.Request) []attribute.KeyValue {
	if req.URL.Scheme == "" {
		req.URL.Scheme = "http"
	}
	if req.URL.Host == "" {
		req.URL.Host = req.Host
	}

	port, _ := strconv.ParseInt(req.URL.Port(), 10, 64)

	return []attribute.KeyValue{
		attribute.Key("http.request.method").String(req.Method),
		attribute.Key("url.schema").String(req.URL.Scheme),
		attribute.Key("server.address").String(req.URL.Hostname()),
		attribute.Key("server.port").Int(int(port)),
	}
}

func httpRouteAttrs(statusCode int, info courierhttp.OperationInfo, req *http.Request) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.Key("http.route").String(info.Route),
		attribute.Key("http.request.method").String(req.Method),
		attribute.Key("http.response.status_code").Int(statusCode),
		attribute.Key("network.protocol.name").String(strings.ToLower(strings.Split(req.Proto, "/")[0])),
		attribute.Key("network.protocol.version").String(fmt.Sprintf("%d.%d", req.ProtoMajor, req.ProtoMinor)),
	}
}

func LogAndMetricHandler() func(handler http.Handler) http.Handler {
	return func(nextHandler http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			ctx = b3.New().Extract(ctx, propagation.HeaderCarrier(req.Header))

			startAt := time.Now()

			info := courierhttp.OperationInfoFromContext(ctx)
			ctx, span := logr.FromContext(ctx).Start(ctx, info.ID)
			defer func() {
				span.End()
			}()

			metricBasicAttrs := httpBasicAttrs(req)

			metrichttp.ServerActiveRequest.Add(ctx, 1, metric.WithAttributes(metricBasicAttrs...))
			defer func() {
				metrichttp.ServerActiveRequest.Add(ctx, -1, metric.WithAttributes(metricBasicAttrs...))
			}()

			loggerRw := newLoggerResponseWriter(rw)

			b3.New().Inject(ctx, propagation.HeaderCarrier(loggerRw.Header()))

			nextHandler.ServeHTTP(loggerRw, req.WithContext(ctx))

			enabledLevel, _ := logr.ParseLevel(strings.ToLower(req.Header.Get("x-log-level")))

			requestCost := time.Since(startAt)
			requestHeader := req.Header

			keyAndValues := []any{
				semconv.HTTPClientIP(util.ClientIP(req)),
				semconv.HTTPMethod(req.Method),
				semconv.HTTPURL(omitAuthorization(req.URL)),
				semconv.HTTPStatusCode(loggerRw.statusCode),
				semconv.UserAgentOriginal(requestHeader.Get("User-Agent")),
				"http.server.duration", fmt.Sprintf("%s", requestCost),
			}

			l := logr.FromContext(ctx)

			if loggerRw.err != nil {
				if loggerRw.statusCode >= http.StatusInternalServerError {
					l.WithValues(keyAndValues...).Error(loggerRw.err)
				} else {
					if enabledLevel <= logr.WarnLevel {
						l.WithValues(keyAndValues...).Warn(loggerRw.err)
					}
				}
			} else {
				if enabledLevel <= logr.InfoLevel {
					l.WithValues(keyAndValues...).Info("success")
				}
			}

			metricsAttrs := append(metricBasicAttrs, httpRouteAttrs(loggerRw.statusCode, info, req)...)

			metrichttp.ServerDuration.Record(ctx, requestCost.Seconds(), metric.WithAttributes(metricsAttrs...))
			metrichttp.ServerRequestSize.Record(ctx, req.ContentLength, metric.WithAttributes(metricsAttrs...))
			metrichttp.ServerResponseSize.Record(ctx, loggerRw.written, metric.WithAttributes(metricsAttrs...))
		})
	}
}

func newLoggerResponseWriter(rw http.ResponseWriter) *loggerResponseWriter {
	h, hok := rw.(http.Hijacker)
	if !hok {
		h = nil
	}

	f, fok := rw.(http.Flusher)
	if !fok {
		f = nil
	}

	return &loggerResponseWriter{
		ResponseWriter: rw,
		Hijacker:       h,
		Flusher:        f,
	}
}

type loggerResponseWriter struct {
	http.ResponseWriter
	http.Hijacker
	http.Flusher

	headerWritten bool
	statusCode    int
	written       int64
	err           error
}

func (rw *loggerResponseWriter) WriteError(err error) {
	rw.err = err
}

func (rw *loggerResponseWriter) Header() http.Header {
	return rw.ResponseWriter.Header()
}

func (rw *loggerResponseWriter) WriteHeader(statusCode int) {
	rw.writeHeader(statusCode)
}

func (rw *loggerResponseWriter) Write(data []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.WriteHeader(http.StatusOK)
	}
	if rw.err == nil && rw.statusCode >= http.StatusBadRequest {
		rw.err = errors.New(string(data))
	}
	n, err := rw.ResponseWriter.Write(data)
	rw.written += int64(n)
	return n, err
}

func (rw *loggerResponseWriter) writeHeader(statusCode int) {
	if !rw.headerWritten {
		rw.ResponseWriter.WriteHeader(statusCode)
		rw.statusCode = statusCode
		rw.headerWritten = true
	}
}

func omitAuthorization(u *url.URL) string {
	query := u.Query()

	query.Del("authorization")
	query.Del("x-param-header-Authorization")

	u.RawQuery = query.Encode()
	return u.String()
}
