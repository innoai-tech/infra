package v0

import (
	"context"

	archivedomain "github.com/innoai-tech/infra/internal/example/domain/archive"
	endpointv0 "github.com/innoai-tech/infra/internal/example/pkg/endpoints/archive/v0"
)

type ArchiveZip struct {
	endpointv0.ArchiveZip

	archivedomain.Service
}

func (r *ArchiveZip) Output(ctx context.Context) (any, error) {
	return r.Service.ArchiveZip(ctx)
}
