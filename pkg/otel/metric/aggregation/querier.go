package aggregation

import (
	"context"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/storage"
)

func NewQueryable(seriesSet storage.SeriesSet) storage.Queryable {
	return storage.QueryableFunc(func(ctx context.Context, mint, maxt int64) (storage.Querier, error) {
		return &mockQuerier{seriesSet: seriesSet}, nil
	})
}

type mockQuerier struct {
	seriesSet storage.SeriesSet
}

func (h *mockQuerier) LabelValues(name string, matchers ...*labels.Matcher) ([]string, storage.Warnings, error) {
	return []string{}, nil, nil
}

func (h *mockQuerier) LabelNames(matchers ...*labels.Matcher) ([]string, storage.Warnings, error) {
	return []string{}, nil, nil
}

func (h *mockQuerier) Close() error {
	return nil
}

func (h *mockQuerier) Select(sortSeries bool, hints *storage.SelectHints, matchers ...*labels.Matcher) storage.SeriesSet {
	return h.seriesSet
}
