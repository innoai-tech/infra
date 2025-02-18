/*
Package http GENERATED BY gengo:injectable
DON'T EDIT THIS FILE
*/
package http

import (
	context "context"

	appinfo "github.com/innoai-tech/infra/pkg/appinfo"
)

func (v *Server) Init(ctx context.Context) error {
	if value, ok := appinfo.InfoFromContext(ctx); ok {
		v.info = value
	}

	if err := v.afterInit(ctx); err != nil {
		return err
	}

	return nil
}
