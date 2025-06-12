# govar

`govar` is a handy Go object inspector and variable dumper. It provides a well-arranged, pretty-printed peeks into selected constants and variables in your program. You can also use it as a friendly assistant when learning about Go data types and structures, and it has no external dependencies! It focuses on:

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

## 🚀 Usage

```go
package main

import (
	"github.com/janvaclavik/govar"
)

func main() {
	// Dump straight to stdout
	govar.Dump(someVarToInspect1, someVarToInspect2, ...)

	// Dump into a string
	outputStr := govar.Sdump(someVarToInspect1, someVarToInspect2, ...)

	// Write to any io.Writer (e.g. file, buffer, logger)
	govar.Fdump(someIOWriter, someVarToInspect1, someVarToInspect2, ...)

	// HTML for web output inside a <pre> block
	html := govar.HTMLdump(someVarToInspect1, someVarToInspect2, ...)

	// Dump to stdout and die (exit the program right after that)
	govar.Die(someVarToInspect1, someVarToInspect2, ...)

	// Returns a list of types in your codebase that implement a given interface
	sliceOfTypes := govar.FindImplementors("somepkg.SomeInterface1")

	// Returns a list of interfaces in your codebase that are implemented by a given type
	sliceOfInterfaces := govar.FindInterfaces("somepkg.SomeType1")

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

## 📇 Author

Created by [Jan Vaclavik](https://github.com/janvaclavik)

