package govar

import "io"

// Drop-in API for dumping colorized variables to any io.Writer
// With everything ON (types, meta-hints, embedded type methods, ...)
func Fdump(w io.Writer, values ...any) {
	defaultConfig := DumperConfig{
		IndentWidth:         3,
		MaxDepth:            15,
		MaxItems:            100,
		MaxStringLen:        10000,
		MaxInlineLength:     80,
		ShowTypes:           true,
		UseColors:           true,
		TrackReferences:     true,
		EmbedTypeMethods:    true,
		ShowMetaInformation: true,
		ShowHexdump:         true,
	}
	d := NewDumper(defaultConfig)
	d.Fdump(w, values...)
}

// Drop-in API for dumping colorized variables to any io.Writer
// With everything ON (types OFF, meta-hints OFF, embedded type methods OFF)
func FdumpValues(w io.Writer, values ...any) {
	defaultConfig := DumperConfig{
		IndentWidth:         3,
		MaxDepth:            15,
		MaxItems:            100,
		MaxStringLen:        10000,
		MaxInlineLength:     80,
		ShowTypes:           false,
		UseColors:           true,
		TrackReferences:     true,
		EmbedTypeMethods:    false,
		ShowMetaInformation: false,
		ShowHexdump:         true,
	}
	d := NewDumper(defaultConfig)
	d.Fdump(w, values...)
}

// Drop-in API for dumping colorized variables to stdio & die right after
// With everything ON (types ON, meta-hints ON, embedded type methods ON)
func Die(values ...any) {
	defaultConfig := DumperConfig{
		IndentWidth:         3,
		MaxDepth:            15,
		MaxItems:            100,
		MaxStringLen:        10000,
		MaxInlineLength:     80,
		ShowTypes:           true,
		UseColors:           true,
		TrackReferences:     true,
		EmbedTypeMethods:    true,
		ShowMetaInformation: true,
		ShowHexdump:         true,
	}
	d := NewDumper(defaultConfig)
	d.Die(values...)
}

// Drop-in API for dumping colorized variables to stdout
// With everything ON (types ON, meta-hints ON, embedded type methods ON)
func Dump(values ...any) {
	defaultConfig := DumperConfig{
		IndentWidth:         3,
		MaxDepth:            15,
		MaxItems:            100,
		MaxStringLen:        10000,
		MaxInlineLength:     80,
		ShowTypes:           true,
		UseColors:           true,
		TrackReferences:     true,
		EmbedTypeMethods:    true,
		ShowMetaInformation: true,
		ShowHexdump:         true,
	}
	d := NewDumper(defaultConfig)
	d.Dump(values...)
}

// Drop-in API for dumping colorized variables to stdout
// With everything ON (types OFF, meta-hints OFF, embedded type methods OFF)
func DumpValues(values ...any) {
	defaultConfig := DumperConfig{
		IndentWidth:         3,
		MaxDepth:            15,
		MaxItems:            100,
		MaxStringLen:        10000,
		MaxInlineLength:     80,
		ShowTypes:           false,
		UseColors:           true,
		TrackReferences:     true,
		EmbedTypeMethods:    false,
		ShowMetaInformation: false,
		ShowHexdump:         true,
	}
	d := NewDumper(defaultConfig)
	d.Dump(values...)
}

// Drop-in API for dumping colorized variables to a string
// With everything ON (types ON, meta-hints ON, embedded type methods ON)
func Sdump(values ...any) string {
	defaultConfig := DumperConfig{
		IndentWidth:         3,
		MaxDepth:            15,
		MaxItems:            100,
		MaxStringLen:        10000,
		MaxInlineLength:     80,
		ShowTypes:           true,
		UseColors:           true,
		TrackReferences:     true,
		EmbedTypeMethods:    true,
		ShowMetaInformation: true,
		ShowHexdump:         true,
	}
	d := NewDumper(defaultConfig)
	return d.Sdump(values...)
}

// Drop-in API for dumping colorized variables to a string
// With everything ON (types OFF, meta-hints OFF, embedded type methods OFF)
func SdumpValues(values ...any) {
	defaultConfig := DumperConfig{
		IndentWidth:         3,
		MaxDepth:            15,
		MaxItems:            100,
		MaxStringLen:        10000,
		MaxInlineLength:     80,
		ShowTypes:           false,
		UseColors:           true,
		TrackReferences:     true,
		EmbedTypeMethods:    false,
		ShowMetaInformation: false,
		ShowHexdump:         true,
	}
	d := NewDumper(defaultConfig)
	d.Sdump(values...)
}

// Drop-in API for dumping colorized variables to string with HTML formatting
// With everything ON (types ON, meta-hints ON, embedded type methods ON)
func SdumpHTML(values ...any) string {
	defaultConfig := DumperConfig{
		IndentWidth:         3,
		MaxDepth:            15,
		MaxItems:            100,
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
	}
	d := NewDumper(defaultConfig)
	return d.SdumpHTML(values...)
}

// Drop-in API for dumping colorized variables to string with HTML formatting
// With everything ON (types OFF, meta-hints OFF, embedded type methods OFF)
func SdumpHTMLValues(values ...any) string {
	defaultConfig := DumperConfig{
		IndentWidth:         3,
		MaxDepth:            15,
		MaxItems:            100,
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
	}
	d := NewDumper(defaultConfig)
	return d.SdumpHTML(values...)
}
