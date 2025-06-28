package govar

import (
	"fmt"
	"html"
)

// Formatter is an interface for applying text formatting styles.
// It defines a single method ApplyFormat, which takes a color code
// and a string, and returns the formatted string.
type Formatter interface {
	// ApplyFormat applies a formatting style to the input string `str`
	// using the given `colorCode`. The exact behavior depends on the
	// Formatter implementation.
	ApplyFormat(colorCode string, str string) string
}

// PlainFormatter implements the Formatter interface by returning
// the input string without any formatting applied.
type PlainFormatter struct{}

func (f PlainFormatter) ApplyFormat(colorCode string, str string) string {
	return str
}

// ANSIcolorFormatter implements the Formatter interface using ANSI
// escape codes to apply terminal color formatting.
type ANSIcolorFormatter struct{}

func (f ANSIcolorFormatter) ApplyFormat(colorCode string, str string) string {
	return colorCode + str + ColorReset
}

// HTMLformatter implements the Formatter interface by wrapping
// the input string in an HTML <span> tag with an inline style
// for color. It can be customized with an optional HTML tag token
// and a flag to enable or disable color usage.
type HTMLformatter struct {
	// HTMLtagToken can be used to add additional HTML attributes or
	// data tokens to the span tag, allowing customization of the output.
	HTMLtagToken string

	// UseColors determines whether the formatter should apply color styles.
	// If false, formatting will be disabled and plain text returned.
	UseColors bool
}

func (f HTMLformatter) ApplyFormat(colorCode string, str string) string {
	if f.UseColors {
		return fmt.Sprintf(`<%s style="color:%s">%s</%s>`, f.HTMLtagToken, ColorPaletteHTML[colorCode], html.EscapeString(str), f.HTMLtagToken)
	} else {
		return fmt.Sprintf(`<%s style="color:#fefefe">%s</%s>`, f.HTMLtagToken, html.EscapeString(str), f.HTMLtagToken)
	}
}
