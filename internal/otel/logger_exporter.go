package otel

import (
	"bytes"
	"context"
	"cuelang.org/go/cue/cuecontext"
	cueformat "cuelang.org/go/cue/format"
	"cuelang.org/go/encoding/gocode/gocodec"
	"fmt"
	"github.com/fatih/color"
	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
)

func SlogExporter(format LogFormat) sdklog.Exporter {
	switch format {
	case LogFormatJSON:
		return &jsonExporter{}
	default:
		return &prettyExporter{}
	}
}

type prettyExporter struct {
	done atomic.Bool
	wg   sync.WaitGroup
}

func (e *prettyExporter) ForceFlush(ctx context.Context) error {
	return nil
}

func (e *prettyExporter) Shutdown(ctx context.Context) error {
	return nil
}

func (e *prettyExporter) Export(ctx context.Context, records []sdklog.Record) error {
	for _, r := range records {
		if r.Severity() >= log.SeverityWarn1 {
			if err := e.print(os.Stderr, r); err != nil {
				return err
			}
			continue
		}
		if err := e.print(os.Stdout, r); err != nil {
			return err
		}
	}

	return nil
}

func (e *prettyExporter) print(f io.Writer, r sdklog.Record) error {
	w := bytes.NewBuffer(nil)

	prefix := color.CyanString("%s:", r.SpanID().String())
	_, _ = fmt.Fprint(w, prefix)
	_, _ = fmt.Fprint(w, " ")
	_, _ = fmt.Fprint(w, strings.ToUpper(severityText(r))[0:4])
	_, _ = fmt.Fprint(w, " ")
	_, _ = fmt.Fprint(w, color.WhiteString(r.Timestamp().Format("15:04:05")))
	_, _ = fmt.Fprint(w, " ")

	_, _ = fmt.Fprint(w, r.Body().AsString())

	attrs := map[string]any{}

	if name := r.InstrumentationScope().Name; name != "" {
		attrs["spanName"] = name
	}

	for attr := range r.WalkAttributes {
		attrs[attr.Key] = LogValue(attr.Value)
	}

	data, err := marshal(attrs)
	if err != nil {
		panic(err)
	}

	if len(data) > 0 {
		_, _ = fmt.Fprint(w, "\t")
		_, _ = fmt.Fprint(w, color.WhiteString(string(data)))
	}

	_, _ = fmt.Fprintln(w)
	_, err = io.Copy(f, w)
	return err
}

func marshal(v any) ([]byte, error) {
	codec := gocodec.New(cuecontext.New(), nil)
	val, err := codec.Decode(v)
	if err != nil {
		return nil, err
	}
	return cueformat.Node(val.Syntax(), cueformat.Simplify())
}
