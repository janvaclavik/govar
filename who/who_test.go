package who

import (
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func mustWriteFile(t *testing.T, root, relPath, content string) {
	fullPath := filepath.Join(root, relPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
}

func TestWhoImplements(t *testing.T) {
	tmpDir := t.TempDir()

	// --- Setup go.mod ---
	goMod := `module testmod

go 1.20
`

	// --- iface/iface.go ---
	ifaceCode := `package iface

type MyInterface interface {
	Foo()
}
`

	// --- impl/impl.go ---
	implCode := `package impl

type MyType struct{}

func (MyType) Foo() {}
`

	// --- Write files ---
	mustWriteFile(t, tmpDir, "go.mod", goMod)
	mustWriteFile(t, tmpDir, "iface/iface.go", ifaceCode)
	mustWriteFile(t, tmpDir, "impl/impl.go", implCode)

	// --- Save old working dir ---
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)

	// --- Change into temp module directory ---
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	// --- Call who.Implements ---
	results, err := Implements("testmod/iface.MyInterface")
	if err != nil {
		t.Fatalf("Implements error: %v", err)
	}
	t.Logf("Found implementers: %v", results)

	want := "testmod/impl.MyType"
	if !slices.Contains(results, want) {
		t.Errorf("Expected to find implementation: %s", want)
	}
}

func TestInterfaces(t *testing.T) {
	tmpDir := t.TempDir()

	// --- Setup module ---
	goMod := `module testmod

go 1.20
`

	// --- iface/iface.go ---
	ifaceCode := `package iface

type MyInterface interface {
	Foo()
}
`

	// --- impl/impl.go ---
	implCode := `package impl

type MyType struct{}

func (MyType) Foo() {}
`

	// --- Write files ---
	mustWriteFile(t, tmpDir, "go.mod", goMod)
	mustWriteFile(t, tmpDir, "iface/iface.go", ifaceCode)
	mustWriteFile(t, tmpDir, "impl/impl.go", implCode)

	// --- Save old working dir and chdir into test module ---
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	// --- Call who.Interfaces ---
	results, err := Interfaces("testmod/impl.MyType")
	if err != nil {
		t.Fatalf("Interfaces error: %v", err)
	}
	t.Logf("Found interfaces: %v", results)

	want := "testmod/iface.MyInterface"
	if !slices.Contains(results, want) {
		t.Errorf("Expected to find implemented interface: %s", want)
	}
}

func TestInterfacesExt(t *testing.T) {
	tmpDir := t.TempDir()

	goMod := `module testmod

go 1.20
`

	// --- impl/impl.go ---
	implCode := `package impl

import "fmt"

type MyType struct{}

func (MyType) String() string {
	return "hello"
}
`

	mustWriteFile(t, tmpDir, "go.mod", goMod)
	mustWriteFile(t, tmpDir, "impl/impl.go", implCode)

	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	results, err := InterfacesExt("testmod/impl.MyType")
	if err != nil {
		t.Fatalf("InterfacesExt error: %v", err)
	}
	t.Logf("Found external interfaces: %v", results)

	want := "fmt.Stringer"
	if !slices.Contains(results, want) {
		t.Errorf("Expected to find external interface %s, got: %v", want, results)
	}
}

func TestSplitTypeName(t *testing.T) {
	tests := []struct {
		input       string
		wantPkgPath string
		wantType    string
		wantErr     bool
	}{
		{"mypkg.MyType", "mypkg", "MyType", false},
		{"net/http.Handler", "net/http", "Handler", false},
		{"invalidname", "", "", true},
		{".OnlyType", "", "OnlyType", false},
		{"OnlyPkg.", "OnlyPkg", "", false},
	}

	for _, tt := range tests {
		pkgPath, typeName, err := splitTypeName(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("splitTypeName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if pkgPath != tt.wantPkgPath || typeName != tt.wantType {
			t.Errorf("splitTypeName(%q) = (%q, %q), want (%q, %q)",
				tt.input, pkgPath, typeName, tt.wantPkgPath, tt.wantType)
		}
	}
}

func TestIsConcreteNamedType(t *testing.T) {
	// Create a named concrete type (struct)
	obj := types.NewTypeName(token.NoPos, nil, "MyStruct", types.NewNamed(
		types.NewTypeName(token.NoPos, nil, "MyStruct", nil),
		types.NewStruct(nil, nil),
		nil,
	))

	// Create an interface type
	iface := types.NewInterfaceType(nil, nil)
	ifaceObj := types.NewTypeName(token.NoPos, nil, "MyInterface", iface)

	tests := []struct {
		name string
		obj  types.Object
		want bool
	}{
		{"StructType", obj, true},
		{"InterfaceType", ifaceObj, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isConcreteNamedType(tt.obj)
			if got != tt.want {
				t.Errorf("isConcreteNamedType() = %v, want %v", got, tt.want)
			}
		})
	}
}
