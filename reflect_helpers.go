// Package govar provides a powerful and highly configurable pretty-printer for Go
// data structures. This file contains helper functions that use the `reflect`
// and `unsafe` packages to perform low-level inspection of values, types,
// and memory addresses.
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
// ("<nil>") if the given interface value is nil. Otherwise returns the real type
// and an empty string for the value. If the value was nil and resolved type is
// also nil, retuned type is "unknown"
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

// deref repeatedly dereferences a pointer or interface until it reaches a non-pointer,
// non-interface value, or nil. It returns the underlying value.
func deref(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return reflect.Value{} // Return an invalid value for nil pointers/interfaces.
		}
		v = v.Elem()
	}
	return v
}

// findCallerInStack inspects the call stack to locate the first caller
// not within the govar package. It returns the file, line number, and the name
// of the govar function that was called (e.g., "govar.Dump").
func findCallerInStack() (string, int, string) {
	govarFuncName := ""
	for i := 1; i < 15; i++ {
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
// named reflect.Type, considering both value and pointer receivers.
// It avoids duplicates if a method exists on both.
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

// getFunctionName returns the full package-qualified function name from a reflect.Value.
func getFunctionName(v reflect.Value) string {
	return runtime.FuncForPC(v.Pointer()).Name()
}

// getIndirectionLevel calculates the level of pointer indirection for a value.
// For example: T -> 0, *T -> 1, **T -> 2. It unwraps interfaces.
// This is crucial for prioritizing definition points.
func getIndirectionLevel(v reflect.Value) int {
	level := 0
	// Unwrap interfaces to get to the concrete value.
	for v.Kind() == reflect.Interface {
		if v.IsNil() {
			return level
		}
		v = v.Elem()
	}
	// Count pointer indirections.
	for v.Kind() == reflect.Ptr {
		level++
		if v.IsNil() {
			break
		}
		v = v.Elem()
	}
	return level
}

// getValPtr gets the memory address of a value as an unsafe.Pointer if possible.
// This is used for basic cycle detection.
func getValPtr(v reflect.Value) unsafe.Pointer {
	switch v.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func:
		if v.IsNil() {
			return nil
		}
		return unsafe.Pointer(v.Pointer())
	case reflect.Interface:
		if v.IsNil() {
			return nil
		}
		return getValPtr(v.Elem())
	default:
		if v.CanAddr() {
			return unsafe.Pointer(v.UnsafeAddr())
		}
		return nil
	}
}

// isCompositeOrInterface returns true if the kind is a composite type or an interface.
// These are the types that can contain other values and need to be traversed.
func isCompositeOrInterface(k reflect.Kind) bool {
	switch k {
	case reflect.Struct, reflect.Map, reflect.Slice, reflect.Array, reflect.Interface:
		return true
	default:
		return false
	}
}

// isNil returns true if v.Kind is a nilable type (Ptr, Slice, Map, Interface, Func, Chan)
// and its value is nil.
func isNil(v reflect.Value) bool {
	// This helper must be able to check unexported fields, so we use tryExport.
	v = tryExport(v)
	if !v.IsValid() {
		return true
	}
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		return v.IsNil()
	}
	return false
}

// isPrimitiveKind returns true if the kind is a basic scalar type (bool, number, string).
func isPrimitiveKind(k reflect.Kind) bool {
	switch k {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.String:
		return true
	default:
		return false
	}
}

// isPointerRef returns true if the value is a pointer or an interface containing a pointer.
func isPointerRef(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Ptr:
		return true
	case reflect.Interface:
		if v.IsNil() {
			return false
		}
		// Recurse on the interface's dynamic value.
		return isPointerRef(v.Elem())
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
// if v is an unaddressable value, it wraps it in a newly allocated pointer.
// Otherwise, returns v unchanged.
func makeAddressable(v reflect.Value) reflect.Value {
	// A guard clause to handle zero/invalid reflect.Values.
	if !v.IsValid() {
		return v
	}

	if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		return v
	}
	if !v.CanAddr() {
		temp := reflect.New(v.Type()).Elem()
		if v.IsValid() {
			temp.Set(v)
		}
		return temp
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

// tryExport returns an interfaceable version of v if possible.
// If v represents an unexported field but is addressable, it uses unsafe
// to create an accessible copy. This is crucial for inspecting private fields.
func tryExport(v reflect.Value) reflect.Value {
	if v.CanInterface() {
		return v
	}
	if v.CanAddr() {
		return reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem()
	}
	// Final fallback: return original value, even if unexported
	return v
}
