![Govar - GO variable dumper logo.](/assets/social_media_card.png)

### **Stop guessing, start seeing.**

`govar` is a modern Go variable dumper that helps you see what's really going on inside your data structures. It was built from the ground up to solve the biggest problem with existing dumpers: visualizing pointers. By assigning stable IDs to memory locations, `govar` is the only tool that can **visually map out complex object graphs, shared pointers, and even cycles.**

Beyond its groundbreaking reference tracking, it also provides beautifully formatted and colorful output for all of Go's built-in types, perfectly untangling nested structs, slices, and maps. Whether you need to dump to your terminal, a string, or a full HTML output for web UIs, `govar` provides a clear and insightful view every time.

\[\>\] Designed for **debugging**, **exploration**, **documentation**, and **interface analysis**.

* ✅ The most readable Go dumper output, period.
* ✅ **Groundbreaking ID/Back-Reference system** to visualize pointers.
* ✅ Goroutine safe and covered with extensive tests.
* ✅ Type & interface introspection tools.
* ✅ Well-documented and easy to customize.


Whether you're debugging, documenting, or just staring into the void of your own data structures — `govar` is here to make sense of it all.

## **🤔 Why use govar?**

`fmt.Printf("%+v", x)` is fine... until it isn't.

Ever felt like you're manually traversing a linked list in your terminal? Or trying to figure out if two pointers in a massive struct point to the same thing? That’s where `govar` shines.

It doesn’t just dump your data; it decodes it. It turns cryptic jungles of pointers, interfaces, and nested maps into a clean, readable, and insightful map of your program's state. With colors, types, method sets, and a reference tracking system that feels like a superpower.

<p align="center">
  <strong>A Glimpse of govar's Power</strong>
</p>

| Standard Data Structure | Complex Pointer Graph | Compact "No Types" |
| :---: | :---: | :---: |
| [![Standard Data Demo](/assets/demo_struct_preview.png)](/assets/demo_struct.png) | [![Complex Pointer Demo](/assets/demo_references_preview.png)](/assets/demo_references.png) | [![Compact No Types Demo](/assets/demo_no_types.png)](/assets/demo_no_types.png) |
| *Click image to see the complete output.* | *Click image to see the complete output.* | *Click image to see the complete output.* |

## **✨ Features at a Glance**

| Feature | Description |
| :---- | :---- |
| 🎨 **Pretty‑prints any Go value** | Nested structs, slices, pointers, maps, funcs, interfaces, channels. |
| 🔗 **Advanced ID/Back-Ref System** | The *only* Go dumper that assigns stable IDs (`&1`, `&2`...) to values and prints back-references (`↩︎ &1`) for pointers. Instantly visualize cycles, shared data, and complex object graphs. |
| 🗣️ **Stringer & Error Aware** | Automatically dumps values using their `fmt.Stringer` or `error` interface description for clearer output, unless disabled. |
| 🌈 **Multiple Output Options** | Colorized ANSI for your terminal, raw text for logs, or full HTML for UIs. |
| 🔎 **Rich Meta Information** | Type hints, interface markers (`⧉`), field visibility (`⯀`, `🞏`), method types (`⦿`), size, capacity, rune length, and more. |
| 🧾 **Formatted Hex Dumps** | Beautifully formatted hexdumps for []byte, []uint8, and similar byte slices that are actually easy to read. |
| 🧠 **Code Introspection (who)** | Find type → interface and interface → type relationships in your codebase without guesswork. |
| 💾 **Dump Anywhere** | To ***stdout***, any ***io.Writer***, a ***string***, or an ***HTML string***. |
| 🧰 **Highly Customizable** | Use `govar.DumperConfig` to control everything from indentation and depth to colors and reference tracking. |

## **🚀 Install**

```bash
go get github.com/janvaclavik/govar
```

## **🛠 Quickstart**

The govar package provides simple, top-level functions for immediate use.

```go
package main

import (
	"github.com/janvaclavik/govar"
)

func main() {
	// Dump to stdout, with types, meta, and colors
	govar.Dump(someVarToInspect1, someVarToInspect2)

	// Dump values only (colored, but no extras)
	govar.DumpValues(someVarToInspect1, someVarToInspect2)

	// Dump to a string
	str := govar.Sdump(someVarToInspect1)

	// Dump to an io.Writer (e.g., a file or buffer)
	govar.Fdump(someIOWriter, someVarToInspect1)

	// Dump to an HTML string
	html := govar.SdumpHTML(someVarToInspect1)

	// The classic "print and die" for quick debugging
	govar.Die(someVarToInspect1)
}
```

## **🔗 Untangle Your Pointers**

govar's killer feature is its ability to track and visualize pointers.

**Without govar**, a simple cyclic struct is impossible to debug with fmt:

```go
type Person struct {
    Name string
    Loves *Person
}
alice := &Person{Name: "Alice"}
bob := &Person{Name: "Bob"}
alice.Loves = bob
bob.Loves = alice // Oh no, a cycle!

// fmt.Printf just gives you an endless, useless loop...
// &{Name:Alice Loves:0x...} &{Name:Bob Loves:0x...} &{Name:Alice Loves:0x...} ...
```

**With govar**, the relationship becomes instantly clear:

```go
govar.Dump(alice, bob)
```

**Output:**

```
*main.Person => &1 {
   ⯀ Name   string       => "Alice"
   ⯀ Loves  *main.Person => ↩︎ &2
}

*main.Person => &2 {
   ⯀ Name   string       => "Bob"
   ⯀ Loves  *main.Person => ↩︎ &1
}
```

* &1 **(ID):** govar saw the "Alice" struct and assigned it the ID &1.
* ↩︎ &2 **(Back-Reference):** This clearly shows that *alice.Loves* points back to the struct that was assigned the ID &2 (Bob). The cycle is immediately obvious.

This works across multiple variables, nested fields, and complex data structures.

## **⚙️ Custom Dumper**

Need more control? Use `govar.NewDumper` with a custom configuration.

```go
import (
	"github.com/janvaclavik/govar"
)

func main() {
	myCfg := govar.DumperConfig{
		IndentWidth:         3,       // Indentation step
		MaxDepth:            15,      // Nesting level limit
		MaxItems:            100,     // Max elements in a collection before truncating
		MaxStringLen:        10000,   // The limit for string dumping
		MaxInlineLength:     80,      // The limit for inline value rendering
		ShowTypes:           true,    // Shows extra type info if true
		UseColors:           true,    // Plain text if false
		TrackReferences:     true,    // Set to false to disable the ID/back-ref system
		EmbedTypeMethods:    true,    // Shows implemented methods on any type
		ShowMetaInformation: true,    // Shows sizes, capacities, "rune length", etc.
		ShowHexdump:         true,    // Shows classic hexdump on byte[] or uint8[]
		IgnoreStringer:      false,   // Ignores fmt.Stringer/error formatting if true
	}

	d := govar.NewDumper(myCfg)

	// Now you can dump data with full control
	d.Dump(myData1, myData2)
}
```

## **🔍 The "Who" Introspection Helpers**

Ever wonder which of your structs implement io.Writer, or what interfaces a specific type satisfies? The govar/who subpackage is a static analysis tool that answers these questions, helping you understand your codebase's type and interface relationships without writing complex reflection code.

```go
package main

import (
	"github.com/janvaclavik/govar/who"
)

func main() {
  // Which types in my code implement this interface?
  types := who.Implements("myrepo/mypkg.SomeInterface")

  // Which interfaces in my code are implemented by this type?
  interfaces := who.Interfaces("myrepo/mypkg.MyType")

  // Which external interfaces (stdlib, etc) are implemented by this type?
  externals := who.InterfacesExt("myrepo/mypkg.MyType")
}
```

### **🧭 Summary of who functions**

| Function | Description |
| :---- | :---- |
| `who.Implements()` | Returns types in your codebase that implement a given interface. |
| `who.Interfaces()` | Lists interfaces in your codebase that a given type implements. |
| `who.InterfacesExt()` | Lists interfaces from stdlib and imported packages a given type satisfies. |

## **🧩 License**

MIT © [janvaclavik](https://github.com/janvaclavik)

## **🙏 Inspired by**

* [davecgh/go-spew](https://github.com/davecgh/go-spew)
* [yassinebenaid/godump](https://github.com/yassinebenaid/godump)
* [nette/tracy](https://github.com/nette/tracy) *(PHP's dump() inspiration)*
* [laravel/laravel](https://github.com/laravel/laravel) *(another PHP's dump() inspiration)*
* [pprint](https://docs.python.org/3/library/pprint.html) *(pprint — Python pretty printer)*

## **📇 Author**

Made with 🍵 and reflective thought by [Jan Vaclavik](https://github.com/janvaclavik)