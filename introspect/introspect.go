package introspect

import (
	"fmt"
	"go/types"
	"slices"
	"strings"

	"golang.org/x/tools/go/packages"
)

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

func FindImplementors(interfaceFullName string) ([]string, error) {
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

// FindInterfaces finds interfaces in the current project that the given type implements.
func FindInterfaces(typeFullName string) ([]string, error) {
	return findInterfaces(typeFullName, false)
}

// FindInterfaces finds interfaces in the current project that the given type implements.
func FindInterfacesStd(typeFullName string) ([]string, error) {

	// First, find all matched interfaces, including stdlib
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
	listStd := []string{}
	for _, iface := range listAll {
		// Adds an interface to the Std list only if it is not in project codebase list already
		if _, ok := codebaseSet[iface]; !ok {
			listStd = append(listStd, iface)
		}
	}

	return listStd, nil
}

func findInterfaces(typeFullName string, includeStd bool) ([]string, error) {
	typePkgPath, typeName, err := splitTypeName(typeFullName)
	// fmt.Println("(target) type name: ", typeName)
	if err != nil {
		return nil, err
	}

	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports | packages.NeedDeps | packages.NeedSyntax,
	}

	loadPattern := "./..."
	if includeStd {
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

// splitTypeName splits "somepkg.MyType" into "somepkg" and "MyType"
func splitTypeName(full string) (pkgPath, typeName string, err error) {
	lastDot := strings.LastIndex(full, ".")
	if lastDot < 0 {
		return "", "", fmt.Errorf("invalid type name: %s", full)
	}
	pkgPath = full[:lastDot]
	typeName = full[lastDot+1:]
	return pkgPath, typeName, nil
}
