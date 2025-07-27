package main

import (
	"fmt"
	"strconv"

	"github.com/janvaclavik/govar"
)

func main() {
	// --- Test Case #1: Cyclic struct references ---
	fmt.Println("--- Test Case #1 ---")
	type Person struct {
		Name  string
		Age   int
		Loves *Person
	}
	alice := &Person{Name: "Alice", Age: 44}
	bob := &Person{Name: "Bob", Age: 23}
	penny := &Person{Name: "Penny", Age: 21}
	joe := &Person{Name: "Joe", Age: 54}
	alice.Loves = bob
	bob.Loves = penny
	penny.Loves = alice
	joe.Loves = penny
	people := []*Person{alice, bob, joe, penny}
	govar.Dump(people)

	// --- Test Case #2 ---
	fmt.Println("\n\n--- Test Case #2 ---")
	type InnerPointers struct {
		IntPtr   *int
		Int8Ptr  *int8
		Int16Ptr *int16
	}
	type SimplePtrTest2 struct {
		IntNoRef, Int int
		Int8          int8
		Int16         int16
		Ptrs          InnerPointers
	}
	type AnotherLevel struct {
		IntAlsoNoRef int
		SubLevel     SimplePtrTest2
	}
	test2_int, test2_int8, test2_int16 := 333, int8(-100), int16(7000)
	test2 := AnotherLevel{555, SimplePtrTest2{20, test2_int, test2_int8, test2_int16,
		InnerPointers{&test2_int, &test2_int8, &test2_int16}},
	}
	govar.Dump(test2)

	// --- Test Case #3 ---
	fmt.Println("\n\n--- Test Case #3 ---")
	type PointerTest struct {
		SomeInt1, SomeInt2, SomeInt3, SomeInt4 int
		Pointers                               struct{ PointerToInt2, PointerToInt3, PointerToInt4 *int }
	}
	test3_val1, test3_val2, test3_val3, test3_val4 := 1000, -2000, -2000, 3000
	test3 := PointerTest{test3_val1, test3_val2, test3_val3, test3_val4,
		struct{ PointerToInt2, PointerToInt3, PointerToInt4 *int }{&test3_val2, &test3_val3, &test3_val4},
	}
	govar.Dump(test3)

	// --- Test Case #4 ---
	fmt.Println("\n\n--- Test Case #4 ---")
	type DeeperIDPointerTest struct {
		PointerToInt2, PointerToInt3, PointerToInt4 *int
		Values                                      struct{ SomeInt1, SomeInt2, SomeInt3, SomeInt4 int }
	}
	test4_int1, test4_int2, test4_int3, test4_int4 := 111, 222, 333, 444
	test4 := DeeperIDPointerTest{&test4_int2, &test4_int3, &test4_int4,
		struct{ SomeInt1, SomeInt2, SomeInt3, SomeInt4 int }{test4_int1, test4_int2, test4_int3, test4_int4},
	}
	govar.Dump(test4)

	// --- Test Case #5: Interfaces, Maps, and Shared Structs ---
	fmt.Println("\n\n--- Test Case #5 ---")
	type Node struct {
		ID   string
		Next *Node
	}
	nodeA := &Node{ID: "A"}
	nodeB := &Node{ID: "B"}
	nodeA.Next = nodeB
	var ifaceA interface{} = nodeA
	dataMap := map[string]interface{}{
		"direct_A": nodeA,
		"direct_B": nodeB,
		"iface_A":  ifaceA,
	}
	govar.Dump(dataMap)

	// --- Test Case #6: Pointer-to-Pointer and Shared Primitives (Variadic version) ---
	fmt.Println("\n\n--- Test Case #6 ---")
	type PtrContainer struct {
		Name   string
		IntPtr *int
	}
	sharedInt := 999
	pSharedInt := &sharedInt
	ppSharedInt := &pSharedInt
	container := struct {
		DeepPtr   **int
		Container PtrContainer
	}{
		DeepPtr: ppSharedInt,
		Container: PtrContainer{
			Name:   "MyContainer",
			IntPtr: pSharedInt,
		},
	}
	govar.Dump(container, sharedInt)

	// --- Test Case #7: Shared Pointers in a Slice (including shared zero value) ---
	fmt.Println("\n\n--- Test Case #7 ---")
	type Point struct{ X, Y int }
	p1 := &Point{X: 1, Y: 1}
	p2 := &Point{X: 2, Y: 2}
	zeroPoint := &Point{}
	pointSlice := []*Point{
		p1,
		p2,
		p1,
		zeroPoint,
		zeroPoint,
	}
	govar.Dump(pointSlice)

	// --- Test Case #8: Embedded Structs and References ---
	fmt.Println("\n\n--- Test Case #8 ---")
	type Engine struct {
		Horsepower int
	}
	type Car struct {
		Engine
		Name string
	}
	mustang := &Car{
		Engine: Engine{Horsepower: 450},
		Name:   "Mustang",
	}
	camaroSpec := &mustang.Engine
	carData := map[string]interface{}{
		"the_car":     mustang,
		"engine_spec": camaroSpec,
	}
	govar.Dump(carData)

	// --- Test Case #9 ---
	fmt.Println("\n\n--- Test Case #9 ---")
	type StringTest struct {
		SomeString, SomeString2 string
		Pointers                struct{ PointerToString2 *string }
	}
	test9_val, test9_val2 := "some string", "second string"
	test9 := StringTest{test9_val, test9_val2,
		struct{ PointerToString2 *string }{&test9_val2},
	}
	govar.Dump(test9)

	// --- Test Case #10: Cross-Variable and Simple Cycle ---
	fmt.Println("\n\n--- Test Case #10: Cross-Variable ---")
	type Node2 struct { // Renamed to avoid conflict with TC #5
		ID   string
		Next *Node2
	}
	sharedNode := &Node2{ID: "Shared"}
	firstVar := map[string]*Node2{"entry1": sharedNode}
	secondVar := &sharedNode
	// This will show cross-variable refs.
	govar.Dump(firstVar, secondVar)

	// --- Test Case #11 ---
	fmt.Println("\n\n--- Test Case #11 ---")
	type UnexportedTest struct {
		ExportedInt1, ExportedInt2, ExportedInt3, ExportedInt4 int
		unexportedPointers                                     struct{ pointerToExported2, pointerToExported3, pointerToExported4 *int }
	}
	ExportedInt1, ExportedInt2, ExportedInt3, ExportedInt4 := 22, 33, 44, 55
	test11 := UnexportedTest{ExportedInt1, ExportedInt2, ExportedInt3, ExportedInt4,
		struct{ pointerToExported2, pointerToExported3, pointerToExported4 *int }{&ExportedInt2, &ExportedInt3, &ExportedInt4},
	}
	govar.Dump(test11)

	// --- Test Case #12 ---
	fmt.Println("\n\n--- Test Case #12 ---")
	var intArray [4]int
	var pointerToIntArray [4]*int
	intArray[0] = 9
	intArray[1] = 8
	intArray[2] = 9
	intArray[3] = 8
	for i := range 4 {
		pointerToIntArray[i] = &intArray[i]
	}
	// This case has no shared pointers, so no IDs should be printed.
	govar.Dump(pointerToIntArray)

	// --- Test Case #13 ---
	fmt.Println("\n\n--- Test Case #13 ---")
	var longStringArray [4]string
	var pointerToStringArray [4]*string
	for i := range 4 {
		longStringArray[i] = strconv.Itoa(i)
		pointerToStringArray[i] = &longStringArray[i]
	}
	govar.Dump(longStringArray, pointerToStringArray)

	// --- Test Case #14 Pointers to Zero-Sized Structs (ZST) ---
	fmt.Println("\n\n--- Test Case #14 ---")
	type ZST struct{}
	z1, z2 := ZST{}, ZST{}
	x1, x2 := &z1, &z2
	govar.Dump(x1, x2, z1)

	// --- Test Case #15 Slice of Interfaces with Shared Primitives ---
	fmt.Println("\n\n--- Test Case #15 ---")
	sharedString := "shared"
	pointerToString := &sharedString
	// A slice containing the value itself and a pointer to it
	mixedSlice := []any{sharedString, pointerToString, &sharedString}
	govar.Dump(mixedSlice)

	// --- Test Case #16 Direct Circular Reference (Slice) ---
	fmt.Println("\n\n--- Test Case #16 ---")
	type RecursiveSlice []any
	s := make(RecursiveSlice, 1)
	s[0] = &s // The slice contains a pointer to itself
	govar.Dump(s)

	// --- Test Case #17 Multiple Pointers to an Embedded Struct Field ---
	fmt.Println("\n\n--- Test Case #17 ---")
	type Inner struct{ Val int }
	type Outer struct{ I Inner }
	o := Outer{I: Inner{Val: 100}}
	v1 := &o.I // Pointer to the embedded field
	v2 := &o.I // Another pointer to the same field
	govar.Dump(o, v1, v2)

	// --- Test Case #18: Pointer-to-Pointer and Shared Primitives (Non-variadic version) ---
	fmt.Println("\n\n--- Test Case #18 ---")
	type Ptr18Container struct {
		Name   string
		IntPtr *int
	}
	sharedInt_18 := 666
	pSharedInt_18 := &sharedInt_18
	ppSharedInt_18 := &pSharedInt_18
	test18 := struct {
		DeepPtr   **int
		Container Ptr18Container
	}{
		DeepPtr: ppSharedInt_18,
		Container: Ptr18Container{
			Name:   "OurContainer",
			IntPtr: pSharedInt_18,
		},
	}
	govar.Dump(test18)

	// --- Test Case #19: Slices/Maps sharing sub-elements ---
	fmt.Println("\n\n--- Test Case #19 ---")
	// A slice that is shared between other slices
	sharedSubSlice := []string{"shared", "data"}
	test19 := [][]string{
		sharedSubSlice,
		{"unique", "data"},
		sharedSubSlice,
	}
	govar.Dump(test19)
}
