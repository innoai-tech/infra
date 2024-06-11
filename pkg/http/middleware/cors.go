package middleware

// copy from https://github.com/gorilla/handlers/blob/master/cors.go

import (
	"net/http"
	"strconv"
	"strings"
)

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
				"X-RateLimit-Limit", // follow https://developer.github.com/v3/rate_limit/
				"X-RateLimit-Remaining",
				"X-RateLimit-Reset",
			}),
			OptionStatusCode(http.StatusNoContent),
		}, opts...)...,
	)
}

// CORSOption represents a functional option for configuring the CORS middleware.
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

// OriginValidator takes an origin string and returns whether or not that origin is allowed.
type OriginValidator func(string) bool

var (
	defaultCorsOptionStatusCode = 200
	defaultCorsMethods          = []string{"GET", "HEAD", "POST"}
	defaultCorsHeaders          = []string{"Accept", "Accept-Language", "Content-Language", "Origin"}
)

const (
	CorsOptionMethod           string = "OPTIONS"
	CorsAllowOriginHeader      string = "Access-Control-Allow-Origin"
	CorsExposeHeadersHeader    string = "Access-Control-Expose-Headers"
	CorsMaxAgeHeader           string = "Access-Control-Max-Age"
	CorsAllowMethodsHeader     string = "Access-Control-Allow-Methods"
	CorsAllowHeadersHeader     string = "Access-Control-Allow-Headers"
	CorsAllowCredentialsHeader string = "Access-Control-Allow-Credentials"
	CorsRequestMethodHeader    string = "Access-Control-Request-Method"
	CorsRequestHeadersHeader   string = "Access-Control-Request-Headers"
	CorsOriginHeader           string = "Origin"
	CorsVaryHeader             string = "Vary"
	CorsOriginMatchAll         string = "*"
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

		if !ch.isMatch(method, defaultCorsMethods) {
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
		for _, o := range ch.allowedOrigins {
			// A configuration of * is different than explicitly setting an allowed
			// origin. Returning arbitrary origin headers in an access control allow
			// origin header is unsafe and is not required by any use case.
			if o == CorsOriginMatchAll {
				returnOrigin = "*"
				break
			}
		}
	}

	w.Header().Set(CorsAllowOriginHeader, returnOrigin)

	if r.Method == CorsOptionMethod {
		w.WriteHeader(ch.optionStatusCode)
		return
	}
	ch.h.ServeHTTP(w, r)
}

// CORS provides Cross-Origin Resource Sharing middleware.
// Example:
//
//	import (
//	    "net/http"
//
//	    "github.com/gorilla/handlers"
//	    "github.com/gorilla/mux"
//	)
//
//	func main() {
//	    r := mux.NewRouter()
//	    r.HandleFunc("/users", UserEndpoint)
//	    r.HandleFunc("/projects", ProjectEndpoint)
//
//	    // Apply the CORS middleware to our top-level router, with the defaults.
//	    http.ListenAndServe(":8000", handlers.CORS()(r))
//	}
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

//
// Functional options for configuring CORS.
//

// AllowedHeaders adds the provided headers to the list of allowed headers in a
// CORS request.
// This is an append operation so the headers Accept, Accept-Language,
// and Content-Language are always allowed.
// Content-Type must be explicitly declared if accepting Content-Types other than
// application/x-www-form-urlencoded, multipart/form-data, or text/plain.
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

// AllowedMethods can be used to explicitly allow methods in the
// Access-Control-Allow-Methods header.
// This is a replacement operation so you must also
// pass GET, HEAD, and POST if you wish to support those methods.
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

// AllowedOrigins sets the allowed origins for CORS requests, as used in the
// 'Allow-Access-Control-Origin' HTTP header.
// Note: Passing in a []string{"*"} will allow any domain.
func AllowedOrigins(origins []string) CORSOption {
	return func(ch *cors) error {
		for _, v := range origins {
			if v == CorsOriginMatchAll {
				ch.allowedOrigins = []string{CorsOriginMatchAll}
				return nil
			}
		}

		ch.allowedOrigins = origins
		return nil
	}
}

// AllowedOriginValidator sets a function for evaluating allowed origins in CORS requests, represented by the
// 'Allow-Access-Control-Origin' HTTP header.
func AllowedOriginValidator(fn OriginValidator) CORSOption {
	return func(ch *cors) error {
		ch.allowedOriginValidator = fn
		return nil
	}
}

// OptionStatusCode sets a custom status code on the OPTIONS requests.
// Default behaviour sets it to 200 to reflect best practices. This is option is not mandatory
// and can be used if you need a custom status code (i.e 204).
//
// More informations on the spec:
// https://fetch.spec.whatwg.org/#cors-preflight-fetch
func OptionStatusCode(code int) CORSOption {
	return func(ch *cors) error {
		ch.optionStatusCode = code
		return nil
	}
}

// ExposedHeaders can be used to specify headers that are available
// and will not be stripped out by the user-agent.
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

// MaxAge determines the maximum age (in seconds) between preflight requests. A
// maximum of 10 minutes is allowed. An age above this value will default to 10
// minutes.
func MaxAge(age int) CORSOption {
	return func(ch *cors) error {
		// Maximum of 10 minutes.
		if age > 600 {
			age = 600
		}

		ch.maxAge = age
		return nil
	}
}

// IgnoreOptions causes the CORS middleware to ignore OPTIONS requests, instead
// passing them through to the next handler. This is useful when your application
// or framework has a pre-existing mechanism for responding to OPTIONS requests.
func IgnoreOptions() CORSOption {
	return func(ch *cors) error {
		ch.ignoreOptions = true
		return nil
	}
}

// AllowCredentials can be used to specify that the user agent may pass
// authentication details along with the request.
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
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}

	return false
}
