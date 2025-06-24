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

type HTMLformatter struct{}

func (f HTMLformatter) ApplyFormat(code string, str string) string {
	return fmt.Sprintf(`<span style="color:%s">%s</span>`, ColorPaletteHTML[code], html.EscapeString(str))
}
