package aggregation

import (
	"context"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/util/annotations"
)

func NewQueryable(seriesSet storage.SeriesSet) storage.Queryable {
	return storage.QueryableFunc(func(mint, maxt int64) (storage.Querier, error) {
		return &mockQuerier{seriesSet: seriesSet}, nil
	})
}

type mockQuerier struct {
	seriesSet storage.SeriesSet
}

func (h *mockQuerier) LabelValues(ctx context.Context, name string, matchers ...*labels.Matcher) ([]string, annotations.Annotations, error) {
	return []string{}, nil, nil
}

func (h *mockQuerier) LabelNames(ctx context.Context, matchers ...*labels.Matcher) ([]string, annotations.Annotations, error) {
	return []string{}, nil, nil
}

func (h *mockQuerier) Close() error {
	return nil
}

func (h *mockQuerier) Select(ctx context.Context, sortSeries bool, hints *storage.SelectHints, matchers ...*labels.Matcher) storage.SeriesSet {
	return h.seriesSet
}
