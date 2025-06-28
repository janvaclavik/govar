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
		{"nil", nil, "unknown => <nil>"},
		{"nil interface", any(nil), "unknown => <nil>"},
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

func TestDumpMaps(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	mixedMap := map[string]any{
		"number":  1,
		"boolean": true,
		"string":  "foo",
	}

	structMap := map[string]Person{
		"alice": {"Alice", 30},
		"bob":   {"Bob", 40},
	}

	var nilMap map[string]int
	emptyMap := map[int]string{}
	ptrMap := &map[string]bool{"ok": true}

	tests := []struct {
		name         string
		input        any
		wantContains string
	}{
		{
			name:         "map string to int",
			input:        map[string]int{"a": 1, "b": 2},
			wantContains: `map[string]int => |2| ["a"  => 1, "b"  => 2]`,
		},
		{
			name:  "map with mixed value types",
			input: mixedMap,
			wantContains: `map[string]any => |3| [
   "boolean"  ⧉ any(bool)   => true
   "number"   ⧉ any(int)    => 1
   "string"   ⧉ any(string) => |R:3| "foo"
]`,
		},
		{
			name:  "map with struct values",
			input: structMap,
			wantContains: `ap[string]govar.Person => |2| [
   "alice"  govar.Person => {⯀ Name string => |R:5| "Alice", ⯀ Age int => 30}
   "bob"    govar.Person => {⯀ Name string => |R:3| "Bob", ⯀ Age int => 40}
]`,
		},
		{
			name:         "nil map",
			input:        nilMap,
			wantContains: `map[string]int => <nil>`,
		},
		{
			name:         "empty map",
			input:        emptyMap,
			wantContains: `map[int]string => |0| []`,
		},
		{
			name:         "pointer to map",
			input:        ptrMap,
			wantContains: `*map[string]bool => |1| ["ok"  => true]`,
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

func TestDumpInterfaces(t *testing.T) {
	type MyInterface interface {
		Dummy()
	}

	var nilInterface MyInterface
	var emptyInterface any
	var ifaceWithInt any = 123
	var ifaceWithString any = "hello"
	var ifaceWithStruct any = struct{ A int }{A: 7}
	var ifaceWithPtr any = &struct{ B string }{B: "world"}
	var ifaceWithIface any = any(123.45)

	tests := []struct {
		name         string
		input        any
		wantContains string
	}{
		{
			name:         "nil interface",
			input:        nilInterface,
			wantContains: `unknown => <nil>`,
		},
		{
			name:         "empty interface nil",
			input:        emptyInterface,
			wantContains: `unknown => <nil>`,
		},
		{
			name:         "interface with int",
			input:        ifaceWithInt,
			wantContains: `int => 123`,
		},
		{
			name:         "interface with string",
			input:        ifaceWithString,
			wantContains: `string => |R:5| "hello"`,
		},
		{
			name:         "interface with struct",
			input:        ifaceWithStruct,
			wantContains: `struct { A int } => {⯀ A int => 7}`,
		},
		{
			name:         "interface with pointer to struct",
			input:        ifaceWithPtr,
			wantContains: `*struct { B string } => {⯀ B string => |R:5| "world"}`,
		},
		{
			name:         "interface with another interface",
			input:        ifaceWithIface,
			wantContains: `float64 => 123.450000`,
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

func TestDumpChannels(t *testing.T) {
	chInt := make(chan int)
	chString := make(chan string, 5)
	var nilChan chan bool
	chRecvOnly := make(<-chan float64)
	chSendOnly := make(chan<- struct{})

	tests := []struct {
		name         string
		input        any
		wantContains string
	}{
		{
			name:         "nil channel",
			input:        nilChan,
			wantContains: `chan bool => <nil>`,
		},
		{
			name:         "unbuffered int channel",
			input:        chInt,
			wantContains: `chan int => |B:0| ⮁`,
		},
		{
			name:         "buffered string channel",
			input:        chString,
			wantContains: `chan string => |B:5| ⮁`,
		},
		{
			name:         "receive-only channel",
			input:        chRecvOnly,
			wantContains: `<-chan float64 => |B:0| 🢃`,
		},
		{
			name:         "send-only channel",
			input:        chSendOnly,
			wantContains: `chan<- struct {} => |B:0| 🡹`,
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

func TestDumpFunctions(t *testing.T) {
	fnSimple := func() {}
	fnWithArgs := func(a int, b string) error { return nil }
	var nilFunc func()
	fnTyped := func(a float64) bool { return a > 0 }

	tests := []struct {
		name         string
		input        any
		wantContains string
	}{
		{
			name:         "nil function",
			input:        nilFunc,
			wantContains: `func() => <nil>`,
		},
		{
			name:         "anonymous simple func",
			input:        fnSimple,
			wantContains: `func() =>`,
		},
		{
			name:         "func with args and return",
			input:        fnWithArgs,
			wantContains: `func(int, string) error =>`,
		},
		{
			name:         "typed func(float64) bool",
			input:        fnTyped,
			wantContains: `func(float64) bool =>`,
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

func TestDumpNestedValues(t *testing.T) {
	type Inner struct {
		Z int
	}
	type Outer struct {
		A int
		B *Inner
		C []Inner
		D map[string]*Inner
	}

	inner := &Inner{Z: 99}
	outer := Outer{
		A: 1,
		B: inner,
		C: []Inner{{Z: 2}, {Z: 3}},
		D: map[string]*Inner{"key": inner},
	}

	tests := []struct {
		name         string
		input        any
		wantContains []string
	}{
		{
			name:  "Nested structs with slice, map, and pointer",
			input: outer,
			wantContains: []string{
				`govar.Outer => {`,
				`⯀ A  int                     => 1`,
				`⯀ B  *govar.Inner            => &1 {⯀ Z int => 99}`, // Pointer to Inner
				`⯀ C  []govar.Inner           => |2| [`,              // Inner's field
				`0 govar.Inner => {⯀ Z int => 2}`,                    // Slice
				`1 govar.Inner => {⯀ Z int => 3}`,
				`⯀ D  map[string]*govar.Inner => |1| [`, // Map
				`"key" => ↩︎ &1`,                        // Map key
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := SdumpNoColors(tt.input)
			for _, want := range tt.wantContains {
				if !strings.Contains(out, want) {
					t.Errorf("Dump %s: output missing expected fragment:\nwant contains:\n%s\ngot:\n%s", tt.name, want, out)
				}
			}
		})
	}
}

func TestDumpEmbeddedStructs(t *testing.T) {
	type Address struct {
		City string
		Zip  string
	}

	type Person struct {
		Name string
		Age  int
		Address
	}

	type Employee struct {
		Person
		Position string
	}

	tests := []struct {
		name         string
		input        any
		wantContains []string
	}{
		{
			name: "Embedded struct fields visible",
			input: Person{
				Name: "John",
				Age:  40,
				Address: Address{
					City: "New York",
					Zip:  "10001",
				},
			},
			wantContains: []string{
				`govar.Person => {`,
				`⯀ Name     string        => |R:4| "John"`,
				`⯀ Age      int           => 40`,
				`⯀ Address  govar.Address => {⯀ City string => |R:8| "New York", ⯀ Zip string => |R:5| "10001"}`,
			},
		},
		{
			name: "Deeply embedded structs",
			input: Employee{
				Person: Person{
					Name: "Alice",
					Age:  30,
					Address: Address{
						City: "San Francisco",
						Zip:  "94105",
					},
				},
				Position: "Engineer",
			},
			wantContains: []string{
				`⯀ Person    govar.Person => {`,
				`⯀ Name     string        => |R:5| "Alice"`,
				`⯀ Age      int           => 30`,
				`⯀ Address  govar.Address => {⯀ City string => |R:13| "San Francisco", ⯀ Zip string => |R:5| "94105"}`,
				`⯀ Position  string       => |R:8| "Engineer`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := SdumpNoColors(tt.input)
			for _, want := range tt.wantContains {
				if !strings.Contains(out, want) {
					t.Errorf("Dump %s: output missing expected fragment:\nwant contains:\n%s\ngot:\n%s", tt.name, want, out)
				}
			}
		})
	}
}

type Describer interface {
	Describe() string
}

type DescribedThing struct {
	What string
}

func (d DescribedThing) Describe() string {
	return "I am " + d.What
}

func TestDumpStructsWithInterfacesAndAnonymousFields(t *testing.T) {

	type ContainerWithInterface struct {
		ID       int
		Contents Describer
	}

	type AnonymousField struct {
		string
		int
	}

	type ComplexStruct struct {
		AnonymousField // embedded anonymously
		Data           Describer
	}

	tests := []struct {
		name         string
		input        any
		wantContains []string
	}{
		{
			name: "Struct with interface field",
			input: ContainerWithInterface{
				ID: 101,
				Contents: DescribedThing{
					What: "a powerful dumper",
				},
			},
			wantContains: []string{
				`govar.ContainerWithInterface => {`,
				`⯀ ID        int                                     => 101`,
				`⯀ Contents  ⧉ govar.Describer(govar.DescribedThing) => {`,
				`⯀ What      string => |R:17| "a powerful dumper"`,
				`⦿ Describe  func(govar.DescribedThing) string`,
			},
		},
		{
			name: "Struct with anonymous fields and interface",
			input: ComplexStruct{
				AnonymousField: AnonymousField{
					string: "anonymous!",
					int:    7,
				},
				Data: DescribedThing{What: "dynamic and described"},
			},
			wantContains: []string{
				`⯀ AnonymousField  govar.AnonymousField                    => {🞏 string string => |R:10| "anonymous!", 🞏 int int => 7}`,
				`⯀ Data            ⧉ govar.Describer(govar.DescribedThing) => {`,
				`⯀ What      string => |R:21| "dynamic and described"`,
				`⦿ Describe  func(govar.DescribedThing) string`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := SdumpNoColors(tt.input)
			for _, want := range tt.wantContains {
				if !strings.Contains(out, want) {
					t.Errorf("Dump %s: output missing expected fragment:\nwant contains:\n%s\ngot:\n%s", tt.name, want, out)
				}
			}
		})
	}
}
