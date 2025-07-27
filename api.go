// Package govar provides a powerful and highly configurable pretty-printer for Go
// data structures. This file defines the convenient top-level API functions
// (e.g., Dump, Sdump, Die) which wrap the core Dumper functionality with
// sensible default configurations.
package govar

import "io"

// DefaultConfig provides a standard, full-featured dumper configuration.
// It enables types, metadata, colors, reference tracking, and method embedding.
var DefaultConfig = DumperConfig{
	IndentWidth:         3,
	MaxDepth:            15,
	MaxItems:            150,
	MaxStringLen:        10000,
	MaxInlineLength:     80,
	ShowTypes:           true,
	UseColors:           true,
	TrackReferences:     true,
	EmbedTypeMethods:    true,
	ShowMetaInformation: true,
	ShowHexdump:         true,
	IgnoreStringer:      false,
	HTMLtagToken:        "span",
	HTMLtagSection:      "pre",
}

// SimpleConfig provides a simplified dumper configuration.
// It disables types, metadata, and method embedding for a more compact output.
var SimpleConfig = DumperConfig{
	IndentWidth:         3,
	MaxDepth:            15,
	MaxItems:            150,
	MaxStringLen:        10000,
	MaxInlineLength:     80,
	ShowTypes:           false,
	UseColors:           true,
	TrackReferences:     true,
	EmbedTypeMethods:    false,
	ShowMetaInformation: false,
	ShowHexdump:         true,
	IgnoreStringer:      false,
	HTMLtagToken:        "span",
	HTMLtagSection:      "pre",
}

// Die dumps the provided values using the DefaultConfig and terminates the program
// with os.Exit(1). It is a convenient shortcut for `govar.Dump(...)` followed by an exit.
func Die(values ...any) {
	d := NewDumper(DefaultConfig)
	d.Die(values...)
}

// Dump prints the given values to stdout using the DefaultConfig.
// It provides a rich, colored output with full type and metadata information.
func Dump(values ...any) {
	d := NewDumper(DefaultConfig)
	d.Dump(values...)
}

// DumpNoColors prints the given values to stdout with full formatting, but with colors disabled.
func DumpNoColors(values ...any) {
	cfg := DefaultConfig
	cfg.UseColors = false
	d := NewDumper(cfg)
	d.Dump(values...)
}

// DumpValues prints the values to stdout using the SimpleConfig.
// This produces a more compact output, omitting types, metadata, and methods.
func DumpValues(values ...any) {
	d := NewDumper(SimpleConfig)
	d.Dump(values...)
}

// Fdump writes the formatted output of the given values to the provided io.Writer
// using the DefaultConfig.
func Fdump(w io.Writer, values ...any) {
	d := NewDumper(DefaultConfig)
	d.Fdump(w, values...)
}

// FdumpNoColors writes formatted output to the given writer, with all formatting
// enabled except for colored output.
func FdumpNoColors(w io.Writer, values ...any) {
	cfg := DefaultConfig
	cfg.UseColors = false
	d := NewDumper(cfg)
	d.Fdump(w, values...)
}

// FdumpValues writes simplified formatted output to the writer using the SimpleConfig.
func FdumpValues(w io.Writer, values ...any) {
	d := NewDumper(SimpleConfig)
	d.Fdump(w, values...)
}

// Sdump returns the full-formatted string representation of the given values
// using the DefaultConfig.
func Sdump(values ...any) string {
	d := NewDumper(DefaultConfig)
	return d.Sdump(values...)
}

// SdumpNoColors returns the formatted string representation with all features enabled
// except for colored output.
func SdumpNoColors(values ...any) string {
	cfg := DefaultConfig
	cfg.UseColors = false
	d := NewDumper(cfg)
	return d.Sdump(values...)
}

// SdumpValues returns the simplified string representation of the given values
// using the SimpleConfig.
func SdumpValues(values ...any) string {
	d := NewDumper(SimpleConfig)
	return d.Sdump(values...)
}

// SdumpHTML returns the HTML-formatted string representation of the values
// using the DefaultConfig. The output is wrapped in HTML tags suitable for embedding in a web page.
func SdumpHTML(values ...any) string {
	d := NewDumper(DefaultConfig)
	return d.SdumpHTML(values...)
}

// SdumpHTMLValues returns a simplified HTML-formatted string of the values
// using the SimpleConfig.
func SdumpHTMLValues(values ...any) string {
	d := NewDumper(SimpleConfig)
	return d.SdumpHTML(values...)
}
