package govar_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/janvaclavik/govar"
)

func TestAPI_StringReturn(t *testing.T) {
	tests := []struct {
		name     string
		call     func() string
		expected string
	}{
		{
			name:     "Sdump",
			call:     func() string { return govar.Sdump(map[string]int{"a": 1}) },
			expected: "map",
		},
		{
			name:     "SdumpNoColors",
			call:     func() string { return govar.SdumpNoColors([]int{1, 2}) },
			expected: "[]int",
		},
		{
			name:     "SdumpValues",
			call:     func() string { return govar.SdumpValues("xyz") },
			expected: "xyz",
		},
		{
			name:     "SdumpHTML",
			call:     func() string { return govar.SdumpHTML("html") },
			expected: "<pre",
		},
		{
			name:     "SdumpHTMLValues",
			call:     func() string { return govar.SdumpHTMLValues("html") },
			expected: "html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := tt.call()
			if !strings.Contains(output, tt.expected) {
				t.Errorf("Expected output to contain %q, got: %s", tt.expected, output)
			}
		})
	}
}

func TestAPI_WriterOutput(t *testing.T) {
	tests := []struct {
		name     string
		call     func(w *bytes.Buffer)
		expected string
	}{
		{
			name:     "Fdump",
			call:     func(w *bytes.Buffer) { govar.Fdump(w, 123) },
			expected: "123",
		},
		{
			name:     "FdumpNoColors",
			call:     func(w *bytes.Buffer) { govar.FdumpNoColors(w, "abc") },
			expected: "abc",
		},
		{
			name:     "FdumpValues",
			call:     func(w *bytes.Buffer) { govar.FdumpValues(w, struct{ X int }{42}) },
			expected: "42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			tt.call(&buf)
			if !strings.Contains(buf.String(), tt.expected) {
				t.Errorf("Expected output to contain %q, got: %s", tt.expected, buf.String())
			}
		})
	}
}
