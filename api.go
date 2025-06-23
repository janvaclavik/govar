package govar

import "io"

// Drop-in API for dumping colorized variables to any io.Writer
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
		HTMLtagToken:        "span",
		HTMLtagSection:      "pre",
		EmbedTypeMethods:    true,
		ShowMetaInformation: true,
	}
	d := NewDumper(defaultConfig)
	d.Fdump(w, values...)
}

// Drop-in API for dumping colorized variables to stdio & die right after
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
		HTMLtagToken:        "span",
		HTMLtagSection:      "pre",
		EmbedTypeMethods:    true,
		ShowMetaInformation: true,
	}
	d := NewDumper(defaultConfig)
	d.Die(values...)
}

// Drop-in API for dumping colorized variables to stdout
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
		HTMLtagToken:        "span",
		HTMLtagSection:      "pre",
		EmbedTypeMethods:    true,
		ShowMetaInformation: true,
	}
	d := NewDumper(defaultConfig)
	d.Dump(values...)
}

// Drop-in API for dumping colorized variables to string
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
	}
	d := NewDumper(defaultConfig)
	return d.SdumpHTML(values...)
}

// Drop-in API for dumping colorized variables to string
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
		HTMLtagToken:        "span",
		HTMLtagSection:      "pre",
		EmbedTypeMethods:    true,
		ShowMetaInformation: true,
	}
	d := NewDumper(defaultConfig)
	return d.Sdump(values...)
}
