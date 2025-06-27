package govar

import (
	"strings"
	"testing"
)

func TestDumpBasicTypes(t *testing.T) {
	tests := []struct {
		name         string
		input        any
		wantContains string
	}{
		{"int", 42, "int => 42"},
		{"int8", int8(8), "int8 => 8"},
		{"int16", int16(16), "int16 => 16"},
		{"int32", int32(32), "int32 => 32"},
		{"int64", int64(64), "int64 => 64"},
		{"uint", uint(42), "uint => 42"},
		{"uint8", uint8(8), "uint8 => 8"},
		{"uint16", uint16(16), "uint16 => 16"},
		{"uint32", uint32(32), "uint32 => 32"},
		{"uint64", uint64(64), "uint64 => 64"},
		{"uintptr", uintptr(0x1234), "uintptr => 4660"},
		{"float32", float32(3.14), "float32 => 3.14"},
		{"float64", float64(6.28), "float64 => 6.28"},
		{"complex64", complex64(1 + 2i), "complex64 => (1+2i)"},
		{"complex128", complex128(3 + 4i), "complex128 => (3+4i)"},
		{"boolTrue", true, "bool => true"},
		{"boolFalse", false, "bool => false"},
		{"string", "test", `string => |R:4| "test"`},
		{"byte", byte('A'), "uint8 => 65"},
		{"rune", rune('♥'), "int32 => 9829"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := SdumpNoColors(tt.input)
			if !strings.Contains(out, tt.wantContains) {
				t.Errorf("Dump %s: got:\n%s\nwant contains:\n%s", tt.name, out, tt.wantContains)
			}
		})
	}
}

func TestDumpPointersAndNil(t *testing.T) {
	val := 100
	str := "hello"

	tests := []struct {
		name         string
		input        any
		wantContains string
	}{
		{"nil", nil, "invalid => <invalid>"},
		{"nil interface", any(nil), "invalid => <invalid>"},
		{"nil pointer", (*int)(nil), "*int => <nil>"},
		{"pointer to int", &val, "*int => 100"},
		{"pointer to string", &str, `*string => |R:5| "hello"`},
		{"pointer to pointer", func() any {
			p := &val
			return &p
		}(), "**int => 100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := SdumpNoColors(tt.input)
			if !strings.Contains(out, tt.wantContains) {
				t.Errorf("Dump %s: got:\n%s\nwant contains:\n%s", tt.name, out, tt.wantContains)
			}
		})
	}
}

// func TestDumpStruct(t *testing.T) {
// 	type person struct {
// 		Name string
// 		Age  int
// 	}

// 	p := person{Name: "Alice", Age: 30}
// 	var buf bytes.Buffer
// 	Dump(&buf, p)

// 	out := buf.String()
// 	if !strings.Contains(out, "Name") || !strings.Contains(out, "\"Alice\"") || !strings.Contains(out, "30") {
// 		t.Errorf("Dump(struct) output = %q; want output containing struct fields", out)
// 	}
// }

// func TestDumpPointer(t *testing.T) {
// 	x := 123
// 	var buf bytes.Buffer
// 	Dump(&buf, &x)

// 	out := buf.String()
// 	if !strings.Contains(out, "int") || !strings.Contains(out, "123") {
// 		t.Errorf("Dump(*int) output = %q; want output containing 'int' and '123'", out)
// 	}
// }

// func TestDumpNil(t *testing.T) {
// 	var buf bytes.Buffer
// 	var ptr *int
// 	Dump(&buf, ptr)

// 	if !strings.Contains(buf.String(), "nil") {
// 		t.Errorf("Dump(nil pointer) output = %q; want output containing 'nil'", buf.String())
// 	}
// }

// func TestDumpSlice(t *testing.T) {
// 	s := []string{"foo", "bar"}
// 	var buf bytes.Buffer
// 	Dump(&buf, s)

// 	out := buf.String()
// 	if !strings.Contains(out, "foo") || !strings.Contains(out, "bar") {
// 		t.Errorf("Dump(slice) output = %q; want output containing slice elements", out)
// 	}
// }
