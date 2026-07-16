package v0

import (
	"context"

	archivedomain "example/domain/archive"
	endpointv0 "example/pkg/endpoints/archive/v0"
)

type ArchiveZip struct {
	endpointv0.ArchiveZip

	archivedomain.Service
}

func (r *ArchiveZip) Output(ctx context.Context) (any, error) {
	return r.Service.ArchiveZip(ctx)
}
