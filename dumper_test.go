package govar

import (
	"strings"
	"testing"
)

func TestDumpBasicTypes(t *testing.T) {
	// Dump int
	result := Sdump(42)
	if !strings.Contains(result, "int") || !strings.Contains(result, "42") {
		t.Errorf("result(int) output = %q; want output containing 'int' and '42'", result)
	}

	// Dump string
	result = Sdump("hello")
	if !strings.Contains(result, "hello") {
		t.Errorf("Sdump(string) output = %q; want output containing '\"hello\"'", result)
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
