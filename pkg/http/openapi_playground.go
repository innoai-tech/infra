package http

import (
	"encoding/base64"
	"net/http"
	"strings"
	"sync"

	"github.com/innoai-tech/infra/pkg/http/webapp"
	openapiview "github.com/innoai-tech/openapi-playground"
	"github.com/octohelm/courier/pkg/courierhttp/handler/httprouter"
)

func init() {
	httprouter.SetOpenAPIViewContents(&openapiView{})
}

type openapiView struct {
	once    sync.Once
	handler http.Handler
}

func (v *openapiView) Upgrade(w http.ResponseWriter, r *http.Request) error {
	v.once.Do(func() {
		basePath := strings.Split(r.URL.Path, "/_view/")[0]

		v.handler = webapp.ServeFS(
			openapiview.Contents,
			webapp.WithBaseHref(basePath+"/_view/"),
			webapp.WithAppConfig(map[string]string{
				"OPENAPI": base64.StdEncoding.EncodeToString([]byte(basePath)),
			}),
		)
	})

	v.handler.ServeHTTP(w, r)

	return nil
}
