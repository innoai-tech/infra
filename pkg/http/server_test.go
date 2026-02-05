package http_test

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/octohelm/exp/xiter"

	"github.com/octohelm/x/cmp"
	. "github.com/octohelm/x/testing/v2"

	exampleapis "github.com/innoai-tech/infra/cmd/example/apis"
	"github.com/innoai-tech/infra/pkg/configuration"
	"github.com/innoai-tech/infra/pkg/configuration/testingutil"
	infrahttp "github.com/innoai-tech/infra/pkg/http"
	"github.com/innoai-tech/infra/pkg/otel"
	"github.com/innoai-tech/infra/pkg/otel/openmetrics"
)

func TestServer(t *testing.T) {
	ctx, c := testingutil.BuildContext(t, func(c *struct {
		otel.Otel

		infrahttp.Server
	},
	) {
		c.Addr = ":0"
		c.Server.ApplyRouter(exampleapis.R)
	})

	cctx, cancel := context.WithCancel(ctx)
	go func() {
		_ = configuration.RunOrServe(cctx, c)
	}()
	t.Cleanup(cancel)

	t.Run("WHEN request some api", func(t *testing.T) {
		Must(t, func() error {
			_, err := http.Get(c.Server.Endpoint() + "/api/example")
			return err
		})

		metricFamilySet := MustValue(t, func() (openmetrics.MetricFamilySet, error) {
			resp, err := http.Get(c.Server.Endpoint() + "/.sys/metrics")
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()
			raw, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			return openmetrics.Parse(raw)
		})

		filterMetricsCount := func(filter func(m *openmetrics.Metric) bool) int {
			return xiter.Count(xiter.Filter(
				metricFamilySet.Metrics(),
				filter,
			))
		}

		Then(t, "contains expect metrics",
			Expect(
				filterMetricsCount(openmetrics.Named("process_cpu_time_total")),
				Be(cmp.Gte(1)),
			),
			Expect(
				filterMetricsCount(openmetrics.Named("go_memory_used")),
				Be(cmp.Gte(1)),
			),
			Expect(
				filterMetricsCount(openmetrics.Named("go_goroutine_count")),
				Be(cmp.Gte(1)),
			),

			Expect(
				filterMetricsCount(openmetrics.All(
					openmetrics.Named("http_server_duration_count"),
					openmetrics.Labeled("http_request_method", "GET"),
					openmetrics.Labeled("http_route", "/api/example"),
				)),
				Be(cmp.Gte(1)),
			),

			Expect(
				filterMetricsCount(openmetrics.All(
					openmetrics.Named("http_server_duration_bucket"),
					openmetrics.Labeled("le", "0.005"),
				)),
				Be(cmp.Gte(1)),
			),
			Expect(
				filterMetricsCount(openmetrics.All(
					openmetrics.Named("http_server_duration_bucket"),
					openmetrics.Labeled("le", "+Inf"),
				)),
				Be(cmp.Gte(1)),
			),
		)
	})
}
