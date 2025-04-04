package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/felixge/httpsnoop"
)

// HandlerLevel gzip compresses HTTP responses with specified compression level
// for clients that support it via the 'Accept-Encoding' header.
//
// The compression level should be gzip.DefaultCompression, gzip.NoCompression,
// or any integer value between gzip.BestSpeed and gzip.BestCompression inclusive.
// gzip.DefaultCompression is used in case of invalid compression level.
func HandlerLevel(level int) func(h http.Handler) http.Handler {
	if level < gzip.DefaultCompression || level > gzip.BestCompression {
		level = gzip.DefaultCompression
	}

	const (
		brEncoding   = "br"
		gzipEncoding = "gzip"
	)

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// detect what encoding to use
			var encoding string
			for _, curEnc := range strings.Split(r.Header.Get(acceptEncoding), ",") {
				curEnc = strings.TrimSpace(curEnc)
				if curEnc == brEncoding || curEnc == gzipEncoding {
					encoding = curEnc
				}
			}

			// if we weren't able to identify an encoding we're familiar with, pass on the
			// request to the handler and return
			if encoding == "" {
				h.ServeHTTP(w, r)
				return
			}

			if r.Header.Get("Upgrade") != "" {
				h.ServeHTTP(w, r)
				return
			}

			// always add Accept-Encoding to Vary to prevent intermediate caches corruption
			w.Header().Add("Vary", acceptEncoding)

			// wrap the ResponseWriter with the writer for the chosen encoding
			var encWriter io.WriteCloser
			switch encoding {
			case gzipEncoding:
				encWriter, _ = gzip.NewWriterLevel(w, level)
			case brEncoding:
				encWriter = brotli.NewWriterLevel(w, level)
			default:
				encoding = ""
			}

			defer func() {
				_ = encWriter.Close()
			}()

			w.Header().Set("Content-Encoding", encoding)
			r.Header.Del(acceptEncoding)

			cw := &compressResponseWriter{
				w:          w,
				compressor: encWriter,
			}

			h.ServeHTTP(httpsnoop.Wrap(w, httpsnoop.Hooks{
				Write: func(httpsnoop.WriteFunc) httpsnoop.WriteFunc {
					return cw.Write
				},
				WriteHeader: func(httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
					return cw.WriteHeader
				},
				Flush: func(httpsnoop.FlushFunc) httpsnoop.FlushFunc {
					return cw.Flush
				},
				ReadFrom: func(rff httpsnoop.ReadFromFunc) httpsnoop.ReadFromFunc {
					return cw.ReadFrom
				},
			}), r)
		})
	}
}

const acceptEncoding string = "Accept-Encoding"

type compressResponseWriter struct {
	compressor io.Writer
	w          http.ResponseWriter
}

func (cw *compressResponseWriter) WriteHeader(c int) {
	cw.w.Header().Del("Content-Length")
	cw.w.WriteHeader(c)
}

func (cw *compressResponseWriter) Write(b []byte) (int, error) {
	h := cw.w.Header()
	if h.Get("Content-Type") == "" {
		h.Set("Content-Type", http.DetectContentType(b))
	}
	h.Del("Content-Length")

	return cw.compressor.Write(b)
}

func (cw *compressResponseWriter) ReadFrom(r io.Reader) (int64, error) {
	return io.Copy(cw.compressor, r)
}

type flusher interface {
	Flush() error
}

func (w *compressResponseWriter) Flush() {
	// Flush compressed data if compressor supports it.
	if f, ok := w.compressor.(flusher); ok {
		_ = f.Flush()
	}
	// Flush HTTP response.
	if f, ok := w.w.(http.Flusher); ok {
		f.Flush()
	}
}
