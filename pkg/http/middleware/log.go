package middleware

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/octohelm/courier/pkg/courierhttp"
	"github.com/octohelm/courier/pkg/courierhttp/util"

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
	logger := log.WithValues(
		"cost", fmt.Sprintf("%0.3fms", float64(cost/time.Millisecond)),
		"method", req.Method,
		"url", omitAuthorization(req.URL),
		"status", resp.StatusCode,
	)

	if req.ContentLength > 0 {
		logger = logger.WithValues(
			"content-type", req.Header.Get("Content-Type"),
			"content-length", FileSize(req.ContentLength),
		)
	}

	if err == nil {
		if resp.StatusCode > http.StatusBadRequest {
			logger.Warn(errors.Wrap(err, "http request failed"))
		} else {
			logger.Info("success")
		}
	} else {
		logger.Warn(errors.Wrap(err, "http request failed"))
	}

	return resp, err
}

func LogHandler() func(handler http.Handler) http.Handler {
	return func(nextHandler http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			ctx := req.Context()

			ctx = b3.New().Extract(ctx, propagation.HeaderCarrier(req.Header))

			startAt := time.Now()

			id := courierhttp.OperationIDFromContext(ctx)

			ctx, span := logr.FromContext(ctx).Start(ctx, id)
			defer func() {
				span.End()
			}()

			loggerRw := newLoggerResponseWriter(rw)

			b3.New().Inject(ctx, propagation.HeaderCarrier(loggerRw.Header()))

			nextHandler.ServeHTTP(loggerRw, req.WithContext(ctx))

			level, _ := logr.ParseLevel(strings.ToLower(req.Header.Get("x-log-level")))

			duration := time.Since(startAt)

			header := req.Header

			keyAndValues := []any{
				"tag", "access",
				"remote_ip", util.ClientIP(req),
				"cost", fmt.Sprintf("%0.3fms", float64(duration/time.Millisecond)),
				"method", req.Method,
				"request_uri", omitAuthorization(req.URL),
				"user_agent", header.Get("User-Agent"),
				"status", loggerRw.statusCode,
			}

			log := logr.FromContext(ctx)

			if loggerRw.err != nil {
				if loggerRw.statusCode >= http.StatusInternalServerError {
					if level >= logr.ErrorLevel {
						log.WithValues(keyAndValues...).Error(loggerRw.err)
					}
				} else {
					if level >= logr.WarnLevel {
						log.WithValues(keyAndValues...).Warn(loggerRw.err)
					}
				}
			} else {
				if level >= logr.InfoLevel {
					log.WithValues(keyAndValues...).Info("")
				}
			}
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
	return rw.ResponseWriter.Write(data)
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
	u.RawQuery = query.Encode()
	return u.String()
}

type FileSize int64

func (f FileSize) String() string {
	b := int64(f)
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
