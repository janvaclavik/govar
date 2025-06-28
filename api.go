package govar

import "io"

// Die dumps the provided values with full formatting (types, metadata, methods, etc.)
// to stdout, then terminates the program with os.Exit(1).
func Die(values ...any) {
	defaultConfig := DumperConfig{
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
	}
	d := NewDumper(defaultConfig)
	d.Die(values...)
}

// Dump prints the given values with full formatting (types, metadata, methods, etc.)
// to stdout. Colors are enabled.
func Dump(values ...any) {
	defaultConfig := DumperConfig{
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
	}
	d := NewDumper(defaultConfig)
	d.Dump(values...)
}

// DumpNoColors prints the given values to stdout with full formatting,
// but colors are disabled
func DumpNoColors(values ...any) {
	defaultConfig := DumperConfig{
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
	}
	d := NewDumper(defaultConfig)
	d.Dump(values...)
}

// DumpValues prints the values to stdout with a simplified format:
// types, metadata, and embedded methods are disabled. Colors are enabled.
func DumpValues(values ...any) {
	defaultConfig := DumperConfig{
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
	}
	d := NewDumper(defaultConfig)
	d.Dump(values...)
}

// Fdump writes the formatted output of the given values to the provided io.Writer.
// All formatting options are enabled, including color.
func Fdump(w io.Writer, values ...any) {
	defaultConfig := DumperConfig{
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
	}
	d := NewDumper(defaultConfig)
	d.Fdump(w, values...)
}

// FdumpNoColors writes formatted output to the given writer, with all formatting
// enabled except colored output
func FdumpNoColors(w io.Writer, values ...any) {
	defaultConfig := DumperConfig{
		IndentWidth:         3,
		MaxDepth:            15,
		MaxItems:            150,
		MaxStringLen:        10000,
		MaxInlineLength:     80,
		ShowTypes:           true,
		UseColors:           false,
		TrackReferences:     true,
		EmbedTypeMethods:    true,
		ShowMetaInformation: true,
		ShowHexdump:         true,
		IgnoreStringer:      false,
	}
	d := NewDumper(defaultConfig)
	d.Fdump(w, values...)
}

// FdumpValues writes the simplified formatted output (types off, metadata off, methods off)
// to the given writer. Colors are enabled.
func FdumpValues(w io.Writer, values ...any) {
	defaultConfig := DumperConfig{
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
	}
	d := NewDumper(defaultConfig)
	d.Fdump(w, values...)
}

// Sdump returns the full-formatted string representation of the given values.
// All features are enabled, including types, metadata, colors, and embedded methods.
func Sdump(values ...any) string {
	defaultConfig := DumperConfig{
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
	}
	d := NewDumper(defaultConfig)
	return d.Sdump(values...)
}

// SdumpNoColors returns the formatted string representation with all features enabled
// except colored output
func SdumpNoColors(values ...any) string {
	defaultConfig := DumperConfig{
		IndentWidth:         3,
		MaxDepth:            15,
		MaxItems:            150,
		MaxStringLen:        10000,
		MaxInlineLength:     80,
		ShowTypes:           true,
		UseColors:           false,
		TrackReferences:     true,
		EmbedTypeMethods:    true,
		ShowMetaInformation: true,
		ShowHexdump:         true,
		IgnoreStringer:      false,
	}
	d := NewDumper(defaultConfig)
	return d.Sdump(values...)
}

// SdumpValues returns the simplified string representation of the given values.
// Type info, metadata, and embedded methods are disabled. Colors are enabled.
func SdumpValues(values ...any) string {
	defaultConfig := DumperConfig{
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
	}
	d := NewDumper(defaultConfig)
	return d.Sdump(values...)
}

// SdumpHTML returns the HTML-formatted string representation of the values
// with full formatting options enabled, including embedded methods, metadata and
// colored output
func SdumpHTML(values ...any) string {
	defaultConfig := DumperConfig{
		IndentWidth:         3,
		MaxDepth:            15,
		MaxItems:            150,
		MaxStringLen:        10000,
		MaxInlineLength:     80,
		ShowTypes:           true,
		UseColors:           true,
		TrackReferences:     true,
		HTMLtagToken:        "span",
		HTMLtagSection:      "pre",
		EmbedTypeMethods:    true,
		ShowMetaInformation: true,
		ShowHexdump:         true,
		IgnoreStringer:      false,
	}
	d := NewDumper(defaultConfig)
	return d.SdumpHTML(values...)
}

// SdumpHTMLValues returns a simplified HTML-formatted string of the values.
// Type information, metadata, and embedded methods are omitted.
func SdumpHTMLValues(values ...any) string {
	defaultConfig := DumperConfig{
		IndentWidth:         3,
		MaxDepth:            15,
		MaxItems:            150,
		MaxStringLen:        10000,
		MaxInlineLength:     80,
		ShowTypes:           false,
		UseColors:           true,
		TrackReferences:     true,
		HTMLtagToken:        "span",
		HTMLtagSection:      "pre",
		EmbedTypeMethods:    false,
		ShowMetaInformation: false,
		ShowHexdump:         true,
		IgnoreStringer:      false,
	}
	d := NewDumper(defaultConfig)
	return d.SdumpHTML(values...)
}
