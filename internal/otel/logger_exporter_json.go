package otel

import (
	"bytes"
	"context"
	"encoding/base64"
	"github.com/go-json-experiment/json/jsontext"
	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"io"
	"os"
	"time"
)

type jsonExporter struct {
}

func (e *jsonExporter) Export(ctx context.Context, records []sdklog.Record) error {
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

func (e *jsonExporter) print(w io.Writer, r sdklog.Record) error {
	b := bytes.NewBuffer(nil)
	enc := jsontext.NewEncoder(b)

	if err := enc.WriteToken(jsontext.ObjectStart); err != nil {
		return err
	}

	if err := e.keyValueTo(enc, "time", r.Timestamp().Format(time.RFC3339)); err != nil {
		return err
	}

	if err := e.keyValueTo(enc, "level", severityText(r)); err != nil {
		return err
	}

	if err := e.keyValueTo(enc, "msg", r.Body().AsString()); err != nil {
		return err
	}

	written := map[string]bool{}

	for attr := range r.WalkAttributes {
		if written[attr.Key] {
			continue
		}
		written[attr.Key] = true

		if err := e.keyValueTo(enc, attr.Key, LogValue(attr.Value)); err != nil {
			return err
		}
	}

	res := r.Resource()
	for _, attr := range res.Attributes() {
		switch attr.Key {
		case "service.name", "service.version":
			if err := e.keyValueTo(enc, string(attr.Key), attr.Value.AsInterface()); err != nil {
				return err
			}
		}
	}

	if err := e.keyValueTo(enc, "spanName", r.InstrumentationScope().Name); err != nil {
		return err
	}

	if err := e.keyValueTo(enc, "traceID", r.TraceID().String()); err != nil {
		return err
	}

	if err := e.keyValueTo(enc, "spanID", r.SpanID().String()); err != nil {
		return err
	}

	if err := enc.WriteToken(jsontext.ObjectEnd); err != nil {
		return err
	}

	_, err := io.Copy(w, b)
	return err
}

func (e *jsonExporter) keyValueTo(enc *jsontext.Encoder, key string, value any) error {
	if err := writeJSONValue(enc, key); err != nil {
		return err
	}
	return writeJSONValue(enc, value)
}

func (e jsonExporter) Shutdown(ctx context.Context) error {
	return nil
}

func (e jsonExporter) ForceFlush(ctx context.Context) error {
	return nil
}

func writeJSONValue(enc *jsontext.Encoder, value any) error {
	switch x := value.(type) {
	case string:
		return enc.WriteToken(jsontext.String(x))
	case int:
		return enc.WriteToken(jsontext.Int(int64(x)))
	case int8:
		return enc.WriteToken(jsontext.Int(int64(x)))
	case int16:
		return enc.WriteToken(jsontext.Int(int64(x)))
	case int32:
		return enc.WriteToken(jsontext.Int(int64(x)))
	case int64:
		return enc.WriteToken(jsontext.Int(int64(x)))
	case uint:
		return enc.WriteToken(jsontext.Uint(uint64(x)))
	case uint8:
		return enc.WriteToken(jsontext.Uint(uint64(x)))
	case uint16:
		return enc.WriteToken(jsontext.Uint(uint64(x)))
	case uint32:
		return enc.WriteToken(jsontext.Uint(uint64(x)))
	case uint64:
		return enc.WriteToken(jsontext.Uint(uint64(x)))
	case float32:
		return enc.WriteToken(jsontext.Float(float64(x)))
	case float64:
		return enc.WriteToken(jsontext.Float(x))
	case bool:
		return enc.WriteToken(jsontext.Bool(x))
	case map[string]any:
		if err := enc.WriteToken(jsontext.ObjectStart); err != nil {
			return err
		}
		for key, value := range x {
			if err := writeJSONValue(enc, key); err != nil {
				return err
			}
			if err := writeJSONValue(enc, value); err != nil {
				return err
			}
		}
		return enc.WriteToken(jsontext.ObjectEnd)
	case []any:
		if err := enc.WriteToken(jsontext.ArrayStart); err != nil {
			return err
		}
		for i := range x {
			if err := writeJSONValue(enc, x[i]); err != nil {
				return err
			}
		}
		return enc.WriteToken(jsontext.ArrayEnd)
	case []byte:
		return enc.WriteToken(jsontext.String(base64.StdEncoding.EncodeToString(x)))
	}
	return nil
}

func severityText(r sdklog.Record) string {
	if txt := r.SeverityText(); txt != "" {
		return txt
	}

	s := r.Severity()

	if s >= log.SeverityFatal {
		return "fatal"
	}
	if s >= log.SeverityError {
		return "error"
	}
	if s >= log.SeverityWarn {
		return "warn"
	}
	if s >= log.SeverityInfo {
		return "info"
	}
	if s >= log.SeverityDebug {
		return "debug"
	}
	return "trace"
}
