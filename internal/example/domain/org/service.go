package org

import (
	"context"
	"errors"
	"net/http"

	apiv0 "github.com/innoai-tech/infra/internal/example/pkg/apis/org/v0"
	"github.com/octohelm/courier/pkg/statuserror"
)

type Service struct{}

func (Service) Create(ctx context.Context, info *apiv0.Info) (any, error) {
	return nil, nil
}

func (Service) List(ctx context.Context) (*apiv0.DataList, error) {
	return &apiv0.DataList{
		Data: []apiv0.Info{
			{
				Name: "demo",
				Type: apiv0.TYPE__GOV,
			},
		},
		Total: 1,
	}, nil
}

func (Service) Get(ctx context.Context, orgName string) (*apiv0.Detail, error) {
	if orgName == "NotFound" {
		return nil, statuserror.Wrap(errors.New("NotFound"), http.StatusNotFound, "NotFound")
	}

	return &apiv0.Detail{
		Info: apiv0.Info{
			Name: orgName,
			Type: apiv0.TYPE__GOV,
		},
	}, nil
}

func (Service) Delete(ctx context.Context, orgName string) error {
	return nil
}
