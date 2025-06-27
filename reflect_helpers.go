package govar

import (
	"reflect"
	"runtime"
	"strings"
	"unsafe"
)

// checkNilInterface returns safe type/value info about possible nil interface
func checkNilInterface(v any) (string, string) {
	rt := reflect.TypeOf(v)

	resolvedType := ""
	resolvedValue := ""
	if v == nil {
		resolvedValue = "<nil>"
		if rt != nil {
			resolvedType = rt.String()
		} else {
			resolvedType = "unknown"
		}
		return resolvedType, resolvedValue
	}
	return rt.String(), resolvedValue
}

// isNil checks if the value is nil on any kind of object
// It does not fail even if the value type cannot be nil (bool, etc...)
func isNil(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Interface, reflect.Func, reflect.Chan:
		return v.IsNil()
	default:
		return false
	}
}

func isSimpleValue(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}

	switch v.Kind() {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32, reflect.Float64,
		reflect.String:
		return true
	default:
		return false
	}
}

func isSimpleCollection(v reflect.Value) bool {
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		if !isSimpleValue(elem) {
			return false
		}
	}
	return true
}

func isSimpleMap(v reflect.Value) bool {
	for _, key := range v.MapKeys() {
		val := v.MapIndex(key)
		if !isSimpleValue(key) || !isSimpleValue(val) {
			return false
		}
	}
	return true
}

// findCallerInStack finds the first non-govar function call in the call-stack.
func findCallerInStack() (string, int, string) {
	govarFuncName := ""
	for i := 2; i < 15; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		if fn == nil || !strings.Contains(fn.Name(), "/"+PackageName) {
			return file, line, govarFuncName
		}
		tmpNameSliced := strings.Split(fn.Name(), "/")
		govarFuncName = tmpNameSliced[len(tmpNameSliced)-1]
	}
	return "", 0, ""
}

func findTypeMethods(typ reflect.Type) []reflect.Method {
	seen := make(map[string]bool)
	methods := []reflect.Method{}

	if typ.Name() == "" || typ.PkgPath() == "" {
		// ignore if it's NOT a named type
		return methods
	}

	// Check value receiver methods
	for i := range typ.NumMethod() {
		m := typ.Method(i)
		if !seen[m.Name] {
			methods = append(methods, m)
			seen[m.Name] = true
		}
	}

	// Check pointer receiver methods if applicable
	if typ.Kind() != reflect.Ptr {
		ptrType := reflect.PointerTo(typ)
		for i := range ptrType.NumMethod() {
			m := ptrType.Method(i)
			if !seen[m.Name] {
				methods = append(methods, m)
				seen[m.Name] = true
			}
		}
	}

	return methods
}

// forceExported returns a value that is guaranteed to be exported, even if it is unexported.
func forceExported(v reflect.Value) reflect.Value {
	if v.CanInterface() {
		return v
	}
	if v.CanAddr() {
		return reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem()
	}
	// Final fallback: return original value, even if unexported
	return v
}

func getFunctionName(v reflect.Value) string {
	return runtime.FuncForPC(v.Pointer()).Name()
}

// makeAddressable ensures the value is addressable, wrapping structs in pointers if necessary.
func makeAddressable(v reflect.Value) reflect.Value {
	// Already addressable? Do nothing
	if v.CanAddr() {
		return v
	}

	// If it's a struct and not addressable, wrap it in a pointer
	if v.Kind() == reflect.Struct {
		ptr := reflect.New(v.Type())
		ptr.Elem().Set(v)
		return ptr.Elem()
	}

	return v
}

// toAddressableByteSlice is a safe fallback helper for
// creating an addressable copy of a potentialy unaddressable array
// Parameters:
//   - v reflect.Value
//
// Returns:
//   - []byte
func toAddressableByteSlice(v reflect.Value) []byte {
	// Allocate and copy
	out := make([]byte, v.Len())
	for i := range v.Len() {
		out[i] = uint8(v.Index(i).Uint())
	}
	return out
}
