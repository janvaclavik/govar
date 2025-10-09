package govar

import (
	"strconv"
	"strings"
	"testing"
)

// TestReferenceScenarios runs a table-driven test covering all 19 complex
// reference and back-reference scenarios.
func TestReferenceScenarios(t *testing.T) {
	// --- Test Case #1: Cyclic struct references ---
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

	// --- Test Case #2: Nested pointers ---
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

	// --- Test Case #3: Ambiguous primitive pointers ---
	type PointerTest struct {
		SomeInt1, SomeInt2, SomeInt3, SomeInt4 int
		Pointers                               struct{ PointerToInt2, PointerToInt3, PointerToInt4 *int }
	}
	test3_val1, test3_val2, test3_val3, test3_val4 := 1000, -2000, -2000, 3000
	test3 := PointerTest{test3_val1, test3_val2, test3_val3, test3_val4,
		struct{ PointerToInt2, PointerToInt3, PointerToInt4 *int }{&test3_val2, &test3_val3, &test3_val4},
	}

	// --- Test Case #4: Deeper ID definition ---
	type DeeperIDPointerTest struct {
		PointerToInt2, PointerToInt3, PointerToInt4 *int
		Values                                      struct{ SomeInt1, SomeInt2, SomeInt3, SomeInt4 int }
	}
	test4_int1, test4_int2, test4_int3, test4_int4 := 111, 222, 333, 444
	test4 := DeeperIDPointerTest{&test4_int2, &test4_int3, &test4_int4,
		struct{ SomeInt1, SomeInt2, SomeInt3, SomeInt4 int }{test4_int1, test4_int2, test4_int3, test4_int4},
	}

	// --- Test Case #5: Interfaces, Maps, and Shared Structs ---
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

	// --- Test Case #6: Pointer-to-Pointer (Variadic) ---
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

	// --- Test Case #7: Shared Pointers in a Slice ---
	type Point struct{ X, Y int }
	p1 := &Point{X: 1, Y: 1}
	p2 := &Point{X: 2, Y: 2}
	zeroPoint := &Point{}
	pointSlice := []*Point{p1, p2, p1, zeroPoint, zeroPoint}

	// --- Test Case #8: Embedded Structs and References ---
	type Engine struct{ Horsepower int }
	type Car struct {
		Engine
		Name string
	}
	mustang := &Car{Engine: Engine{Horsepower: 450}, Name: "Mustang"}
	camaroSpec := &mustang.Engine
	carData := map[string]interface{}{"the_car": mustang, "engine_spec": camaroSpec}

	// --- Test Case #9: String pointers ---
	type StringTest struct {
		SomeString, SomeString2 string
		Pointers                struct{ PointerToString2 *string }
	}
	test9_val, test9_val2 := "some string", "second string"
	test9 := StringTest{test9_val, test9_val2, struct{ PointerToString2 *string }{&test9_val2}}

	// --- Test Case #10: Cross-Variable references ---
	type Node2 struct {
		ID   string
		Next *Node2
	}
	sharedNode := &Node2{ID: "Shared"}
	firstVar := map[string]*Node2{"entry1": sharedNode}
	secondVar := &sharedNode

	// --- Test Case #11: Unexported fields with pointers ---
	type UnexportedTest struct {
		ExportedInt1, ExportedInt2, ExportedInt3, ExportedInt4 int
		unexportedPointers                                     struct{ pointerToExported2, pointerToExported3, pointerToExported4 *int }
	}
	expInt1, expInt2, expInt3, expInt4 := 22, 33, 44, 55
	test11 := UnexportedTest{expInt1, expInt2, expInt3, expInt4,
		struct{ pointerToExported2, pointerToExported3, pointerToExported4 *int }{&expInt2, &expInt3, &expInt4},
	}

	// --- Test Case #12: No shared pointers ---
	var intArray [4]int
	var pointerToIntArray [4]*int
	intArray[0], intArray[1], intArray[2], intArray[3] = 9, 8, 9, 8
	for i := range 4 {
		pointerToIntArray[i] = &intArray[i]
	}

	// --- Test Case #13: Shared primitive pointers ---
	var longStringArray [4]string
	var pointerToStringArray [4]*string
	for i := range 4 {
		longStringArray[i] = strconv.Itoa(i)
		pointerToStringArray[i] = &longStringArray[i]
	}

	// --- Test Case #14: Pointers to Zero-Sized Structs (ZST) ---
	type ZST struct{}
	z1, z2 := ZST{}, ZST{}
	x1, x2 := &z1, &z2

	// --- Test Case #15: Slice of Interfaces with Shared Primitives ---
	sharedString := "shared"
	pointerToString := &sharedString
	mixedSlice := []any{sharedString, pointerToString, &sharedString}

	// --- Test Case #16: Direct Circular Reference (Slice) ---
	type RecursiveSlice []any
	s := make(RecursiveSlice, 1)
	s[0] = &s

	// --- Test Case #17: Multiple Pointers to an Embedded Struct Field ---
	type Inner struct{ Val int }
	type Outer struct{ I Inner }
	o := Outer{I: Inner{Val: 100}}
	v1 := &o.I
	v2 := &o.I

	// --- Test Case #18: Pointer-to-Pointer (Non-variadic) ---
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
		DeepPtr:   ppSharedInt_18,
		Container: Ptr18Container{Name: "OurContainer", IntPtr: pSharedInt_18},
	}

	// --- Test Case #19: Slices sharing sub-elements ---
	sharedSubSlice := []string{"shared", "data"}
	test19 := [][]string{sharedSubSlice, {"unique", "data"}, sharedSubSlice}

	// --- Test Case #20: Unreferenced Slice (Variadic) ---
	type Nod struct {
		Name     string
		Children []*Nod
		Link     *Nod
	}

	// Create the same 3-node hierarchy as before
	tc20_child2 := &Nod{Name: "Child-2"}
	tc20_child1 := &Nod{Name: "Child-1", Children: []*Nod{tc20_child2}}
	tc20_root := &Nod{Name: "Root", Children: []*Nod{tc20_child1}}
	tc20_root.Link = tc20_child2

	// --- Test Table ---
	testCases := []struct {
		name     string
		values   []any
		expected string
	}{
		{
			"TC1_CyclicStruct",
			[]any{people},
			`[]*govar.Person => |4| [
   0 => &1 {
      ⯀ Name   string        => |R:5| "Alice"
      ⯀ Age    int           => 44
      ⯀ Loves  *govar.Person => ↩︎ &2
   }
   1 => &2 {
      ⯀ Name   string        => |R:3| "Bob"
      ⯀ Age    int           => 23
      ⯀ Loves  *govar.Person => ↩︎ &3
   }
   2 => {
      ⯀ Name   string        => |R:3| "Joe"
      ⯀ Age    int           => 54
      ⯀ Loves  *govar.Person => ↩︎ &3
   }
   3 => &3 {
      ⯀ Name   string        => |R:5| "Penny"
      ⯀ Age    int           => 21
      ⯀ Loves  *govar.Person => ↩︎ &1
   }
]
`,
		},
		{
			"TC2_NestedPointers",
			[]any{test2},
			`govar.AnotherLevel => {
   ⯀ IntAlsoNoRef  int                  => 555
   ⯀ SubLevel      govar.SimplePtrTest2 => {
      ⯀ IntNoRef  int                 => 20
      ⯀ Int       int                 => &1 333
      ⯀ Int8      int8                => &2 -100
      ⯀ Int16     int16               => &3 7000
      ⯀ Ptrs      govar.InnerPointers => {
         ⯀ IntPtr    *int   => ↩︎ &1
         ⯀ Int8Ptr   *int8  => ↩︎ &2
         ⯀ Int16Ptr  *int16 => ↩︎ &3
      }
   }
}
`,
		},
		{
			"TC3_AmbiguousPrimitives",
			[]any{test3},
			`govar.PointerTest => {
   ⯀ SomeInt1  int                                                                   => 1000
   ⯀ SomeInt2  int                                                                   => &1 -2000
   ⯀ SomeInt3  int                                                                   => &2 -2000
   ⯀ SomeInt4  int                                                                   => &3 3000
   ⯀ Pointers  struct { PointerToInt2 *int; PointerToInt3 *int; PointerToInt4 *int } => {
      ⯀ PointerToInt2  *int => ↩︎ &1
      ⯀ PointerToInt3  *int => ↩︎ &2
      ⯀ PointerToInt4  *int => ↩︎ &3
   }
}
`,
		},
		{
			"TC4_DeeperID",
			[]any{test4},
			`govar.DeeperIDPointerTest => {
   ⯀ PointerToInt2  *int                                                              => ↩︎ &1
   ⯀ PointerToInt3  *int                                                              => ↩︎ &2
   ⯀ PointerToInt4  *int                                                              => ↩︎ &3
   ⯀ Values         struct { SomeInt1 int; SomeInt2 int; SomeInt3 int; SomeInt4 int } => {
      ⯀ SomeInt1  int => 111
      ⯀ SomeInt2  int => &1 222
      ⯀ SomeInt3  int => &2 333
      ⯀ SomeInt4  int => &3 444
   }
}
`,
		},
		{
			"TC5_InterfaceMap",
			[]any{dataMap},
			`map[string]any => |3| [
   "direct_A"  ⧉ any(*govar.Node) => &1 {
      ⯀ ID    string      => |R:1| "A"
      ⯀ Next  *govar.Node => ↩︎ &2
   }
   "direct_B"  ⧉ any(*govar.Node) => &2 {
      ⯀ ID    string      => |R:1| "B"
      ⯀ Next  *govar.Node => <nil>
   }
   "iface_A"   ⧉ any(*govar.Node) => ↩︎ &1
]
`,
		},
		{
			"TC6_PointerToPointerVariadic",
			[]any{container, sharedInt},
			`struct { DeepPtr **int; Container govar.PtrContainer } => {
   ⯀ DeepPtr    **int              => ↩︎ &1
   ⯀ Container  govar.PtrContainer => {
      ⯀ Name    string => |R:11| "MyContainer"
      ⯀ IntPtr  *int   => ↩︎ &1
   }
}

int => &1 999
`,
		},
		{
			"TC7_SharedPointersInSlice",
			[]any{pointSlice},
			`[]*govar.Point => |5| [
   0 => &1 {⯀ X int => 1, ⯀ Y int => 1}
   1 => {⯀ X int => 2, ⯀ Y int => 2}
   2 => ↩︎ &1
   3 => &2 {⯀ X int => 0, ⯀ Y int => 0}
   4 => ↩︎ &2
]
`,
		},
		{
			"TC8_EmbeddedStructRef",
			[]any{carData},
			`map[string]any => |2| [
   "engine_spec"  ⧉ any(*govar.Engine) => ↩︎ &1
   "the_car"      ⧉ any(*govar.Car)    => {
      ⯀ Engine  govar.Engine => &1 {⯀ Horsepower int => 450}
      ⯀ Name    string       => |R:7| "Mustang"
   }
]
`,
		},
		{
			"TC9_StringPointer",
			[]any{test9},
			`govar.StringTest => {
   ⯀ SomeString   string                              => |R:11| "some string"
   ⯀ SomeString2  string                              => &1 |R:13| "second string"
   ⯀ Pointers     struct { PointerToString2 *string } => {
      ⯀ PointerToString2  *string => ↩︎ &1
   }
}
`,
		},
		{
			"TC10_CrossVariable",
			[]any{firstVar, secondVar},
			`map[string]*govar.Node2 => |1| [
   "entry1" => &1 {
      ⯀ ID    string       => |R:6| "Shared"
      ⯀ Next  *govar.Node2 => <nil>
   }
]

**govar.Node2 => ↩︎ &1
`,
		},
		{
			"TC11_UnexportedPointers",
			[]any{test11},
			`govar.UnexportedTest => {
   ⯀ ExportedInt1        int                                                                                  => 22
   ⯀ ExportedInt2        int                                                                                  => &1 33
   ⯀ ExportedInt3        int                                                                                  => &2 44
   ⯀ ExportedInt4        int                                                                                  => &3 55
   🞏 unexportedPointers  struct { pointerToExported2 *int; pointerToExported3 *int; pointerToExported4 *int } => {
      🞏 pointerToExported2  *int => ↩︎ &1
      🞏 pointerToExported3  *int => ↩︎ &2
      🞏 pointerToExported4  *int => ↩︎ &3
   }
}
`,
		},
		{
			"TC12_NoSharedPointers",
			[]any{pointerToIntArray},
			`[4]*int => |4| [
   0 => 9
   1 => 8
   2 => 9
   3 => 8
]
`,
		},
		{
			"TC13_SharedPrimitivePointers",
			[]any{longStringArray, pointerToStringArray},
			`[4]string => |4| [0 => &1 |R:1| "0", 1 => &2 |R:1| "1", 2 => &3 |R:1| "2", 3 => &4 |R:1| "3"]

[4]*string => |4| [
   0 => ↩︎ &1
   1 => ↩︎ &2
   2 => ↩︎ &3
   3 => ↩︎ &4
]
`,
		},
		{
			"TC14_ZST",
			[]any{x1, x2, z1},
			`*govar.ZST => ↩︎ &1

*govar.ZST => ↩︎ &1

govar.ZST => &1 {}
`,
		},
		{
			"TC15_InterfaceSlicePrimitives",
			[]any{mixedSlice},
			`[]any => |3| [
   0 ⧉ any(string)  => &1 |R:6| "shared"
   1 ⧉ any(*string) => ↩︎ &1
   2 ⧉ any(*string) => ↩︎ &1
]
`,
		},
		{
			"TC16_RecursiveSlice",
			[]any{s},
			`govar.RecursiveSlice => &1 |1| [
   0 ⧉ any(*govar.RecursiveSlice) => ↩︎ &1
]
`,
		},
		{
			"TC17_EmbeddedFieldPointers",
			[]any{o, v1, v2},
			`govar.Outer => {
   ⯀ I  govar.Inner => &1 {⯀ Val int => 100}
}

*govar.Inner => ↩︎ &1

*govar.Inner => ↩︎ &1
`,
		},
		{
			"TC18_PointerToPointer",
			[]any{test18},
			`struct { DeepPtr **int; Container govar.Ptr18Container } => {
   ⯀ DeepPtr    **int                => ↩︎ &1
   ⯀ Container  govar.Ptr18Container => {
      ⯀ Name    string => |R:12| "OurContainer"
      ⯀ IntPtr  *int   => &1 666
   }
}
`,
		},
		{
			"TC19_SharedSubSlice",
			[]any{test19},
			`[][]string => |3| [
   0 []string => &1 |2| [0 => |R:6| "shared", 1 => |R:4| "data"]
   1 []string => |2| [0 => |R:6| "unique", 1 => |R:4| "data"]
   2 []string => ↩︎ &1
]
`,
		},
		{
			"TC20_UnreferencedSliceVariadic",
			[]any{tc20_root, tc20_child1, tc20_child2},
			`*govar.Nod => {
   ⯀ Name      string       => |R:4| "Root"
   ⯀ Children  []*govar.Nod => |1| [
      0 => ↩︎ &2
   ]
   ⯀ Link      *govar.Nod   => ↩︎ &1
}

*govar.Nod => &2 {
   ⯀ Name      string       => |R:7| "Child-1"
   ⯀ Children  []*govar.Nod => |1| [
      0 => ↩︎ &1
   ]
   ⯀ Link      *govar.Nod   => <nil>
}

*govar.Nod => &1 {
   ⯀ Name      string       => |R:7| "Child-2"
   ⯀ Children  []*govar.Nod => <nil>
   ⯀ Link      *govar.Nod   => <nil>
}
`,
		},
	}

	// Create a dumper with a fixed config for stable test output.
	// No colors, no header.
	dumper := NewDumper(DumperConfig{
		IndentWidth:         3,
		MaxDepth:            15,
		MaxItems:            150,
		MaxStringLen:        10000,
		MaxInlineLength:     80,
		ShowTypes:           true,
		UseColors:           false, // No colors for stable comparison
		TrackReferences:     true,
		EmbedTypeMethods:    true,
		ShowMetaInformation: true,
		ShowHexdump:         true,
		IgnoreStringer:      false,
	})

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Sdump generates the output without the header.
			actual := dumper.Sdump(tc.values...)
			// Trim whitespace to make comparisons robust.
			actual = strings.TrimSpace(actual)
			expected := strings.TrimSpace(tc.expected)

			if !strings.Contains(actual, expected) {
				t.Errorf("Output mismatch.\nEXPECTED:\n%s\n\nGOT:\n%s", expected, actual)
			}
		})
	}
}
