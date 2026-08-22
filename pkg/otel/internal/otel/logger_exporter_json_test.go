package otel

import (
	"bytes"
	"encoding/json"
	"encoding/json/jsontext"
	"testing"

	"go.opentelemetry.io/otel/attribute"
)

// writeJSONValueForTest 将 value 编码为独立 JSON 值（去掉 jsontext 的尾随换行）
func writeJSONValueForTest(t *testing.T, value any) json.RawMessage {
	t.Helper()

	b := &bytes.Buffer{}
	enc := jsontext.NewEncoder(b)

	if err := writeJSONValue(enc, value); err != nil {
		t.Fatalf("writeJSONValue(%#v) 失败: %v", value, err)
	}

	return json.RawMessage(bytes.TrimSpace(b.Bytes()))
}

func TestWriteJSONValue(t *testing.T) {
	t.Run("string 应编码为字符串", func(t *testing.T) {
		got := writeJSONValueForTest(t, "trace.id")
		expect := json.RawMessage(`"trace.id"`)
		if string(got) != string(expect) {
			t.Fatalf("got %s, expect %s", got, expect)
		}
	})

	t.Run("int 应编码为数字", func(t *testing.T) {
		got := writeJSONValueForTest(t, 42)
		expect := json.RawMessage(`42`)
		if string(got) != string(expect) {
			t.Fatalf("got %s, expect %s", got, expect)
		}
	})

	t.Run("bool 应编码为布尔", func(t *testing.T) {
		got := writeJSONValueForTest(t, true)
		expect := json.RawMessage(`true`)
		if string(got) != string(expect) {
			t.Fatalf("got %s, expect %s", got, expect)
		}
	})
}

func TestKeyValueTo(t *testing.T) {
	b := &bytes.Buffer{}
	enc := jsontext.NewEncoder(b)

	if err := enc.WriteToken(jsontext.BeginObject); err != nil {
		t.Fatal(err)
	}

	e := &jsonExporter{}

	if err := e.keyValueTo(enc, "time", "2026-08-22T14:46:02Z"); err != nil {
		t.Fatal(err)
	}
	if err := e.keyValueTo(enc, attribute.Key("traceID"), "ef343ab0d21947cd"); err != nil {
		t.Fatal(err)
	}
	if err := e.keyValueTo(enc, "count", 1); err != nil {
		t.Fatal(err)
	}

	if err := enc.WriteToken(jsontext.EndObject); err != nil {
		t.Fatal(err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(b.Bytes(), &decoded); err != nil {
		t.Fatalf("输出不是合法 JSON: %v\noutput: %s", err, b.Bytes())
	}

	if decoded["time"] != "2026-08-22T14:46:02Z" {
		t.Fatalf("time 字段不一致: %#v", decoded["time"])
	}
	if decoded["traceID"] != "ef343ab0d21947cd" {
		t.Fatalf("traceID 字段不一致: %#v", decoded["traceID"])
	}
	if decoded["count"] != float64(1) {
		t.Fatalf("count 字段不一致: %#v", decoded["count"])
	}
}
