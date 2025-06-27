package main

import (
	"errors"
	"fmt"
	"io"
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

	// All primitive types
	govar.Dump(node)

	// Other
	for i := range 15 {
		longArray[i] = 9
		longArrayPointers[i] = &longArray[i]
	}

	chanSendOnly <- 1
	chanBidirect <- 2
	close(chanClosed)

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

}
