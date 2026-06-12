package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/felixge/httpsnoop"
)

// HandlerLevel 根据指定压缩级别对 HTTP 响应进行压缩，
// 仅对通过 'Accept-Encoding' 头部声明支持的客户端生效。
//
// 压缩级别应为 gzip.DefaultCompression、gzip.NoCompression，
// 或介于 gzip.BestSpeed 与 gzip.BestCompression 之间的任意整数值。
// 若传入无效级别，则默认使用 gzip.DefaultCompression。
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
			// 检测应使用的编码
			var encoding string
			for curEnc := range strings.SplitSeq(r.Header.Get(acceptEncoding), ",") {
				curEnc = strings.TrimSpace(curEnc)
				if curEnc == brEncoding || curEnc == gzipEncoding {
					encoding = curEnc
				}
			}

			// 未识别到支持的编码，将请求透传给处理器并返回
			if encoding == "" {
				h.ServeHTTP(w, r)
				return
			}

			if r.Header.Get("Upgrade") != "" {
				h.ServeHTTP(w, r)
				return
			}

			// 始终将 Accept-Encoding 加入 Vary，防止中间缓存污染
			w.Header().Add("Vary", acceptEncoding)

			// 用选定编码的 writer 包装 ResponseWriter
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
	// 若压缩器支持，刷新压缩数据。
	if f, ok := w.compressor.(flusher); ok {
		_ = f.Flush()
	}
	// 刷新 HTTP 响应。
	if f, ok := w.w.(http.Flusher); ok {
		f.Flush()
	}
}
