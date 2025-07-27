package govar

import (
	"reflect"
	"strings"
	"testing"
)

func TestCheckNilInterface(t *testing.T) {
	var x any
	typ, val := checkNilInterface(x)
	if !strings.Contains(typ, "unknown") && typ != "" {
		t.Errorf("expected nil or empty type for nil interface, got %q", typ)
	}
	if val != "<nil>" {
		t.Errorf("expected empty value for nil interface, got %q", val)
	}

	var p *int
	var y any = p
	typ, val = checkNilInterface(y)
	if !strings.Contains(typ, "*int") {
		t.Errorf("expected *int type, got %q", typ)
	}
	if val != "" {
		t.Errorf("expected empty value for nil pointer interface, got %q", val)
	}
}

func TestFindCallerInStack(t *testing.T) {
	file, line, govarFunc := findCallerInStack()
	if file == "" || line <= 0 {
		t.Errorf("did not find valid caller info; file=%q line=%d", file, line)
	}
	if !strings.Contains(govarFunc, "Test") {
		t.Errorf("expected govarFunc to mention test func, got %q", govarFunc)
	}
}

type exampleStruct struct{}

func (exampleStruct) Foo()  {}
func (*exampleStruct) Bar() {}

func TestFindTypeMethods(t *testing.T) {
	typ := reflect.TypeOf(exampleStruct{})
	methods := findTypeMethods(typ)
	got := make(map[string]bool)
	for _, m := range methods {
		got[m.Name] = true
	}
	for _, want := range []string{"Foo", "Bar"} {
		if !got[want] {
			t.Errorf("expected method %q not found in findTypeMethods result", want)
		}
	}
}

func TestForceExported(t *testing.T) {
	type s struct {
		x int
	}

	val := s{x: 123}
	rv := reflect.ValueOf(val)
	field := rv.Field(0) // unexported field

	if field.CanInterface() {
		t.Fatal("expected field to be non-interfaceable")
	}
	exported := tryExport(field)

	// Check if we *safely* got the value
	if exported.CanInterface() {
		got := exported.Interface().(int)
		if got != 123 {
			t.Errorf("forceExported value mismatch: got %v, want 123", got)
		}
	} else {
		t.Log("forceExported could not access unexported field (as expected in some cases)")
	}
}

func TestGetFunctionName(t *testing.T) {
	f := func() {}
	name := getFunctionName(reflect.ValueOf(f))
	if !strings.Contains(name, "TestGetFunctionName") {
		t.Errorf("expected function name to include test name, got %q", name)
	}
}

func TestMakeAddressable(t *testing.T) {
	s := struct{ A int }{A: 7}
	v := reflect.ValueOf(s)
	v2 := makeAddressable(v)
	if !v2.CanAddr() {
		t.Fatal("makeAddressable should return addressable value for struct")
	}
	if v2.FieldByName("A").Interface().(int) != 7 {
		t.Errorf("struct field mismatch: got %v", v2.FieldByName("A").Interface())
	}
}

func TestSortMapKeys(t *testing.T) {
	m := map[string]int{"a": 2, "b": 1}
	rv := reflect.ValueOf(m)
	keys := sortMapKeys(rv)
	if len(keys) != 2 || keys[0].String() != "a" {
		t.Errorf("expected sorted keys [a b], got %v", keys)
	}

	m2 := map[int]string{2: "2", 1: "1"}
	rv2 := reflect.ValueOf(m2)
	keys2 := sortMapKeys(rv2)
	if len(keys2) != 2 || keys2[0].Int() != 1 {
		t.Errorf("expected sorted int keys [1 2], got %v", keys2)
	}
}

func TestToAddressableByteSlice(t *testing.T) {
	arr := [3]uint8{1, 2, 3}
	rv := reflect.ValueOf(arr)
	b := toAddressableByteSlice(rv)
	if len(b) != 3 || b[0] != 1 || b[1] != 2 || b[2] != 3 {
		t.Errorf("toAddressableByteSlice failed, got %v", b)
	}
}

func TestIsNilAndSimple(t *testing.T) {
	if !isNil(reflect.ValueOf((*int)(nil))) {
		t.Error("expected isNil to detect nil pointer")
	}
	if isNil(reflect.ValueOf(0)) {
		t.Error("expected isNil(false) to be false")
	}
	if !isSimpleValue(reflect.ValueOf(3.14)) {
		t.Error("expected isSimpleValue float to be true")
	}
	// Slice of ints is simple collection
	slice := reflect.ValueOf([]int{1, 2})
	if !isSimpleCollection(slice) {
		t.Error("expected isSimpleCollection true for simple slice")
	}
	// A slice of slices is not simple since nested is not simple value
	nested := reflect.ValueOf([][]int{{1}})
	if isSimpleCollection(nested) {
		t.Error("expected isSimpleCollection false for nested slice")
	}
}

func TestIsSimpleMap(t *testing.T) {
	m := reflect.ValueOf(map[string]int{"a": 1})
	if !isSimpleMap(m) {
		t.Error("expected isSimpleMap true for simple map")
	}
	m2 := reflect.ValueOf(map[string][]int{"a": {1}})
	if isSimpleMap(m2) {
		t.Error("expected isSimpleMap false for map with complex values")
	}
}

func TestDeref(t *testing.T) {
	num := 42
	pNum := &num
	ppNum := &pNum
	var iface any = ppNum

	// Test dereferencing a pointer
	v1 := deref(reflect.ValueOf(pNum))
	if v1.Int() != 42 {
		t.Errorf("deref(pNum) = %v, want 42", v1.Int())
	}

	// Test dereferencing multiple levels
	v2 := deref(reflect.ValueOf(ppNum))
	if v2.Int() != 42 {
		t.Errorf("deref(ppNum) = %v, want 42", v2.Int())
	}

	// Test dereferencing through an interface
	v3 := deref(reflect.ValueOf(iface))
	if v3.Int() != 42 {
		t.Errorf("deref(interface) = %v, want 42", v3.Int())
	}

	// Test dereferencing a nil pointer
	var nilPtr *int
	v4 := deref(reflect.ValueOf(nilPtr))
	if v4.IsValid() {
		t.Errorf("deref(nilPtr) should be invalid, but was valid")
	}
}

func TestGetIndirectionLevel(t *testing.T) {
	num := 42
	pNum := &num
	ppNum := &pNum
	var iface any = pNum

	testCases := []struct {
		name     string
		value    any
		expected int
	}{
		{"non-pointer", num, 0},
		{"single pointer", pNum, 1},
		{"double pointer", ppNum, 2},
		{"interface with pointer", iface, 1},
		{"nil pointer", (*int)(nil), 1},
		{"nil interface", (any)(nil), 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			level := getIndirectionLevel(reflect.ValueOf(tc.value))
			if level != tc.expected {
				t.Errorf("getIndirectionLevel() = %d; want %d", level, tc.expected)
			}
		})
	}
}

func TestGetValPtr(t *testing.T) {
	num := 10
	pNum := &num
	s := []int{1}
	m := map[string]int{"a": 1}

	// Test pointer types
	if p := getValPtr(reflect.ValueOf(pNum)); p == nil {
		t.Error("getValPtr for a pointer should not be nil")
	}
	if p := getValPtr(reflect.ValueOf(s)); p == nil {
		t.Error("getValPtr for a slice should not be nil")
	}
	if p := getValPtr(reflect.ValueOf(m)); p == nil {
		t.Error("getValPtr for a map should not be nil")
	}

	// Test an addressable non-pointer value
	// We get an addressable value by taking a pointer and then getting its element.
	addressableNum := reflect.ValueOf(&num).Elem()
	if p := getValPtr(addressableNum); p == nil {
		t.Error("getValPtr for an addressable value should not be nil")
	}

	// Test nil cases
	if p := getValPtr(reflect.ValueOf((*int)(nil))); p != nil {
		t.Error("getValPtr for a nil pointer should be nil")
	}
}

func TestIsCompositeOrInterface(t *testing.T) {
	testCases := []struct {
		name     string
		kind     reflect.Kind
		expected bool
	}{
		{"struct", reflect.Struct, true},
		{"slice", reflect.Slice, true},
		{"map", reflect.Map, true},
		{"array", reflect.Array, true},
		{"interface", reflect.Interface, true},
		{"int", reflect.Int, false},
		{"string", reflect.String, false},
		{"pointer", reflect.Ptr, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isCompositeOrInterface(tc.kind); got != tc.expected {
				t.Errorf("isCompositeOrInterface(%v) = %v; want %v", tc.kind, got, tc.expected)
			}
		})
	}
}

func TestIsNil(t *testing.T) {
	if !isNil(reflect.ValueOf((*int)(nil))) {
		t.Error("expected isNil to be true for nil pointer")
	}
	if isNil(reflect.ValueOf(0)) {
		t.Error("expected isNil to be false for zero int")
	}
	if isNil(reflect.ValueOf("")) {
		t.Error("expected isNil to be false for empty string")
	}
	var s []int
	if !isNil(reflect.ValueOf(s)) {
		t.Error("expected isNil to be true for nil slice")
	}
}

func TestIsPointerRef(t *testing.T) {
	num := 42
	pNum := &num
	var ifaceWithPtr any = pNum
	var ifaceWithVal any = num

	if !isPointerRef(reflect.ValueOf(pNum)) {
		t.Error("isPointerRef should be true for a pointer")
	}
	if !isPointerRef(reflect.ValueOf(ifaceWithPtr)) {
		t.Error("isPointerRef should be true for an interface containing a pointer")
	}
	if isPointerRef(reflect.ValueOf(num)) {
		t.Error("isPointerRef should be false for a non-pointer value")
	}
	if isPointerRef(reflect.ValueOf(ifaceWithVal)) {
		t.Error("isPointerRef should be false for an interface containing a non-pointer value")
	}
}

func TestIsPrimitiveKind(t *testing.T) {
	if !isPrimitiveKind(reflect.Int) {
		t.Error("isPrimitiveKind(Int) should be true")
	}
	if !isPrimitiveKind(reflect.String) {
		t.Error("isPrimitiveKind(String) should be true")
	}
	if isPrimitiveKind(reflect.Struct) {
		t.Error("isPrimitiveKind(Struct) should be false")
	}
	if isPrimitiveKind(reflect.Slice) {
		t.Error("isPrimitiveKind(Slice) should be false")
	}
}

func TestIsSimpleCollection(t *testing.T) {
	// Slice of ints is simple collection
	slice := reflect.ValueOf([]int{1, 2})
	if !isSimpleCollection(slice) {
		t.Error("expected isSimpleCollection true for simple slice")
	}
	// A slice of slices is not simple since nested is not simple value
	nested := reflect.ValueOf([][]int{{1}})
	if isSimpleCollection(nested) {
		t.Error("expected isSimpleCollection false for nested slice")
	}
}

func TestTryExport(t *testing.T) {
	type s struct {
		x int
	}

	val := s{x: 123}
	// Get an addressable reflect.Value of the struct.
	rv := reflect.ValueOf(&val).Elem()
	field := rv.Field(0) // This field is now addressable.

	if field.CanInterface() {
		t.Fatal("Test setup error: field should be unexported and not interfaceable")
	}

	exported := tryExport(field)

	// Check if we *safely* got the value.
	if exported.CanInterface() {
		got := exported.Interface().(int)
		if got != 123 {
			t.Errorf("tryExport value mismatch: got %v, want 123", got)
		}
	} else {
		t.Errorf("tryExport could not access unexported field")
	}
}
