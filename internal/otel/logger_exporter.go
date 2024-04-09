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
	"golang.org/x/term"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
)

func findTTY() (*os.File, bool) {
	// some of these may be redirected
	for _, f := range []*os.File{os.Stderr, os.Stdout, os.Stdin} {
		if term.IsTerminal(int(f.Fd())) {
			return f, true
		}
	}
	return nil, false
}

var isTTY = false

func init() {
	_, tty := findTTY()
	if tty {
		isTTY = true
	}
	if os.Getenv("TTY") == "0" {
		isTTY = false
	}
}

func SlogExporter() sdklog.Exporter {
	if isTTY {
		return &prettyExporter{}
	}

	return &jsonExporter{}
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

	b := bytes.NewBuffer(nil)

	written := map[string]bool{}

	for attr := range r.WalkAttributes {
		if written[attr.Key] {
			continue
		}
		written[attr.Key] = true

		_, _ = fmt.Fprint(b, " ")
		_, _ = fmt.Fprint(b, attr.Key)
		_, _ = fmt.Fprint(b, "=")
		v, err := marshal(LogValue(attr.Value))
		if err != nil {
			return err
		}
		_, _ = fmt.Fprint(b, string(v))
	}

	if name := r.InstrumentationScope().Name; name != "" {
		_, _ = fmt.Fprintf(b, " spanName=%s", name)
	}

	if b.Len() > 0 {
		_, _ = fmt.Fprint(w, color.WhiteString(b.String()))
	}

	_, _ = fmt.Fprintln(w)

	_, err := io.Copy(f, w)
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
