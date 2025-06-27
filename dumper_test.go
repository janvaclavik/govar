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

func TestDumpArraysAndSlices(t *testing.T) {
	tests := []struct {
		name         string
		input        any
		wantContains string
	}{
		{"empty array", [0]int{}, "[0]int => |0| []"},
		{"array of ints", [3]int{1, 2, 3}, "[3]int => |3| [0 => 1, 1 => 2, 2 => 3]"},
		{"empty slice", []string{}, "[]string => |0| []"},
		{"slice of strings", []string{"a", "b"}, `|2| [0 => |R:1| "a", 1 => |R:1| "b"]`},
		{"nil slice", []int(nil), "<nil>"},
		{"slice of structs", []struct{ A int }{{1}, {2}},
			`[]struct { A int } => |2| [
   0 struct { A int } => {⯀ A int => 1}
   1 struct { A int } => {⯀ A int => 2}
]`},
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

func TestDumpStructs(t *testing.T) {
	type Basic struct {
		Name string
		Age  int
	}

	type Nested struct {
		ID    int
		Inner Basic
	}

	type Anonymous struct {
		string
		int
	}

	b := Basic{"Alice", 30}
	n := Nested{1, b}
	a := Anonymous{"hidden", 42}
	ptr := &Basic{"Bob", 40}

	tests := []struct {
		name         string
		input        any
		wantContains string
	}{
		{
			name:         "basic struct",
			input:        b,
			wantContains: `govar.Basic => {⯀ Name string => |R:5| "Alice", ⯀ Age int => 30}`,
		},
		{
			name:         "pointer to struct",
			input:        ptr,
			wantContains: `*govar.Basic => {⯀ Name string => |R:3| "Bob", ⯀ Age int => 40}`,
		},
		{
			name:  "nested struct",
			input: n,
			wantContains: `govar.Nested => {
   ⯀ ID     int         => 1
   ⯀ Inner  govar.Basic => {⯀ Name string => |R:5| "Alice", ⯀ Age int => 30}
}`,
		},
		{
			name:         "anonymous fields struct",
			input:        a,
			wantContains: `govar.Anonymous => {🞏 string string => |R:6| "hidden", 🞏 int int => 42}`,
		},
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
