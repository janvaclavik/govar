package govar

import (
	"fmt"
	"html"
)

type Formatter interface {
	ApplyFormat(code string, str string) string
}

type PlainFormatter struct{}

func (f PlainFormatter) ApplyFormat(code string, str string) string {
	return str
}

type ANSIcolorFormatter struct{}

func (f ANSIcolorFormatter) ApplyFormat(code string, str string) string {
	return code + str + ColorReset
}

type HTMLformatter struct {
	HTMLtagToken string
	UseColors    bool
}

func (f HTMLformatter) ApplyFormat(code string, str string) string {
	if f.UseColors {
		return fmt.Sprintf(`<%s style="color:%s">%s</%s>`, f.HTMLtagToken, ColorPaletteHTML[code], html.EscapeString(str), f.HTMLtagToken)
	} else {
		return fmt.Sprintf(`<%s style="color:#fefefe">%s</%s>`, f.HTMLtagToken, html.EscapeString(str), f.HTMLtagToken)
	}
}
