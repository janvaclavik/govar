# govar

`govar` is a handy Go object inspector and variable dumper. It provides a well-arranged, pretty-printed peeks into selected constants and variables in your program. You can also use it as a friendly assistant when learning about Go data types, structures and interface. `govar` has no external dependencies and is go-routine friendly! It focuses on:

- **Readable, styled output** ( *stdout* | *string* | *io.Writer* | *HTML* )
- **Complete type and value information for any constant or variable**, including structured data and functions
- **Handy helper tools** for discovering and inspecting Go types and interfaces in current codebase

Whether you're debugging, documenting, learning or building dev tools — `govar` makes Go data easier to understand and explore.

## ❓ OK, but why should I pick this dumper?

There are a few of these tools already, but whenever I used them, I always found myself missing some important feature. There is a feature list down below, so go check it out (and for more on this, you can check the [comparison with other tools](https://github.com/janvaclavik/govar#comparison) section). The aim of `govar` is to be useful for inspection, studying and also testing purposes.

## ✨ Features

- 📦 Pretty-print any Go value (supporting structured/nested types, pointers, arrays, slices, maps, funcs, interfaces, channels, etc.)
- ✅ time.Time (nicely formatted)
- 🧠 Struct field inspection with visibility markers (`+`, `-`)
- 🧠 Variadic function param signature
- 🧠 Displays size and capacity of arrays, slices and maps
- 🪄 Control character escaping (`\n`, `\t`, etc.)
- 🎨 ANSI color or HTML output (or even plain, uncolored output)
- 🔄 Cycle-safe reference tracking
- 🧠 Smart function signature rendering: `func(string, context.Context) error`
- 🎨 Optional HTML output with structured and styled formatting
- 🔎 Discover all types in your project that implement a given interface
- ⚙️ Simple API and clean defaults for drop-in debugging or custom tooling

## 🚀 Getting Started

```bash
go get github.com/janvaclavik/govar
```

## 🚀 Usage (dumper)

```go
package main

import (
	"github.com/janvaclavik/govar"
)

func main() {
	// Dump straight to stdout
	// (with colors ON, type info ON, meta-hints ON)
	govar.Dump(someVarToInspect1, someVarToInspect2, ...)

	// Dump straight to stdout, values only
	// (with colors OFF, type info OFF, meta-hints OFF)
	govar.DumpValues(someVarToInspect1, someVarToInspect2, ...)

	// Dump into a string
	// (with colors ON, type info ON, meta-hints ON)
	outputStr := govar.Sdump(someVarToInspect1, someVarToInspect2, ...)

	// Dump into a string
	// (with colors OFF, type info OFF, meta-hints OFF)
	outputStr2 := govar.SdumpValues(someVarToInspect1, someVarToInspect2, ...)

	// Write to any io.Writer
	// (with colors ON, type info ON, meta-hints ON)
	govar.Fdump(someIOWriter, someVarToInspect1, someVarToInspect2, ...)

	// Write to any io.Writer, values only
	// (with colors OFF, type info OFF, meta-hints OFF)
	govar.FdumpValues(someIOWriter, someVarToInspect1, someVarToInspect2, ...)

	// HTML for web output inside a <pre> block
	// (with colors ON, type info ON, meta-hints ON)
	html := govar.HTMLdump(someVarToInspect1, someVarToInspect2, ...)

	// HTML output inside a <pre> block, values only
	// (with colors OFF, type info OFF, meta-hints OFF)
	html2 := govar.HTMLdumpValues(someVarToInspect1, someVarToInspect2, ...)

	// Dump to stdout and die (exit the program right after that)
	govar.Die(someVarToInspect1, someVarToInspect2, ...)

	// Returns a sorted slice of types in your codebase that implement a given interface
	// (must be a full interface type name: github.com/some_repo/some_pkg/<some_subpkg/>.SomeInterface)
	// (or fmt.Stringer, io.Reader, ...)
	sliceOfTypes := introspect.FindImplementors("github.com/myrepo/mypkg/main.SomeInterface1")

	// Returns a sorted slice of interfaces in current project codebase that are implemented by a given type
	// (must be a full type name: github.com/some_repo/some_pkg/<some_subpkg/>.MyType)
	sliceOfInterfaces := introspect.FindInterfaces("github.com/myrepo/mypkg/main.MyType")

	// Returns a sorted slice of interfaces in Go std lib that are implemented by a given type
	// (must be a full type name: github.com/some_repo/some_pkg/<some_subpkg/>.MyType)
	sliceOfInterfaces := introspect.FindInterfacesStd("github.com/myrepo/mypkg/main.MyType")

}
```

## 🚀 Usage (introspect)

```go
package main

import (
	"github.com/janvaclavik/govar/introspect"
)

func main() {
	// Returns a sorted slice of types in your codebase that implement a given interface
	// (must be a full interface name: github.com/some_repo/some_pkg/<some_subpkg/>.SomeInterface)
	// (or fmt.Stringer, io.Reader, ...)
	sliceOfTypes := introspect.FindImplementors("github.com/myrepo/mypkg/main.SomeInterface1")

	// Returns a slice of interfaces in the project codebase that are implemented by a given type
	// (must be a full type name: github.com/some_repo/some_pkg/<some_subpkg/>.MyType)
	sliceOfInterfaces := introspect.FindInterfaces("github.com/myrepo/mypkg/main.MyType")

	// Returns a slice of interfaces in Go std lib that are implemented by a given type
	// (must be a full type name: github.com/some_repo/some_pkg/<some_subpkg/>.MyType)
	sliceOfInterfaces := introspect.FindInterfacesStd("github.com/myrepo/mypkg/main.MyType")
}
```

## ⚖️ Comparison with other tools

TODO

## 🧩 License

MIT © [janvaclavik](https://github.com/janvaclavik)

## 🌐 Inspired by
- [davecgh/go-spew](https://github.com/davecgh/go-spew)
- [yassinebenaid/godump](https://github.com/yassinebenaid/godump)
- [goforj/godump](https://github.com/goforj/godump)
- [nette/tracy](https://github.com/nette/tracy) (has dump() for *PHP*)
- [laravel/laravel](https://github.com/laravel/laravel) (has dump() for *PHP*)
- [pprint](https://docs.python.org/3/library/pprint.html) (pprint — Data pretty printer for *Python*)

## 📇 Author

Created by [Jan Vaclavik](https://github.com/janvaclavik)

