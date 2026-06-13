package middleware

import (
	"net/http"
	"slices"
	"strconv"
	"strings"
)

// DefaultCORS 创建使用默认宽松策略的 CORS 中间件。
//
// 源自 github.com/gorilla/handlers 的 CORS 实现。
func DefaultCORS(opts ...CORSOption) func(http.Handler) http.Handler {
	return CORS(
		append([]CORSOption{
			AllowedOrigins([]string{"*"}),
			AllowedMethods([]string{
				http.MethodConnect,
				http.MethodPut,
				http.MethodPatch,
				http.MethodDelete,
				http.MethodOptions,
			}),
			AllowCredentials(),
			AllowedHeaders([]string{
				CorsRequestMethodHeader,
				CorsRequestHeadersHeader,
				"Content-Type",
				"Authorization",
				"User-Agent",
			}),
			ExposedHeaders([]string{
				"Content-Type",
				"Origin",
				"B3",
				"WWW-Authenticate",
				"Location",
				"X-Requested-With",
				"X-RateLimit-Limit", // 遵循 GitHub API 速率限制规范
				"X-RateLimit-Remaining",
				"X-RateLimit-Reset",
			}),
			OptionStatusCode(http.StatusNoContent),
		}, opts...)...,
	)
}

// CORSOption 表示用于配置 CORS 中间件的函数选项。
type CORSOption func(*cors) error

type cors struct {
	h                      http.Handler
	allowedHeaders         []string
	allowedMethods         []string
	allowedOrigins         []string
	allowedOriginValidator OriginValidator
	exposedHeaders         []string
	maxAge                 int
	ignoreOptions          bool
	allowCredentials       bool
	optionStatusCode       int
}

// OriginValidator 接收一个 origin 字符串，返回该 origin 是否被允许。
type OriginValidator func(string) bool

var (
	defaultCorsOptionStatusCode = 200
	defaultCorsMethods          = []string{"GET", "HEAD", "POST"}
	defaultCorsHeaders          = []string{"Accept", "Accept-Language", "Content-Language", "Origin"}
)

// CORS 相关的 HTTP 头部常量。
const (
	// CorsOptionMethod 表示 CORS 预检请求的方法名。
	CorsOptionMethod string = "OPTIONS"
	// CorsAllowOriginHeader 表示允许的来源头部名。
	CorsAllowOriginHeader string = "Access-Control-Allow-Origin"
	// CorsExposeHeadersHeader 表示暴露的头部名。
	CorsExposeHeadersHeader string = "Access-Control-Expose-Headers"
	// CorsMaxAgeHeader 表示预检请求缓存时间头部名。
	CorsMaxAgeHeader string = "Access-Control-Max-Age"
	// CorsAllowMethodsHeader 表示允许的方法头部名。
	CorsAllowMethodsHeader string = "Access-Control-Allow-Methods"
	// CorsAllowHeadersHeader 表示允许的请求头部名。
	CorsAllowHeadersHeader string = "Access-Control-Allow-Headers"
	// CorsAllowCredentialsHeader 表示允许凭证的头部名。
	CorsAllowCredentialsHeader string = "Access-Control-Allow-Credentials"
	// CorsRequestMethodHeader 表示请求的方法头部名。
	CorsRequestMethodHeader string = "Access-Control-Request-Method"
	// CorsRequestHeadersHeader 表示请求的头部名。
	CorsRequestHeadersHeader string = "Access-Control-Request-Headers"
	// CorsOriginHeader 表示来源头部名。
	CorsOriginHeader string = "Origin"
	// CorsVaryHeader 表示 Vary 头部名。
	CorsVaryHeader string = "Vary"
	// CorsOriginMatchAll 表示允许所有来源的通配符。
	CorsOriginMatchAll string = "*"
)

func (ch *cors) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get(CorsOriginHeader)
	if !ch.isOriginAllowed(origin) {
		if r.Method != CorsOptionMethod || ch.ignoreOptions {
			ch.h.ServeHTTP(w, r)
		}
		return
	}

	if r.Method == CorsOptionMethod {
		if ch.ignoreOptions {
			ch.h.ServeHTTP(w, r)
			return
		}

		if _, ok := r.Header[CorsRequestMethodHeader]; !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		method := r.Header.Get(CorsRequestMethodHeader)
		if !ch.isMatch(method, ch.allowedMethods) {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		requestHeaders := strings.Split(r.Header.Get(CorsRequestHeadersHeader), ",")
		allowedHeaders := make([]string, 0)
		for _, v := range requestHeaders {
			canonicalHeader := http.CanonicalHeaderKey(strings.TrimSpace(v))
			if canonicalHeader == "" {
				continue
			}

			if !ch.isMatch(canonicalHeader, ch.allowedHeaders) {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			allowedHeaders = append(allowedHeaders, canonicalHeader)
		}

		if len(allowedHeaders) > 0 {
			w.Header().Set(CorsAllowHeadersHeader, strings.Join(allowedHeaders, ","))
		}

		if ch.maxAge > 0 {
			w.Header().Set(CorsMaxAgeHeader, strconv.Itoa(ch.maxAge))
		}

		if !slices.Contains(defaultCorsMethods, method) {
			w.Header().Set(CorsAllowMethodsHeader, method)
		}
	} else {
		if len(ch.exposedHeaders) > 0 {
			w.Header().Set(CorsExposeHeadersHeader, strings.Join(ch.exposedHeaders, ","))
		}
	}

	if ch.allowCredentials {
		w.Header().Set(CorsAllowCredentialsHeader, "true")
	}

	if len(ch.allowedOrigins) > 1 {
		w.Header().Set(CorsVaryHeader, CorsOriginHeader)
	}

	returnOrigin := origin
	if ch.allowedOriginValidator == nil && len(ch.allowedOrigins) == 0 {
		returnOrigin = "*"
	} else {
		if slices.Contains(ch.allowedOrigins, CorsOriginMatchAll) {
			returnOrigin = "*"
		}
	}

	w.Header().Set(CorsAllowOriginHeader, returnOrigin)

	if r.Method == CorsOptionMethod {
		w.WriteHeader(ch.optionStatusCode)
		return
	}
	ch.h.ServeHTTP(w, r)
}

// CORS 提供跨域资源共享中间件。
func CORS(opts ...CORSOption) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		ch := parseCORSOptions(opts...)
		ch.h = h
		return ch
	}
}

func parseCORSOptions(opts ...CORSOption) *cors {
	ch := &cors{
		allowedMethods:   defaultCorsMethods,
		allowedHeaders:   defaultCorsHeaders,
		allowedOrigins:   []string{},
		optionStatusCode: defaultCorsOptionStatusCode,
	}

	for _, option := range opts {
		_ = option(ch)
	}

	return ch
}

// CORS 的函数选项配置。

// AllowedHeaders 将给定头部追加到 CORS 请求的允许头部列表中。
// 这是一个追加操作，Accept、Accept-Language 和 Content-Language 始终被允许。
// 如需接收 application/x-www-form-urlencoded、multipart/form-data 或 text/plain
// 以外的 Content-Type，则必须显式声明 Content-Type。
func AllowedHeaders(headers []string) CORSOption {
	return func(ch *cors) error {
		for _, v := range append(defaultCorsHeaders, headers...) {
			normalizedHeader := http.CanonicalHeaderKey(strings.TrimSpace(v))
			if normalizedHeader == "" {
				continue
			}

			if !ch.isMatch(normalizedHeader, ch.allowedHeaders) {
				ch.allowedHeaders = append(ch.allowedHeaders, normalizedHeader)
			}
		}

		return nil
	}
}

// AllowedMethods 用于在 Access-Control-Allow-Methods 头部中显式允许指定方法。
// 这是一个替换操作，因此如需支持 GET、HEAD 和 POST 方法，也必须一并传入。
func AllowedMethods(methods []string) CORSOption {
	return func(ch *cors) error {
		ch.allowedMethods = []string{}
		for _, v := range append(defaultCorsMethods, methods...) {
			normalizedMethod := strings.ToUpper(strings.TrimSpace(v))
			if normalizedMethod == "" {
				continue
			}

			if !ch.isMatch(normalizedMethod, ch.allowedMethods) {
				ch.allowedMethods = append(ch.allowedMethods, normalizedMethod)
			}
		}

		return nil
	}
}

// AllowedOrigins 设置 CORS 请求的允许来源，对应 'Access-Control-Allow-Origin' HTTP 头部。
// 注意：传入 []string{"*"} 将允许任意域名。
func AllowedOrigins(origins []string) CORSOption {
	return func(ch *cors) error {
		if slices.Contains(origins, CorsOriginMatchAll) {
			ch.allowedOrigins = []string{CorsOriginMatchAll}
			return nil
		}

		ch.allowedOrigins = origins
		return nil
	}
}

// AllowedOriginValidator 设置校验 CORS 请求来源的函数，对应 'Access-Control-Allow-Origin' HTTP 头部。
func AllowedOriginValidator(fn OriginValidator) CORSOption {
	return func(ch *cors) error {
		ch.allowedOriginValidator = fn
		return nil
	}
}

// OptionStatusCode 设置 OPTIONS 请求的自定义状态码。
// 默认行为遵循最佳实践设为 200。此选项非必选，可用于设置自定义状态码（如 204）。
//
// 规范详见：
// https://fetch.spec.whatwg.org/#cors-preflight-fetch
func OptionStatusCode(code int) CORSOption {
	return func(ch *cors) error {
		ch.optionStatusCode = code
		return nil
	}
}

// ExposedHeaders 指定对客户端可见且不会被 user-agent 过滤掉的响应头部。
func ExposedHeaders(headers []string) CORSOption {
	return func(ch *cors) error {
		ch.exposedHeaders = []string{}
		for _, v := range headers {
			normalizedHeader := http.CanonicalHeaderKey(strings.TrimSpace(v))
			if normalizedHeader == "" {
				continue
			}

			if !ch.isMatch(normalizedHeader, ch.exposedHeaders) {
				ch.exposedHeaders = append(ch.exposedHeaders, normalizedHeader)
			}
		}

		return nil
	}
}

// MaxAge 决定预检请求之间的最大缓存时间（秒）。最大允许 10 分钟，超过该值将默认使用 10 分钟。
func MaxAge(age int) CORSOption {
	return func(ch *cors) error {
		// 最长 10 分钟。
		if age > 600 {
			age = 600
		}

		ch.maxAge = age
		return nil
	}
}

// IgnoreOptions 使 CORS 中间件忽略 OPTIONS 请求，将其透传给下一个处理器。
// 适用于应用或框架已有处理 OPTIONS 请求的机制时。
func IgnoreOptions() CORSOption {
	return func(ch *cors) error {
		ch.ignoreOptions = true
		return nil
	}
}

// AllowCredentials 指定允许 user-agent 随请求传递认证凭据。
func AllowCredentials() CORSOption {
	return func(ch *cors) error {
		ch.allowCredentials = true
		return nil
	}
}

func (ch *cors) isOriginAllowed(origin string) bool {
	if origin == "" {
		return false
	}

	if ch.allowedOriginValidator != nil {
		return ch.allowedOriginValidator(origin)
	}

	if len(ch.allowedOrigins) == 0 {
		return true
	}

	for _, allowedOrigin := range ch.allowedOrigins {
		if allowedOrigin == origin || allowedOrigin == CorsOriginMatchAll {
			return true
		}
	}

	return false
}

func (ch *cors) isMatch(needle string, haystack []string) bool {
	return slices.Contains(haystack, needle)
}
