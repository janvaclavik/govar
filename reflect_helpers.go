package govar

import (
	"fmt"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"unsafe"
)

// checkNilInterface returns the dynamic type and an empty value description
// if the given interface value is nil. Otherwise returns the real type
// and an empty string for the value.
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

// findCallerInStack inspects the call stack to locate the first caller
// not within the govar package. It returns (file, line, govarFuncName).
// govarFuncName is the last govar function in the stack, if any.
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

// findTypeMethods returns all exported methods associated with the given
// named reflect.Type (value or pointer receiver), excluding duplicates.
// Unnamed types (no PkgPath or Name) return an empty method slice.
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

// forceExported ensures unexported struct fields are accessible via reflection.
// If the value cannot be interfaced, but is addressable,
// it bypasses access restrictions via unsafe.NewAt.
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

// getFunctionName returns the full package-qualified function name from a reflect.Value.
func getFunctionName(v reflect.Value) string {
	return runtime.FuncForPC(v.Pointer()).Name()
}

// isNil returns true if v.Kind is a nilable type (Ptr, Slice, Map, Interface, Func, Chan)
// and the value is actually nil.
func isNil(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Interface, reflect.Func, reflect.Chan:
		return v.IsNil()
	default:
		return false
	}
}

// isSimpleCollection returns true if every element of the slice/array
// is a simple value (bool, number, or string).
func isSimpleCollection(v reflect.Value) bool {
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		if !isSimpleValue(elem) {
			return false
		}
	}
	return true
}

// isSimpleMap returns true if every key and value of the map is a simple value.
func isSimpleMap(v reflect.Value) bool {
	for _, key := range v.MapKeys() {
		val := v.MapIndex(key)
		if !isSimpleValue(key) || !isSimpleValue(val) {
			return false
		}
	}
	return true
}

// isSimpleValue returns true if the kind is one of the basic scalar types.
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

// makeAddressable returns an addressable value:
// if v is an unaddressable struct, it wraps it in a newly allocated pointer.
// Otherwise, returns v unchanged.
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

// sortMapKeys returns map keys sorted by a natural order for primitive types,
// or lexicographically by fmt.Sprintf for complex types.
func sortMapKeys(m reflect.Value) []reflect.Value {
	if m.Kind() != reflect.Map {
		return []reflect.Value{}
	}

	keys := m.MapKeys()
	if len(keys) == 0 {
		return []reflect.Value{}
	}

	// Sorting based on key type
	switch keys[0].Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].Int() < keys[j].Int()
		})
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].Uint() < keys[j].Uint()
		})
	case reflect.Float32, reflect.Float64:
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].Float() < keys[j].Float()
		})
	case reflect.String:
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].String() < keys[j].String()
		})
	default:
		// For complex object try their fmt string repres.
		sort.Slice(keys, func(i, j int) bool {
			return fmt.Sprintf("%+v", keys[i].Interface()) < fmt.Sprintf("%+v", keys[j].Interface())
		})
	}

	return keys
}

// toAddressableByteSlice returns a copy of a byte-like array/slice,
// ensuring the returned slice is addressable.
func toAddressableByteSlice(v reflect.Value) []byte {
	// Allocate and copy
	out := make([]byte, v.Len())
	for i := range v.Len() {
		out[i] = uint8(v.Index(i).Uint())
	}
	return out
}
