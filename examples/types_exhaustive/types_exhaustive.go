package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net"
	"strings"
	"time"
	"unsafe"

	"github.com/janvaclavik/govar"
)

type IntType int
type Int8Type int8
type Int16Type int16
type Int32Type int32
type Int64Type int64
type UintType uint
type Uint8Type uint8
type Uint16Type uint16
type Uint32Type uint32
type Uint64Type uint64
type Float32Type float32
type Float64Type float64
type Complex64Type complex64
type Complex128Type complex128
type Bool1Type bool
type Bool2Type bool
type StringType string
type UintptrType uintptr

type IntPtrType *int
type Int8PtrType *int8
type Int16PtrType *int16
type Int32PtrType *int32
type Int64PtrType *int64
type UintPtrType *uint
type Uint8PtrType *uint8
type Uint16PtrType *uint16
type Uint32PtrType *uint32
type Uint64PtrType *uint64
type Float32PtrType *float32
type Float64PtrType *float64
type Complex64PtrType *complex64
type Complex128PtrType *complex128
type Bool1PtrType *bool
type Bool2PtrType *bool
type StringPtrType *string
type UintptrPtrType *uintptr

type FuncType func()
type Func2Type func(int) float64
type Func3Type func(...*any) any
type Func4Type func(byte, ...[]*complex128) bool

type ChanType chan struct{}
type Chan1Type <-chan struct{}
type Chan2Type chan<- struct{}

type UnsafePointer unsafe.Pointer

type Node struct {
	Int        int
	Int8       int8
	Int16      int16
	Int32      int32
	Int64      int64
	Uint       uint
	Uint8      uint8
	Uint16     uint16
	Uint32     uint32
	Uint64     uint64
	Float32    float32
	Float64    float64
	Complex64  complex64
	Complex128 complex128
	Bool1      bool
	Bool2      bool
	String     string

	Uintptr uintptr

	IntPtr        *int
	Int8Ptr       *int8
	Int16Ptr      *int16
	Int32Ptr      *int32
	Int64Ptr      *int64
	UintPtr       *uint
	Uint8Ptr      *uint8
	Uint16Ptr     *uint16
	Uint32Ptr     *uint32
	Uint64Ptr     *uint64
	Float32Ptr    *float32
	Float64Ptr    *float64
	Complex64Ptr  *complex64
	Complex128Ptr *complex128
	Bool1Ptr      *bool
	Bool2Ptr      *bool
	StringPtr     *string

	UintptrPtr *uintptr

	TypedInt        IntType
	TypedInt8       Int8Type
	TypedInt16      Int16Type
	TypedInt32      Int32Type
	TypedInt64      Int64Type
	TypedUint       UintType
	TypedUint8      Uint8Type
	TypedUint16     Uint16Type
	TypedUint32     Uint32Type
	TypedUint64     Uint64Type
	TypedFloat32    Float32Type
	TypedFloat64    Float64Type
	TypedComplex64  Complex64Type
	TypedComplex128 Complex128Type
	TypedBool1      Bool1Type
	TypedBool2      Bool2Type
	TypedString     StringType

	TypedUintptr UintptrType

	TypedIntPtr        IntPtrType
	TypedInt8Ptr       Int8PtrType
	TypedInt16Ptr      Int16PtrType
	TypedInt32Ptr      Int32PtrType
	TypedInt64Ptr      Int64PtrType
	TypedUintPtr       UintPtrType
	TypedUint8Ptr      Uint8PtrType
	TypedUint16Ptr     Uint16PtrType
	TypedUint32Ptr     Uint32PtrType
	TypedUint64Ptr     Uint64PtrType
	TypedFloat32Ptr    Float32PtrType
	TypedFloat64Ptr    Float64PtrType
	TypedComplex64Ptr  Complex64PtrType
	TypedComplex128Ptr Complex128PtrType
	TypedBool1Ptr      Bool1PtrType
	TypedBool2Ptr      Bool2PtrType
	TypedStringPtr     StringPtrType

	TypedUintptrPtr UintptrPtrType

	PtrTypedInt        *IntType
	PtrTypedInt8       *Int8Type
	PtrTypedInt16      *Int16Type
	PtrTypedInt32      *Int32Type
	PtrTypedInt64      *Int64Type
	PtrTypedUint       *UintType
	PtrTypedUint8      *Uint8Type
	PtrTypedUint16     *Uint16Type
	PtrTypedUint32     *Uint32Type
	PtrTypedUint64     *Uint64Type
	PtrTypedFloat32    *Float32Type
	PtrTypedFloat64    *Float64Type
	PtrTypedComplex64  *Complex64Type
	PtrTypedComplex128 *Complex128Type
	PtrTypedBool1      *Bool1Type
	PtrTypedBool2      *Bool2Type
	PtrTypedString     *StringType

	PtrTypedUintptr *UintptrType

	NilPointer *int

	Func  func()
	Func2 func(int) float64
	Func3 func(...*any) any
	Func4 func(byte, ...[]*complex128) bool

	FuncPtr  *func()
	Func2Ptr *func(int) float64
	Func3Ptr *func(...*any) any
	Func4Ptr *func(byte, ...[]*complex128) bool

	TypedFunc  FuncType
	TypedFunc2 Func2Type
	TypedFunc3 Func3Type
	TypedFunc4 Func4Type

	PtrTypedFunc  *FuncType
	PtrTypedFunc2 *Func2Type
	PtrTypedFunc3 *Func3Type
	PtrTypedFunc4 *Func4Type

	NilFunc func()

	Chan  chan struct{}
	Chan1 <-chan struct{}
	Chan2 chan<- struct{}

	ChanPtr  *chan struct{}
	Chan1Ptr *<-chan struct{}
	Chan2Ptr *chan<- struct{}

	TypedChan  ChanType
	TypedChan1 Chan1Type
	TypedChan2 Chan2Type

	PtrTypedChan  *ChanType
	PtrTypedChan1 *Chan1Type
	PtrTypedChan2 *Chan2Type

	BufferedChan chan struct{}
	NilChan      chan struct{}

	UnsafePointer1     unsafe.Pointer
	UnsafePointer2     *unsafe.Pointer
	NamedUnsafePointer UnsafePointer
}

type Inner struct {
	ID            int
	Notes         []string
	privateNote   string
	privateCustom any
}

type Outer struct {
	Name      string
	Timestamp time.Time
	Ptr       *Inner
	Inner     Inner
	Secret    string // unexported by convention
}

type Recursive struct {
	Self *Recursive
}

type InterfaceImpl struct{}

func (InterfaceImpl) Read(p []byte) (int, error) {
	return 0, nil
}

type Composite struct {
	Numbers    []int
	Names      [2]string
	DataMap    map[string]int
	Interface  any
	IO         io.Reader
	Err        error
	Func       func(string) string
	AnonStruct struct{ Value int }
	Recursive  *Recursive
}

type MyStringer struct{}

func (MyStringer) String() string {
	return "I'm a Stringer!"
}

type WithMethods struct{}

func (WithMethods) Hello() string  { return "hello" }
func (*WithMethods) World() string { return "world" }
func (WithMethods) Do(x int) int   { return x * 2 }
func (*WithMethods) Complex(x string) func(int) string {
	return func(i int) string {
		return "i"
	}
}

const ConstInt = 20

type Footester interface {
	Footest(int)
}

var MyNilFunction func(int) = nil

func MyHandler(name string, ctx context.Context) int {
	return 0
}

func MyFuncWithInterface(parameter Footester) error {
	return nil
}

func MyVariadic(myBoy any, values ...int) error {
	for _, val := range values {
		fmt.Printf("value: %d", val)
	}
	return nil
}

type NotSimpleType int

func (m *NotSimpleType) SimpleMethod(sayThis string) string {
	return fmt.Sprintf("Hi there, I am %d and you want me to say: %s", m, sayThis)
}

type InnerStruct struct {
	ID           int
	Notes        []string
	Whatever     interface{}
	WhateverElse any
}

type Methodist struct {
	Height               float32
	PrimitiveWithMethods NotSimpleType
}

func (m *Methodist) SomeMethod() {
	m.Height = 399
}

func (m *Methodist) String() string {
	return "I am a Methodist type struct and this is my fmt.Stringer representation!"
}

func (m Methodist) OtherMethod() error {
	fmt.Print(m.Height)
	return nil
}

func (m Methodist) MethodAny(x int, anything interface{}) error {
	fmt.Print(m.Height)
	return nil
}

type Ref struct {
	Self *Ref
}

type SmallStruct struct {
	ID           int
	privateName  string
	privateSlice []float32
	InterfaceSub Footester
	InterfaceAny any
}

type Everything struct {
	String            string
	Bool              bool
	Bool2             bool
	Int               int
	Float             float64
	Complex           complex128
	Time              time.Time
	Duration          time.Duration
	PtrString         *string
	PtrDuration       *time.Duration
	SliceInts         []int
	ArrayStrings      [2]string
	MapValues         map[string]int
	Nested            InnerStruct
	NestedPtr         *InnerStruct
	InterfaceAny      any
	InterfaceSpecific Footester
	Recursive         *Ref
	privateField      string
	privateStruct     InnerStruct
}

func (i InnerStruct) Footest(value int) {
	fmt.Println("heh")
}

type EmbeddedStruct struct {
	Height float32
	Weight float32
}

type SomeTypeWithInterfaces struct {
	VisibleField    int
	invisibleField  int
	FirstInterface  fmt.Stringer
	secondInterface Footester
	EmbeddedStruct
}

type CustomError struct {
	internal float64
}

func (e CustomError) Error() string {
	return "something strange happened in my custom setup"
}

type NotSimpleSlice []string

func (s NotSimpleSlice) CoolMethod() string {
	return "something"
}

type NotSimpleBool bool

func (b NotSimpleBool) CoolMethod() bool {
	return false
}

type NotSimpleChan chan<- int

func (ch NotSimpleChan) ChanMethod() int {
	return 1
}

type NotSimpleFunc func() string

func (f NotSimpleFunc) FuncMethod() int {
	return 22
}

type NotSimpleMap map[string]string

func (m NotSimpleMap) MapMethod() int {
	return 5
}

type ListNode struct {
	Value int
	Next  *ListNode
}

var (
	// Primitive
	booleanTrue         = true
	booleanFalse        = false
	intVal              = 42
	intZero             = 0
	uintVal      uint64 = 999
	floatVal            = 3.1415
	complexVal          = complex(1.2, 3.4)
	byteVal             = byte('A')
	runeVal             = '⚙'
	stringVal           = "Sample string"
	utf8String          = "č, š, ž, ě, ř, ť"
	emptyStr            = ""

	// Arrays and slices
	arrayVal           = [3]int{1, 2, 3}
	emptyArray         = [0]int{}
	sliceVal           = []float64{1.1, 2.2, 3.3}
	sliceExoticStrings = []string{"žluť", "kůň", "žhář"}
	nilSlice           []string
	emptySlice         = []string{}
	multiDimArray      = [2][2]int{{1, 2}, {3, 4}}
	matrix             = [5][5]int{
		{1, 2, 3, 4, 5},
		{0, 1, 2, 3, 4},
		{0, 0, 1, 2, 3},
		{0, 0, 0, 1, 2},
		{0, 0, 0, 0, 1},
	}

	longArray         [15]int
	longArrayPointers [15]*int

	// Maps
	mapVal   = map[string]int{"one": 1, "two": 2}
	emptyMap = map[int]bool{}
	nilMap   map[string]string
	mapAny   = map[any]any{"x": 1, 2: "y", complex(1, 2): true}

	// Structs
	inner     = Inner{ID: 1, Notes: []string{"note1", "note2"}, privateNote: "Only mine", privateCustom: 3}
	outer     = Outer{Name: "Test", Timestamp: time.Now(), Ptr: &inner, Inner: inner, Secret: "shhh"}
	composite = Composite{
		Numbers:    []int{10, 20, 30},
		Names:      [2]string{"foo", "bar"},
		DataMap:    map[string]int{"k1": 1},
		Interface:  "anything",
		IO:         InterfaceImpl{},
		Err:        errors.New("something went wrong"),
		Func:       func(s string) string { return "Hello, " + s },
		AnonStruct: struct{ Value int }{42},
		Recursive:  &Recursive{nil},
	}

	// Pointers
	ptrInt    = &intVal
	nilPtr    *Outer
	doublePtr = &ptrInt

	// Interfaces
	emptyInterface any
	typedInterface io.Reader    = InterfaceImpl{}
	stringerVal    fmt.Stringer = MyStringer{}

	// Functions
	funcVal      = func(a int) bool { return a > 0 }
	nilFunc      func()
	methodHolder = WithMethods{}

	// Channels
	chanSendOnly chan<- int = make(chan int, 1)
	chanRecvOnly <-chan int = make(chan int)
	chanBidirect chan int   = make(chan int, 2)
	chanClosed              = make(chan int)
)

func main() {

	node := Node{
		Int:        123,
		Int8:       -45,
		Int16:      6789,
		Int32:      -987,
		Int64:      3849876543247876432,
		Uint:       837,
		Uint8:      38,
		Uint16:     3847,
		Uint32:     9843,
		Uint64:     2834,
		Float32:    123.475,
		Float64:    -12345.09876,
		Complex64:  12.987i,
		Complex128: -473i,
		Bool1:      true,
		Bool2:      false,
		String:     "foo bar",

		Uintptr: 1234567890,

		TypedInt:        IntType(123),
		TypedInt8:       Int8Type(-45),
		TypedInt16:      Int16Type(6789),
		TypedInt32:      Int32Type(-987),
		TypedInt64:      Int64Type(3849876543247876432),
		TypedUint:       UintType(837),
		TypedUint8:      Uint8Type(38),
		TypedUint16:     Uint16Type(3847),
		TypedUint32:     Uint32Type(9843),
		TypedUint64:     Uint64Type(2834),
		TypedFloat32:    Float32Type(123.475),
		TypedFloat64:    Float64Type(-12345.09876),
		TypedComplex64:  Complex64Type(12.987i),
		TypedComplex128: Complex128Type(-473i),
		TypedBool1:      Bool1Type(true),
		TypedBool2:      Bool2Type(false),
		TypedString:     StringType("foo bar"),

		TypedUintptr: UintptrType(1234567890),

		UnsafePointer1:     nil,
		NamedUnsafePointer: nil,

		Chan:         make(chan struct{}),
		Chan1:        make(chan struct{}),
		Chan2:        make(chan struct{}),
		BufferedChan: make(chan struct{}, 255),
	}

	node.IntPtr = &node.Int
	node.Int8Ptr = &node.Int8
	node.Int16Ptr = &node.Int16
	node.Int32Ptr = &node.Int32
	node.Int64Ptr = &node.Int64
	node.UintPtr = &node.Uint
	node.Uint8Ptr = &node.Uint8
	node.Uint16Ptr = &node.Uint16
	node.Uint32Ptr = &node.Uint32
	node.Uint64Ptr = &node.Uint64
	node.Float32Ptr = &node.Float32
	node.Float64Ptr = &node.Float64
	node.Complex64Ptr = &node.Complex64
	node.Complex128Ptr = &node.Complex128
	node.Bool1Ptr = &node.Bool1
	node.Bool2Ptr = &node.Bool2
	node.StringPtr = &node.String

	node.UintptrPtr = &node.Uintptr

	node.TypedIntPtr = node.IntPtr
	node.TypedInt8Ptr = node.Int8Ptr
	node.TypedInt16Ptr = node.Int16Ptr
	node.TypedInt32Ptr = node.Int32Ptr
	node.TypedInt64Ptr = node.Int64Ptr
	node.TypedUintPtr = node.UintPtr
	node.TypedUint8Ptr = node.Uint8Ptr
	node.TypedUint16Ptr = node.Uint16Ptr
	node.TypedUint32Ptr = node.Uint32Ptr
	node.TypedUint64Ptr = node.Uint64Ptr
	node.TypedFloat32Ptr = node.Float32Ptr
	node.TypedFloat64Ptr = node.Float64Ptr
	node.TypedComplex64Ptr = node.Complex64Ptr
	node.TypedComplex128Ptr = node.Complex128Ptr
	node.TypedBool1Ptr = node.Bool1Ptr
	node.TypedBool2Ptr = node.Bool2Ptr
	node.TypedStringPtr = node.StringPtr

	node.TypedUintptrPtr = node.UintptrPtr

	node.PtrTypedInt = &node.TypedInt
	node.PtrTypedInt8 = &node.TypedInt8
	node.PtrTypedInt16 = &node.TypedInt16
	node.PtrTypedInt32 = &node.TypedInt32
	node.PtrTypedInt64 = &node.TypedInt64
	node.PtrTypedUint = &node.TypedUint
	node.PtrTypedUint8 = &node.TypedUint8
	node.PtrTypedUint16 = &node.TypedUint16
	node.PtrTypedUint32 = &node.TypedUint32
	node.PtrTypedUint64 = &node.TypedUint64
	node.PtrTypedFloat32 = &node.TypedFloat32
	node.PtrTypedFloat64 = &node.TypedFloat64
	node.PtrTypedComplex64 = &node.TypedComplex64
	node.PtrTypedComplex128 = &node.TypedComplex128
	node.PtrTypedBool1 = &node.TypedBool1
	node.PtrTypedBool2 = &node.TypedBool2
	node.PtrTypedString = &node.TypedString

	node.PtrTypedUintptr = &node.TypedUintptr

	node.Func = func() {}
	node.Func2 = func(int) float64 { return 0 }
	node.Func3 = func(...*any) any { return nil }
	node.Func4 = func(byte, ...[]*complex128) bool { return false }

	node.TypedFunc = func() {}
	node.TypedFunc2 = func(int) float64 { return 0 }
	node.TypedFunc3 = func(...*any) any { return nil }
	node.TypedFunc4 = func(byte, ...[]*complex128) bool { return false }

	node.FuncPtr = &node.Func
	node.Func2Ptr = &node.Func2
	node.Func3Ptr = &node.Func3
	node.Func4Ptr = &node.Func4
	node.PtrTypedFunc = &node.TypedFunc
	node.PtrTypedFunc2 = &node.TypedFunc2
	node.PtrTypedFunc3 = &node.TypedFunc3
	node.PtrTypedFunc4 = &node.TypedFunc4

	ch := make(chan struct{})
	var ch2 <-chan struct{} = ch
	var ch3 chan<- struct{} = ch

	tch := ChanType(ch)
	tch1 := Chan1Type(ch2)
	tch2 := Chan2Type(ch3)

	node.ChanPtr = &ch
	node.Chan1Ptr = &ch2
	node.Chan2Ptr = &ch3
	node.TypedChan = ch
	node.TypedChan1 = ch2
	node.TypedChan2 = ch3
	node.PtrTypedChan = &tch
	node.PtrTypedChan1 = &tch1
	node.PtrTypedChan2 = &tch2

	node.UnsafePointer2 = (*unsafe.Pointer)(unsafe.Pointer(&node))

	// Print all the primitive types
	fmt.Println("Primitive type list:")
	govar.Dump(node)

	for i := range 15 {
		longArray[i] = 9
		longArrayPointers[i] = &longArray[i]
	}

	chanSendOnly <- 1
	chanBidirect <- 2
	close(chanClosed)

	// Other - types, collections and constructions
	fmt.Println("Now lets print out all kinds of possible Go types, collections and constructions:")
	govar.Dump(booleanTrue, booleanFalse, intVal, intZero, uintVal, floatVal, complexVal,
		byteVal, runeVal, stringVal, utf8String, emptyStr,
		arrayVal, emptyArray, sliceVal, sliceExoticStrings, nilSlice, emptySlice, multiDimArray, matrix, longArray, longArrayPointers,
		mapVal, emptyMap, nilMap, mapAny,
		inner, outer, composite,
		ptrInt, nilPtr, doublePtr,
		emptyInterface, typedInterface, stringerVal,
		funcVal, nilFunc, methodHolder,
		chanSendOnly, chanRecvOnly, chanBidirect, chanClosed,
	)

	// And more stuff....
	fmt.Println("Now lets print some other stuff, maybe less common:")

	simpleMethodist := Methodist{Height: 111, PrimitiveWithMethods: 11}
	simpleString := "Sample text"
	simpleStringWithNewline := "Sometext before newline\nsome text after newline"
	hundredLetterString := "Code flows like water-clean, sharp, alive. Debug dreams, peek deep & craft truth from logic’s light."
	notSimpleType := NotSimpleType(10)
	notSimpleSlice := NotSimpleSlice{}
	notSimpleBool := NotSimpleBool(true)
	notSimpleChan := make(NotSimpleChan)
	notSimpleFunc := NotSimpleFunc(func() string { return "Hi! I am a func with a method on myself!" })
	notSimpleMap := NotSimpleMap(map[string]string{"hi": "there"})

	anonFunc := func(foo string) string {
		return strings.Join([]string{"prefix", foo, "postfix"}, "-")
	}

	emptySlice := []string{}

	now := time.Now()
	ptrStr := "Hello"
	dur := time.Minute * 20

	bigStruct := Everything{
		String:            "test",
		Bool:              true,
		Bool2:             false,
		Int:               11,
		Float:             1.1111,
		Complex:           2.17 + 5.3i,
		Time:              now,
		Duration:          dur,
		PtrString:         &ptrStr,
		PtrDuration:       &dur,
		SliceInts:         []int{1, 2, 3},
		ArrayStrings:      [2]string{"foo", "bar"},
		MapValues:         map[string]int{"a": 1, "b": 2},
		Nested:            InnerStruct{ID: 10, Notes: []string{"alpha", "beta"}, Whatever: "Something"},
		NestedPtr:         &InnerStruct{ID: 99, Notes: []string{"x", "y"}},
		InterfaceAny:      map[string]bool{"ok": true},
		InterfaceSpecific: InnerStruct{ID: 5, Notes: []string{"private"}, Whatever: 1, WhateverElse: 2},
		Recursive:         &Ref{},
		privateField:      "private value",
		privateStruct:     InnerStruct{ID: 5, Notes: []string{"private"}},
	}
	bigStruct.Recursive.Self = bigStruct.Recursive // cycle

	smallStruct := SmallStruct{
		ID:           77,
		privateName:  "Adi\nYogi",
		privateSlice: []float32{2.5, 3, 4.777},
		InterfaceSub: nil,
		InterfaceAny: nil,
	}

	matrix := [4][4]int{
		{1, 2, 3, 4},
		{0, 1, 2, 3},
		{0, 0, 1, 2},
		{0, 0, 0, 1},
	}

	var longArray [9]int
	var longArrayPointers [9]*int
	for i := range 9 {
		longArray[i] = 9
		longArrayPointers[i] = &longArray[i]
	}
	var anyArray [3]any = [3]any{"first", nil, [3]any{1, 5, 10}}
	var anySlice []any = []any{1, "neco", 5.16, nil, smallStruct}

	var stringAnyMap map[string]any = make(map[string]any)
	stringAnyMap["nothing"] = nil
	stringAnyMap["one"] = 1
	stringAnyMap["two"] = 2.0
	stringAnyMap["three"] = "three"
	stringAnyMap["four"] = [4]int{0, 2, 4, 6}
	stringAnyMap["five"] = 5 + 0i
	stringAnyMap["six"] = SmallStruct{ID: 10, InterfaceAny: 4.44}
	stringAnyMap["prevline\nnextline"] = nil

	var anyAnyMap map[any]any = make(map[any]any, 100)
	anyAnyMap["hello"] = 34
	anyAnyMap[3] = "some stringy thingy"
	anyAnyMap[3.14] = true
	anyAnyMap[false] = []int{4, 3, 2}
	anyAnyMap[0.867+0.5i] = "index by a complex number"

	TestError := errors.New("something wrong")
	NextError := fmt.Errorf("something went wrong with this random number: %d", rand.IntN(1000))
	testSndChan := make(chan<- int)
	testRcvChan := make(<-chan int)
	testBidirectionalChan := make(chan int)
	testSndBufferedChan := make(chan<- int, 5)

	sliceOfSmallStruct := []SmallStruct{{ID: 44}, {ID: 45}, {ID: 46}}
	fewRunes := "ŠROTÍŘ 🔨🔥"
	interestingRunes := "Ahoj, tohle je čeština a ta má zvláštní písmena, jako třeba: č, š, ž nebo dokonce ř! 🇨🇿🌍"

	type tinyStruct struct {
		x    int
		y    int
		name string
	}
	smallishStruct := tinyStruct{x: 5, y: 0, name: "Gřandalf"}

	sliceWithCap := make([]float32, 2, 10)
	sliceWithCap[0] = 2.5
	sliceWithCap[1] = 0.8

	strangeMap := map[[10]int]string{{1, 2, 3, 4, 5, 7}: "ahoj", {2, 1, 0, 1, 2, 100}: "hello"}
	strangeMap2 := map[tinyStruct]string{{x: 1}: "wow", {x: 2, y: 3}: "OMG"}

	randomByteArray := [6]byte{123, 34, 101, 114, 114, 111}

	coolByteArray := []byte("Hex dumping is a method of displaying binary data in a human-readable hexadecimal format.\n" +
		"\t It helps developers inspect memory, debug data structures, or analyze file contents.\n" +
		"\t Each byte is shown as a two-digit hex value, often alongside ASCII for clarity during low-level debugging.\n")

	type CustomMap[T int | float64, V string] map[T]V

	mapGenerics := make(CustomMap[int, string])
	mapGenerics[3] = "three"

	littleMap := map[int]int{1: 2, 2: 4, 4: 6}

	newObjectWithInterfaces := SomeTypeWithInterfaces{VisibleField: 1, invisibleField: 2, FirstInterface: &Methodist{}, secondInterface: InnerStruct{}}
	newObjectWithInterfaces.Height = 186
	newMapWithInterfaces := map[string]fmt.Stringer{"something": &Methodist{Height: 100}, "smthElse": &Methodist{Height: 10}}

	n1 := &ListNode{Value: 1}
	n2 := &ListNode{Value: 2, Next: n1}
	n1.Next = n2

	govar.Dump(net.Dialer{})
	govar.Dump(simpleString)
	govar.Dump(simpleStringWithNewline)
	govar.Dump(hundredLetterString)
	govar.Dump(notSimpleType)
	govar.Dump(notSimpleSlice)
	govar.Dump(notSimpleBool)
	govar.Dump(notSimpleChan)
	govar.Dump(notSimpleFunc)
	govar.Dump(notSimpleMap)
	govar.Dump(emptySlice)
	govar.Dump(simpleMethodist)
	govar.Dump(TestError)
	govar.Dump(NextError)
	govar.Dump(CustomError{internal: 55.4})
	govar.Dump(MyNilFunction)
	govar.Dump(anonFunc)
	govar.Dump(MyHandler)
	govar.Dump(MyFuncWithInterface)
	govar.Dump(MyVariadic)
	govar.Dump(littleMap)
	govar.Dump(smallStruct)
	govar.Dump(bigStruct)
	govar.Dump(matrix)
	govar.Dump(longArray)
	govar.Dump(longArrayPointers)
	govar.Dump(anyArray)
	govar.Dump(anySlice)
	govar.Dump(stringAnyMap)
	govar.Dump(anyAnyMap)
	govar.Dump(testSndChan)
	govar.Dump(testRcvChan)
	govar.Dump(testBidirectionalChan)
	govar.Dump(testSndBufferedChan)
	govar.Dump(sliceOfSmallStruct)
	govar.Dump(fewRunes)
	govar.Dump(interestingRunes)
	govar.Dump(smallishStruct)
	govar.Dump(sliceWithCap)
	govar.Dump(strangeMap)
	govar.Dump(strangeMap2)
	govar.Dump(randomByteArray)
	govar.Dump(coolByteArray)
	govar.Dump(mapGenerics)
	govar.Dump(newObjectWithInterfaces)
	govar.Dump(newMapWithInterfaces)
	govar.Dump(n1)
	govar.Dump(n2)

}
