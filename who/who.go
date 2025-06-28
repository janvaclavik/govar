// Package who provides utilities for analyzing Go packages to determine
// which types implement specific interfaces, and which interfaces are
// implemented by specific types within a project or across all dependencies.
package who

import (
	"fmt"
	"go/types"
	"slices"
	"strings"

	"golang.org/x/tools/go/packages"
)

// isConcreteNamedType checks whether the given object is a concrete (non-interface) named type.
func isConcreteNamedType(obj types.Object) bool {
	// Must be a type declaration
	typeName, ok := obj.(*types.TypeName)
	if !ok {
		return false
	}

	// Get the actual type info
	named, ok := typeName.Type().(*types.Named)
	if !ok {
		return false
	}

	// Check if the underlying type is NOT an interface
	_, isInterface := named.Underlying().(*types.Interface)
	return !isInterface
}

// Implements finds all concrete types in the current module and its dependencies
// that implement the interface identified by the fully-qualified name
// (e.g. "net/http.Handler").
//
// Returns a sorted list of fully-qualified type names like "mypkg.MyType".
// Returns an error if the interface cannot be resolved or packages fail to load.
func Implements(interfaceFullName string) ([]string, error) {
	// 1. Parse "pkgpath.InterfaceName"
	typePkgPath, typeName, err := splitTypeName(interfaceFullName)
	if err != nil {
		return nil, err
	}

	// 2. Load all packages
	cfg := &packages.Config{Mode: packages.LoadTypes | packages.LoadSyntax | packages.NeedDeps}
	// pkgs, err := packages.Load(cfg, "./...")
	pkgs, err := packages.Load(cfg, "all")
	if err != nil {
		return nil, err
	}

	// 3. Locate the target interface object
	var targetIface *types.Interface
	for _, pkg := range pkgs {
		if pkg.PkgPath != typePkgPath {
			continue
		}
		obj := pkg.Types.Scope().Lookup(typeName)
		if obj == nil {
			continue
		}
		iface, ok := obj.Type().Underlying().(*types.Interface)
		if ok {
			targetIface = iface
			break
		}
	}

	if targetIface == nil {
		return nil, fmt.Errorf("interface not found: %s", interfaceFullName)
	}

	// 4. Iterate over all named types and check if they implement the interface
	var result []string
	for _, pkg := range pkgs {
		scope := pkg.Types.Scope()
		for _, name := range scope.Names() {
			obj := scope.Lookup(name)
			if obj == nil {
				continue
			}

			named, ok := obj.Type().(*types.Named)
			if !ok || !isConcreteNamedType(obj) {
				continue
			}

			// Check both T and *T
			if types.Implements(named, targetIface) || types.Implements(types.NewPointer(named), targetIface) {
				result = append(result, fmt.Sprintf("%s.%s", pkg.PkgPath, obj.Name()))
			}
		}
	}

	slices.Sort(result)

	return result, nil
}

// Interfaces finds all project-local interfaces that are implemented by the given type,
// specified by fully-qualified name (e.g. "mypkg.MyStruct").
// This does not include interfaces from the standard library or external modules.
func Interfaces(typeFullName string) ([]string, error) {
	return findInterfaces(typeFullName, false)
}

// InterfacesExt returns interfaces implemented by the given type
// from the standard library and external dependencies only (excluding project-local interfaces).
// It excludes interfaces found in the current project (i.e. those returned by Interfaces()).
func InterfacesExt(typeFullName string) ([]string, error) {

	// First, find all matched interfaces, including stdlib and external imports
	listAll, err := findInterfaces(typeFullName, true)
	if err != nil {
		return nil, err
	}

	// Second, codebase (project-defined) interfaces only
	listCodebase, err := findInterfaces(typeFullName, false)
	if err != nil {
		return nil, err
	}

	// Make a map (name => empty struct) from the codebase interfaces
	codebaseSet := map[string]struct{}{}
	for _, iface := range listCodebase {
		codebaseSet[iface] = struct{}{}
	}

	// Init the result list
	listExt := []string{}
	for _, iface := range listAll {
		// Adds an interface to the Ext list only if it is not in project codebase list already
		if _, ok := codebaseSet[iface]; !ok {
			listExt = append(listExt, iface)
		}
	}

	return listExt, nil
}

// findInterfaces returns all interfaces (optionally including external ones) that the
// specified type implements, based on its fully-qualified name.
// This is a shared internal helper used by Interfaces and InterfacesExt.
func findInterfaces(typeFullName string, includeExt bool) ([]string, error) {
	typePkgPath, typeName, err := splitTypeName(typeFullName)
	// fmt.Println("(target) type name: ", typeName)
	if err != nil {
		return nil, err
	}

	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports | packages.NeedDeps | packages.NeedSyntax,
	}

	loadPattern := "./..."
	if includeExt {
		loadPattern = "all"
	}

	pkgs, err := packages.Load(cfg, loadPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to load packages: %w", err)
	}

	var targetType types.Type

	// The result var
	var implementedInterfaces []string

	// Step 1: Find the target type.
	for _, pkg := range pkgs {
		obj := pkg.Types.Scope().Lookup(typeName)

		if obj == nil {
			continue
		}
		targetType = obj.Type()
		break
	}
	if targetType == nil {
		return nil, fmt.Errorf("type %s not found in package %s", typeName, typePkgPath)
	}

	for _, pkg := range pkgs {
		scope := pkg.Types.Scope()
		for _, name := range scope.Names() {
			obj := scope.Lookup(name)
			if obj == nil {
				continue
			}

			if iface, ok := obj.Type().Underlying().(*types.Interface); ok {
				// Check both T and *T
				if types.Implements(targetType, iface) || types.Implements(types.NewPointer(targetType), iface) {
					var ifacePkgPath string
					if obj.Pkg() != nil {
						ifacePkgPath = obj.Pkg().Path()
					} else {
						ifacePkgPath = "builtin"
					}
					fullIfaceName := fmt.Sprintf("%s.%s", ifacePkgPath, obj.Name())
					implementedInterfaces = append(implementedInterfaces, fullIfaceName)
				}
			}
		}
	}

	slices.Sort(implementedInterfaces)

	return implementedInterfaces, nil
}

// splitTypeName splits a fully-qualified type string such as "mypkg.MyType"
// into its package path ("mypkg") and type name ("MyType") components.
//
// Returns an error if the format is invalid (e.g. missing a dot separator).
func splitTypeName(full string) (pkgPath, typeName string, err error) {
	lastDot := strings.LastIndex(full, ".")
	if lastDot < 0 {
		return "", "", fmt.Errorf("invalid type name: %s", full)
	}
	pkgPath = full[:lastDot]
	typeName = full[lastDot+1:]
	return pkgPath, typeName, nil
}
