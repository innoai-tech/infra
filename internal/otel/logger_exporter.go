package otel

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	cueformat "cuelang.org/go/cue/format"
	"github.com/fatih/color"
	"github.com/go-json-experiment/json"
	"github.com/go-json-experiment/json/jsontext"
	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
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
		attrs["trace.span.name"] = name
	}

	for attr := range r.WalkAttributes {
		attrs[attr.Key] = LogValue(attr.Value)
	}

	data, err := marshal(attrs)
	if err != nil {
		panic(fmt.Errorf("failed to marshal: %+v", attrs))
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
	raw, err := json.Marshal(v,
		json.Deterministic(true),
		jsontext.AllowInvalidUTF8(true),
		jsontext.WithIndent("  "),
	)
	if err != nil {
		return nil, err
	}
	// skip [] {}
	if len(raw) == 2 && (raw[0] == '[' || raw[0] == '{') {
		return nil, nil
	}
	return cueformat.Source(raw, cueformat.Simplify())
}
