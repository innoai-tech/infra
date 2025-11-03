package http

import (
	"encoding/base64"
	"net/http"
	"strings"
	"sync"

	openapiview "github.com/innoai-tech/openapi-playground"
	"github.com/octohelm/courier/pkg/courierhttp/handler/httprouter"

	"github.com/innoai-tech/infra/pkg/http/basehref"
	"github.com/innoai-tech/infra/pkg/http/webapp"
)

func init() {
	httprouter.SetOpenAPIViewContents(&openapiView{})
}

type openapiView struct {
	views sync.Map
}

func (v *openapiView) Upgrade(w http.ResponseWriter, r *http.Request) error {
	basePath := strings.Split(r.URL.Path, "/_view/")[0]

	getHandler, _ := v.views.LoadOrStore(basePath, sync.OnceValue(func() http.Handler {
		return webapp.ServeFS(
			openapiview.Contents,
			webapp.WithBaseHref(basePath+"/_view/"),
			webapp.WithAppConfig(map[string]string{
				"OPENAPI": base64.StdEncoding.EncodeToString([]byte(basePath)),
			}),
		)
	}))

	// openapi playground should ignore HeaderAppBaseHref
	r.Header.Del(basehref.HeaderAppBaseHref)

	getHandler.(func() http.Handler)().ServeHTTP(w, r)

	return nil
}
