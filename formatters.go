package govar

import (
	"fmt"
	"html"
)

type Formatter interface {
	ApplyFormat(colorCode string, str string) string
}

type PlainFormatter struct{}

func (f PlainFormatter) ApplyFormat(colorCode string, str string) string {
	return str
}

type ANSIcolorFormatter struct{}

func (f ANSIcolorFormatter) ApplyFormat(colorCode string, str string) string {
	return colorCode + str + ColorReset
}

type HTMLformatter struct {
	HTMLtagToken string
	UseColors    bool
}

func (f HTMLformatter) ApplyFormat(colorCode string, str string) string {
	if f.UseColors {
		return fmt.Sprintf(`<%s style="color:%s">%s</%s>`, f.HTMLtagToken, ColorPaletteHTML[colorCode], html.EscapeString(str), f.HTMLtagToken)
	} else {
		return fmt.Sprintf(`<%s style="color:#fefefe">%s</%s>`, f.HTMLtagToken, html.EscapeString(str), f.HTMLtagToken)
	}
}
