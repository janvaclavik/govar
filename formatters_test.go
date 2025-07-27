package govar_test

import (
	"testing"

	"github.com/janvaclavik/govar"
)

func TestFormatters(t *testing.T) {
	tests := []struct {
		name      string
		formatter govar.Formatter
		colorCode string
		input     string
		want      string
	}{
		{
			name:      "PlainFormatter ignores formatting",
			formatter: &govar.PlainFormatter{},
			colorCode: govar.ColorSkyBlue, // blue, but irrelevant
			input:     "hello",
			want:      "hello",
		},
		{
			name:      "ANSIcolorFormatter wraps input in ANSI escape",
			formatter: &govar.ANSIcolorFormatter{},
			colorCode: govar.ColorSkyBlue, // blue
			input:     "hello",
			want:      "\x1b[38;5;123mhello\x1b[0m",
		},
		{
			name:      "HTMLformatter wraps input in <html style>",
			formatter: &govar.HTMLformatter{UseColors: true, HTMLtagToken: "span"},
			colorCode: govar.ColorSkyBlue, // blue
			input:     "hello",
			want:      `<span style="color:#77DDEE">hello</span>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.formatter.ApplyFormat(tt.colorCode, tt.input)
			if got != tt.want {
				t.Errorf("ApplyFormat() = %q, want %q", got, tt.want)
			}
		})
	}
}
