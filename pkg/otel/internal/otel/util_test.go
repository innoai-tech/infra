package otel

import (
	"fmt"
	"log/slog"
	"reflect"
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"

	testingv2 "github.com/octohelm/x/testing/v2"
)

type stringer string

func (s stringer) String() string { return "stringer:" + string(s) }

type unwrapper struct{ v any }

func (u unwrapper) Unwrap() any { return u.v }

// keyValuesEqual 逐项比较 attribute.KeyValue 列表。
// 不用 testingv2.Equal：其失败时经 go-cmp 生成 diff 会因 attribute.Value 的未导出字段 panic。
func keyValuesEqual(expects ...attribute.KeyValue) func([]attribute.KeyValue) error {
	return func(got []attribute.KeyValue) error {
		if len(got) != len(expects) {
			return fmt.Errorf("数量不一致: got %d, expect %d\n  got:    %v\n  expect: %v", len(got), len(expects), got, expects)
		}
		for i := range expects {
			if got[i].Key != expects[i].Key || !reflect.DeepEqual(got[i].Value, expects[i].Value) {
				return fmt.Errorf("kv[%d] 不一致:\n  got:    %v\n  expect: %v", i, got[i], expects[i])
			}
		}
		return nil
	}
}

// valueEqual 比较 attribute.Value，理由同 keyValuesEqual。
func valueEqual(expect attribute.Value) func(attribute.Value) error {
	return func(got attribute.Value) error {
		if !reflect.DeepEqual(got, expect) {
			return fmt.Errorf("值不一致:\n  got:    %v\n  expect: %v", got, expect)
		}
		return nil
	}
}

func TestNormalizeKeyValues(t *testing.T) {
	testingv2.Then(t, "从 []slog.Attr 归一化",
		testingv2.Expect(
			normalizeKeyValues([]any{
				[]slog.Attr{slog.String("a", "1"), slog.Int("n", 2)},
			}),
			testingv2.Be(keyValuesEqual(
				attribute.String("a", "1"),
				attribute.Int("n", 2),
			)),
		),
	)

	testingv2.Then(t, "从 slog.Attr 归一化",
		testingv2.Expect(
			normalizeKeyValues([]any{slog.Bool("ok", true)}),
			testingv2.Be(keyValuesEqual(
				attribute.Bool("ok", true),
			)),
		),
	)

	testingv2.Then(t, "从 slog.Any 归一化（值走 LogAnyValue）",
		testingv2.Expect(
			normalizeKeyValues([]any{slog.Any("d", 2*time.Second)}),
			testingv2.Be(keyValuesEqual(
				attribute.String("d", "2s"),
			)),
		),
	)

	testingv2.Then(t, "从 []attribute.KeyValue 归一化",
		testingv2.Expect(
			normalizeKeyValues([]any{
				[]attribute.KeyValue{attribute.String("a", "1")},
			}),
			testingv2.Be(keyValuesEqual(
				attribute.String("a", "1"),
			)),
		),
	)

	testingv2.Then(t, "从 attribute.KeyValue 归一化",
		testingv2.Expect(
			normalizeKeyValues([]any{attribute.Float64("f", 1.5)}),
			testingv2.Be(keyValuesEqual(
				attribute.Float64("f", 1.5),
			)),
		),
	)

	testingv2.Then(t, "从 key/value 对归一化",
		testingv2.Expect(
			normalizeKeyValues([]any{"k", "v", "n", 1}),
			testingv2.Be(keyValuesEqual(
				attribute.String("k", "v"),
				attribute.Int("n", 1),
			)),
		),
	)

	testingv2.Then(t, "混合输入归一化",
		testingv2.Expect(
			normalizeKeyValues([]any{
				slog.String("s", "x"),
				attribute.Int("i", 1),
				"t", true,
			}),
			testingv2.Be(keyValuesEqual(
				attribute.String("s", "x"),
				attribute.Int("i", 1),
				attribute.Bool("t", true),
			)),
		),
	)
}

func TestNormalizeKeyValuesPanic(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for unsupported type")
		}
	}()

	normalizeKeyValues([]any{42})
}

func TestLogValue(t *testing.T) {
	testingv2.Then(t, "BOOL",
		testingv2.Expect(LogValue(attribute.BoolValue(true)), testingv2.Equal[any](true)),
	)

	testingv2.Then(t, "BOOLSLICE",
		testingv2.Expect(
			LogValue(attribute.BoolSliceValue([]bool{true, false})),
			testingv2.Equal[any]([]bool{true, false}),
		),
	)

	testingv2.Then(t, "INT64",
		testingv2.Expect(LogValue(attribute.Int64Value(42)), testingv2.Equal[any](int64(42))),
	)

	testingv2.Then(t, "INT64SLICE",
		testingv2.Expect(
			LogValue(attribute.Int64SliceValue([]int64{1, 2})),
			testingv2.Equal[any]([]int64{1, 2}),
		),
	)

	testingv2.Then(t, "FLOAT64",
		testingv2.Expect(LogValue(attribute.Float64Value(1.5)), testingv2.Equal[any](1.5)),
	)

	testingv2.Then(t, "FLOAT64SLICE",
		testingv2.Expect(
			LogValue(attribute.Float64SliceValue([]float64{1.5, 2.5})),
			testingv2.Equal[any]([]float64{1.5, 2.5}),
		),
	)

	testingv2.Then(t, "STRING",
		testingv2.Expect(LogValue(attribute.StringValue("hello")), testingv2.Equal[any]("hello")),
	)

	testingv2.Then(t, "STRINGSLICE",
		testingv2.Expect(
			LogValue(attribute.StringSliceValue([]string{"a", "b"})),
			testingv2.Equal[any]([]string{"a", "b"}),
		),
	)

	testingv2.Then(t, "BYTESLICE",
		testingv2.Expect(
			LogValue(attribute.ByteSliceValue([]byte{1, 2})),
			testingv2.Equal[any]([]byte{1, 2}),
		),
	)

	testingv2.Then(t, "SLICE",
		testingv2.Expect(
			LogValue(attribute.SliceValue(attribute.Int64Value(1), attribute.StringValue("a"))),
			testingv2.Equal[any]([]any{int64(1), "a"}),
		),
	)

	testingv2.Then(t, "MAP",
		testingv2.Expect(
			LogValue(attribute.MapValue(
				attribute.String("k", "v"),
				attribute.Int64("n", 1),
			)),
			testingv2.Equal[any](map[string]any{"k": "v", "n": int64(1)}),
		),
	)

	testingv2.Then(t, "EMPTY",
		testingv2.Expect(LogValue(attribute.Value{}), testingv2.Equal[any](nil)),
	)
}

func TestLogAnyValue(t *testing.T) {
	now := time.Date(2026, 8, 19, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		name   string
		value  any
		expect attribute.Value
	}{
		{"time.Time", now, attribute.StringValue("2026-08-19 12:00:00 +0000 UTC")},
		{"time.Duration", 2 * time.Second, attribute.StringValue("2s")},
		{"fmt.Stringer", stringer("x"), attribute.StringValue("stringer:x")},

		{"[]byte", []byte{1, 2}, attribute.ByteSliceValue([]byte{1, 2})},
		{"[]uint8 同 []byte", []uint8{1, 2}, attribute.ByteSliceValue([]byte{1, 2})},

		{"string", "hello", attribute.StringValue("hello")},

		{"uint", uint(1), attribute.Int64Value(1)},
		{"uint8", uint8(1), attribute.Int64Value(1)},
		{"uint16", uint16(1), attribute.Int64Value(1)},
		{"uint32", uint32(1), attribute.Int64Value(1)},
		{"int", 1, attribute.Int64Value(1)},
		{"int8", int8(1), attribute.Int64Value(1)},
		{"int16", int16(1), attribute.Int64Value(1)},
		{"int32", int32(1), attribute.Int64Value(1)},
		{"int64", int64(1), attribute.Int64Value(1)},

		{"float32", float32(1.5), attribute.Float64Value(1.5)},
		{"float64", 1.5, attribute.Float64Value(1.5)},
		{"bool", true, attribute.BoolValue(true)},

		{"[]string", []string{"a", "b"}, attribute.StringSliceValue([]string{"a", "b"})},
		{"[]bool", []bool{true, false}, attribute.BoolSliceValue([]bool{true, false})},
		{"[]int", []int{1, 2}, attribute.Int64SliceValue([]int64{1, 2})},
		{"[]int8", []int8{1, 2}, attribute.Int64SliceValue([]int64{1, 2})},
		{"[]int16", []int16{1, 2}, attribute.Int64SliceValue([]int64{1, 2})},
		{"[]int32", []int32{1, 2}, attribute.Int64SliceValue([]int64{1, 2})},
		{"[]int64", []int64{1, 2}, attribute.Int64SliceValue([]int64{1, 2})},
		{"[]uint", []uint{1, 2}, attribute.Int64SliceValue([]int64{1, 2})},
		{"[]uint16", []uint16{1, 2}, attribute.Int64SliceValue([]int64{1, 2})},
		{"[]uint32", []uint32{1, 2}, attribute.Int64SliceValue([]int64{1, 2})},
		{"[]uint64", []uint64{1, 2}, attribute.Int64SliceValue([]int64{1, 2})},
		{"[]float32", []float32{1.5, 2.5}, attribute.Float64SliceValue([]float64{1.5, 2.5})},
		{"[]float64", []float64{1.5, 2.5}, attribute.Float64SliceValue([]float64{1.5, 2.5})},

		{"[]any", []any{1, "a"}, attribute.SliceValue(
			attribute.Int64Value(1),
			attribute.StringValue("a"),
		)},
		{"[]any 嵌套", []any{[]any{1, 2}, map[string]any{"k": "v"}}, attribute.SliceValue(
			attribute.SliceValue(attribute.Int64Value(1), attribute.Int64Value(2)),
			attribute.MapValue(attribute.String("k", "v")),
		)},

		{"map[string]any", map[string]any{"a": 1}, attribute.MapValue(attribute.Int64("a", 1))},

		{"Unwrap", unwrapper{v: 42}, attribute.Int64Value(42)},
		{"未知类型回退为字符串", struct{ Name string }{Name: "x"}, attribute.StringValue("{x}")},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			testingv2.Then(t, "转换为 attribute.Value",
				testingv2.Expect(LogAnyValue(c.value), testingv2.Be(valueEqual(c.expect))),
			)
		})
	}
}
