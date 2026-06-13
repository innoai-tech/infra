package basehref

import (
	"net/http"
	"net/url"
	"path"
	"strings"
)

// HeaderAppBaseHref 是用于传递应用基础路径的 HTTP 头部名称。
const (
	HeaderAppBaseHref = "X-App-Base-Href"
)

// BaseHref 描述请求的基础 URL 信息，包含协议、主机和基础路径。
type BaseHref struct {
	// Schema 协议，如 http 或 https
	Schema   string
	// Host 主机地址
	Host     string
	// BasePath 基础路径
	BasePath string
}

// Path 返回拼接基础路径后的完整 URL。
func (h *BaseHref) Path(p string) string {
	return (&url.URL{
		Scheme: h.Schema,
		Host:   h.Host,
		Path:   path.Clean(h.BasePath + p),
	}).String()
}

// Origin 返回仅包含协议和主机的根 URL。
func (h *BaseHref) Origin() string {
	return (&url.URL{
		Scheme: h.Schema,
		Host:   h.Host,
	}).String()
}

// FromHttpRequest 从 HTTP 请求中提取 BaseHref 信息。
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
