package v0

import (
	"context"
	"net/http"

	orgdomain "github.com/innoai-tech/infra/internal/example/domain/org"
	apiv0 "github.com/innoai-tech/infra/internal/example/pkg/apis/org/v0"
	endpointv0 "github.com/innoai-tech/infra/internal/example/pkg/endpoints/org/v0"
	"github.com/octohelm/courier/pkg/courierhttp"
)

type CreateOrg struct {
	endpointv0.CreateOrg

	orgdomain.Service
}

func (r *CreateOrg) Output(ctx context.Context) (any, error) {
	return r.Service.Create(ctx, &r.Body)
}

type DeleteOrg struct {
	endpointv0.DeleteOrg

	orgdomain.Service
}

func (r *DeleteOrg) Output(ctx context.Context) (any, error) {
	return nil, r.Service.Delete(ctx, r.OrgName)
}

type GetOrg struct {
	endpointv0.GetOrg

	orgdomain.Service
}

func (r *GetOrg) Output(ctx context.Context) (any, error) {
	return r.Service.Get(ctx, r.OrgName)
}

type ListOrg struct {
	endpointv0.ListOrg

	orgdomain.Service
}

func (r *ListOrg) Output(ctx context.Context) (any, error) {
	data, err := r.Service.List(ctx)
	if err != nil {
		return nil, err
	}

	return courierhttp.Wrap(
		data,
		courierhttp.WithStatusCode(http.StatusOK),
		courierhttp.WithMetadata("X-Custom", "X"),
	), nil
}

var _ = apiv0.Info{}
