package basehref

import (
	"net/http"
	"net/url"
	"path"
	"strings"
)

const (
	HeaderAppBaseHref = "X-App-Base-Href"
)

type BaseHref struct {
	Schema   string
	Host     string
	BasePath string
}

func (h *BaseHref) Path(p string) string {
	return (&url.URL{
		Scheme: h.Schema,
		Host:   h.Host,
		Path:   path.Clean(h.BasePath + p),
	}).String()
}

func (h *BaseHref) Origin() string {
	return (&url.URL{
		Scheme: h.Schema,
		Host:   h.Host,
	}).String()
}

func FromHttpRequest(r *http.Request) *BaseHref {
	b := &BaseHref{}
	b.Schema = "http"
	b.Host = r.Host
	b.BasePath = r.Header.Get(HeaderAppBaseHref)

	if r.TLS != nil {
		b.Schema = "https"
	} else if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		b.Schema = strings.ToLower(proto)
	} else if referer := r.Header.Get("Referer"); referer != "" {
		refererURL, err := url.Parse(referer)
		if err == nil {
			b.Schema = refererURL.Scheme
		}
	}

	return b
}
