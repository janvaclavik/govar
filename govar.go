package govar

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"text/tabwriter"
	"unicode/utf8"
	"unsafe"
)

const (
	PackageName = "govar"
	Version     = "0.4.0"
)

// (OLD) ANSI Colors
const (
	ColorDarkGray  = "\033[38;5;238m"
	ColorGray      = "\033[38;5;245m"
	ColorLightGray = "\033[38;5;250m"
	ColorWhite     = "\033[38;5;15m"

	ColorMutedRed = "\033[38;5;131m"
	ColorRed      = "\033[38;5;160m"
	ColorLightRed = "\033[38;5;196m"

	ColorOrange = "\033[38;5;208m"
	ColorYellow = "\033[38;5;226m"
	ColorGold   = "\033[38;5;178m"

	ColorGreen = "\033[38;5;34m"
	ColorLime  = "\033[38;5;113m"
	ColorTeal  = "\033[38;5;37m"
	ColorAqua  = "\033[38;5;86m"

	ColorMutedBlue = "\033[38;5;25m"
	ColorSkyBlue   = "\033[38;5;117m"
	ColorBlue      = "\033[38;5;33m"
	ColorCyan      = "\033[38;5;51m"

	ColorViolet  = "\033[38;5;135m"
	ColorPink    = "\033[38;5;218m"
	ColorMagenta = "\033[38;5;201m"

	ColorReset = "\033[0m"
)

// (OLD) ColorPaletteHTML maps color codes to HTML colors.
var ColorPaletteHTML = map[string]string{
	ColorDarkGray:  "#444444",
	ColorGray:      "#8a8a8a",
	ColorLightGray: "#bcbcbc",
	ColorWhite:     "#fff",

	ColorMutedRed: "#aa4444",
	ColorRed:      "#d70000",
	ColorLightRed: "#ff2b2b",

	ColorOrange: "#ff8700",
	ColorYellow: "#ffff00",
	ColorGold:   "#d7af5f",

	ColorGreen: "#00af00",
	ColorLime:  "#80ff80",
	ColorTeal:  "#00afaf",
	ColorAqua:  "#5fd7af",

	ColorMutedBlue: "#336699",
	ColorSkyBlue:   "#87d7ff",
	ColorBlue:      "#0087ff",
	ColorCyan:      "#00ffff",

	ColorViolet:  "#af5fff",
	ColorPink:    "#ffafd7",
	ColorMagenta: "#ff5fff",
}

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
	return fmt.Sprintf(`<span style="color:%s">%s</span>`, ColorPaletteHTML[code], str)
}

type DumperConfig struct {
	IndentWidth         int
	MaxDepth            int
	MaxItems            int
	MaxStringLen        int
	MaxInlineLength     int
	UseColors           bool
	TrackReferences     bool
	HTMLtagToken        string
	HTMLtagSection      string
	ShowTypes           bool
	EmbedTypeMethods    bool
	ShowMetaInformation bool
}

type Dumper struct {
	nextRefID    int
	referenceMap map[uintptr]int
	config       DumperConfig
	Formatter
}

func NewDumper(cfg DumperConfig) *Dumper {
	return &Dumper{nextRefID: 1, referenceMap: map[uintptr]int{}, config: cfg, Formatter: PlainFormatter{}}
}

// Drop-in API for dumping colorized variables to any io.Writer
func Fdump(w io.Writer, values ...any) {
	defaultConfig := DumperConfig{
		IndentWidth:         3,
		MaxDepth:            15,
		MaxItems:            100,
		MaxStringLen:        10000,
		MaxInlineLength:     80,
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

// Dump prints the values to stdout with colorized output.
func (d *Dumper) Dump(vs ...any) {
	// Enable HTML coloring
	d.Formatter = ANSIcolorFormatter{}

	d.renderHeader(os.Stdout)
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	d.renderAllValues(tw, vs...)
	tw.Flush()
}

// Fdump writes the formatted dump of values to the given io.Writer.
func (d *Dumper) Fdump(w io.Writer, vs ...any) {
	d.renderHeader(w)
	tw := tabwriter.NewWriter(w, 0, 0, 1, ' ', 0)
	d.renderAllValues(tw, vs...)
	tw.Flush()
}

// Sdump dumps the values as a string with colorized output.
func (d *Dumper) Sdump(vs ...any) string {
	var sb strings.Builder
	d.renderHeader(&sb)
	tw := tabwriter.NewWriter(&sb, 0, 0, 1, ' ', 0)
	d.renderAllValues(tw, vs...)
	tw.Flush()
	return sb.String()
}

// HTMLdump dumps the values as HTML inside a <pre> tag with colorized output.
func (d *Dumper) SdumpHTML(vs ...any) string {
	// Enable HTML coloring
	d.Formatter = HTMLformatter{}

	var sb strings.Builder
	sb.WriteString(`<pre class="govar" style="background-color:black; color:white; padding:4px; border-radius: 4px">` + "\n")

	tw := tabwriter.NewWriter(&sb, 0, 0, 1, ' ', 0)
	d.renderHeader(&sb)
	d.renderAllValues(tw, vs...)
	tw.Flush()

	sb.WriteString("</pre>")
	return sb.String()
}

// Die is a debug function that prints the values and exits the program.
func (d *Dumper) Die(vs ...any) {
	Dump(vs...)
	os.Exit(1)
}

func (d *Dumper) estimatedInlineLength(v reflect.Value) int {
	length := 0
	switch v.Kind() {
	case reflect.String:
		strVal := v.String()
		runeCount := utf8.RuneCountInString(strVal)
		length += runeCount + 2
		if d.config.ShowMetaInformation {
			meta := fmt.Sprintf(" [run=%d]", runeCount)
			length += len(meta)
		}
		return length
	case reflect.Int, reflect.Int64:
		return len(strconv.FormatInt(v.Int(), 10))
	case reflect.Uint, reflect.Uint64:
		return len(strconv.FormatUint(v.Uint(), 10))
	case reflect.Float64:
		return len(strconv.FormatFloat(v.Float(), 'f', -1, 64))
	case reflect.Bool:
		if v.Bool() {
			return length + 4
		}
		return length + 5

	case reflect.Array, reflect.Slice:
		// TODO: should account for type lengths if ShowTypes is ON
		length += 2 // braces
		if d.config.ShowMetaInformation {
			if v.Kind() == reflect.Slice {
				length += len(fmt.Sprintf("[len=%d cap=%d] ", v.Len(), v.Cap()))
			} else {
				length += len(fmt.Sprintf("[len=%d] ", v.Len()))
			}
		}
		for i := range v.Len() {
			if i > 0 {
				length += 2 // comma and space
			}
			length += 1 + 4 + d.estimatedInlineLength(v.Index(i)) // i => val
		}
		return length

	case reflect.Map:
		// TODO: should account for type lengths if ShowTypes is ON
		length += 2 // braces
		if d.config.ShowMetaInformation {
			length += len(fmt.Sprintf("[len=%d] ", v.Len()))
		}
		for i, key := range v.MapKeys() {
			if i > 0 {
				length += 2 // comma and space
			}
			val := v.MapIndex(key)
			length += d.estimatedInlineLength(key) + 4 + d.estimatedInlineLength(val) // key => val
		}
		return length

	case reflect.Struct:
		// TODO: should account for type lengths if ShowTypes is ON
		length += 2 // braces
		t := v.Type()
		for i := range v.NumField() {
			if i > 0 {
				length += 2 // comma and space
			}
			name := t.Field(i).Name
			length += len(name) + 4 + d.estimatedInlineLength(v.Field(i)) // Name => val
		}
		return length

	default:
		return 10 // fallback
	}
}

// Returns a string representation for a value type (and handle any type)
func (d *Dumper) formatType(v reflect.Value, isInCollection bool) string {

	if !v.IsValid() {
		return d.ApplyFormat(ColorDarkGray, "invalid")
	}

	// print element type signature
	vKind := v.Kind()
	expectedType := ""
	if vKind == reflect.Array || vKind == reflect.Slice || vKind == reflect.Map || vKind == reflect.Struct || vKind == reflect.Interface {
		expectedType = d.ApplyFormat(ColorDarkGray, v.Type().String())
	} else if !isInCollection {
		expectedType = d.ApplyFormat(ColorDarkGray, v.Type().String())
	}

	// if element type is just "any", print the actual variable type
	actualType := ""
	if vKind == reflect.Interface && v.Type().NumMethod() == 0 && !v.IsNil() {
		actualType = d.ApplyFormat(ColorDarkGray, " ("+v.Elem().Type().String()+")")
	}
	formattedType := expectedType + actualType

	// Modernize the 'interface {}' to 'any'
	formattedType = strings.ReplaceAll(formattedType, "interface {}", "any")
	return formattedType
}

// renderHeader prints the header for the dump output, including the file and line number.
func (d *Dumper) renderHeader(out io.Writer) {
	file, line, govarFuncName := findCallerInStack()
	if file == "" {
		return
	}

	relPath := file
	if wd, err := os.Getwd(); err == nil {
		if rel, err := filepath.Rel(wd, file); err == nil {
			relPath = rel
		}
	}

	headerTitle := d.ApplyFormat(ColorOrange, "[>] "+govarFuncName)
	headerLocation := d.ApplyFormat(ColorGray, fmt.Sprintf("  ←  %s:%d", relPath, line))
	header := headerTitle + headerLocation
	fmt.Fprintln(out, d.ApplyFormat(ColorGray, header))
}

// renderAllValues writes all the values to the tabwriter, handling references and indentation.
func (d *Dumper) renderAllValues(tw *tabwriter.Writer, vs ...any) {
	d.referenceMap = map[uintptr]int{} // reset each time
	visited := map[uintptr]bool{}
	for _, v := range vs {
		rv := reflect.ValueOf(v)
		rv = makeAddressable(rv)

		// Render value's type signature
		fmt.Fprint(tw, d.ApplyFormat(ColorGray, d.formatType(rv, false)))
		fmt.Fprint(tw, " => ")
		// Render the value itself
		d.renderValue(tw, rv, 0, visited)

		fmt.Fprintln(tw)
	}
}

// renderValue recursively writes the value with indentation and handles references.
func (d *Dumper) renderValue(tw *tabwriter.Writer, v reflect.Value, level int, visited map[uintptr]bool) {
	if level > d.config.MaxDepth {
		fmt.Fprint(tw, d.ApplyFormat(ColorGray, "… (max depth reached)"))
		return
	}
	if !v.IsValid() {
		fmt.Fprint(tw, d.ApplyFormat(ColorRed, "<invalid>"))
		return
	}

	if isNil(v) {
		fmt.Fprint(tw, d.ApplyFormat(ColorMutedRed, "<nil>"))
		return
	}

	if str := d.asStringer(v); str != "" {
		fmt.Fprint(tw, str)
		if d.config.ShowMetaInformation {
			fmt.Fprint(tw, d.ApplyFormat(ColorMutedBlue, " [⧉ fmt.Stringer]"))
		}
		return
	}

	switch v.Kind() {
	case reflect.Ptr:
		// If a pointer type is addressable and known, show a reference marker
		// If a pointer type is addressable and new, store it in the reference map
		if v.CanAddr() {
			ptr := v.Pointer()
			if id, ok := d.referenceMap[ptr]; ok {
				fmt.Fprintf(tw, d.ApplyFormat(ColorLightGray, "↩︎ &%d"), id)
				return
			} else {
				d.referenceMap[ptr] = d.nextRefID
				d.nextRefID++
			}
		}
		// Continue with rendering the value that the pointer points to
		d.renderValue(tw, v.Elem(), level, visited)
	case reflect.Interface:
		// Continue with rendering the value that the interface contains
		d.renderValue(tw, v.Elem(), level, visited)
	case reflect.UnsafePointer:
		fmt.Fprint(tw, d.ApplyFormat(ColorGray, fmt.Sprintf("unsafe.Pointer(%#x)", v.Pointer())))
	case reflect.Bool:
		if v.Bool() {
			fmt.Fprint(tw, d.ApplyFormat(ColorGreen, "true"))
		} else {
			fmt.Fprint(tw, d.ApplyFormat(ColorLightRed, "false"))
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fmt.Fprint(tw, d.ApplyFormat(ColorCyan, fmt.Sprint(v.Int())))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		fmt.Fprint(tw, d.ApplyFormat(ColorCyan, fmt.Sprint(v.Uint())))
	case reflect.Float32, reflect.Float64:
		fmt.Fprint(tw, d.ApplyFormat(ColorCyan, fmt.Sprintf("%f", v.Float())))
	case reflect.Complex64, reflect.Complex128:
		fmt.Fprint(tw, d.ApplyFormat(ColorCyan, fmt.Sprintf("%v", v.Complex())))
	case reflect.String:
		strLen := utf8.RuneCountInString(v.String())
		str := d.stringEscape(v.String())
		str = d.ApplyFormat(ColorOrange, `"`) + d.ApplyFormat(ColorLime, str) + d.ApplyFormat(ColorOrange, `"`)
		fmt.Fprint(tw, str)
		if d.config.ShowMetaInformation {
			fmt.Fprint(tw, d.ApplyFormat(ColorMutedBlue, fmt.Sprintf(" [run=%d]", strLen)))
		}
	case reflect.Struct:
		t := v.Type()
		fmt.Fprint(tw, "{")
		if !d.shouldRenderInline(v) {
			fmt.Fprintln(tw)
		}

		visibleFields := reflect.VisibleFields(t)
		for i, field := range visibleFields {
			fieldVal := v.FieldByIndex(field.Index)
			symbol := "⯀ "
			if field.PkgPath != "" {
				symbol = "🞏 "
				fieldVal = forceExported(fieldVal)
			}
			if !d.shouldRenderInline(v) {
				// print visibility and symbol name, with indent
				d.renderIndent(tw, level+1, d.ApplyFormat(ColorOrange, symbol)+field.Name)
			} else {
				// inline render of the field
				fmt.Fprintf(tw, d.ApplyFormat(ColorOrange, symbol)+field.Name)
			}
			// print field type signature
			formattedType := d.formatType(fieldVal, false)
			fmt.Fprintf(tw, "	%s	=> ", formattedType)

			// Try the stringer interface on this struct field first
			if str := d.asStringer(fieldVal); str != "" {
				fmt.Fprint(tw, str)
				if d.config.ShowMetaInformation {
					fmt.Fprint(tw, d.ApplyFormat(ColorMutedBlue, " [⧉ fmt.Stringer]"))
				}
			} else {
				// or recursively render the field value itself
				if !d.shouldRenderInline(v) {
					d.renderValue(tw, fieldVal, level+1, visited)
				} else {
					// inline render
					d.renderValue(tw, fieldVal, level, visited)
				}
			}

			if !d.shouldRenderInline(v) {
				fmt.Fprintln(tw)
			} else {
				if i != len(visibleFields)-1 {
					fmt.Fprint(tw, ", ")
				}
			}
		}
		// print all of struct's type methods (never inline)
		if d.config.EmbedTypeMethods {
			d.renderTypeMethods(tw, t, level+1)
		}

		if !d.shouldRenderInline(v) {
			d.renderIndent(tw, level, "")
		}
		fmt.Fprint(tw, "}")
	case reflect.Map:
		if d.config.ShowMetaInformation {
			mapLen := fmt.Sprintf("[len=%d] ", v.Len())
			fmt.Fprint(tw, d.ApplyFormat(ColorMutedBlue, mapLen))
		}

		fmt.Fprint(tw, "{")
		if !d.shouldRenderInline(v) {
			fmt.Fprintln(tw)
		}

		keys := v.MapKeys()
		for i, key := range keys {
			if i >= d.config.MaxItems {
				d.renderIndent(tw, level+1, d.ApplyFormat(ColorGray, "… (truncated)"))
				break
			}
			keyStr := fmt.Sprintf("%v", key.Interface())

			// print element type signature
			formattedType := d.formatType(v.MapIndex(key), true)

			if !d.shouldRenderInline(v) {
				// indent, render key and type
				d.renderIndent(tw, level+1, fmt.Sprintf("%s %s	=> ", d.ApplyFormat(ColorViolet, keyStr), formattedType))
				// recursively print the array value itself, increase indent level
				d.renderValue(tw, v.MapIndex(key), level+1, visited)
			} else {
				// do not indent, render key and type
				fmt.Fprintf(tw, "%s %s	=> ", d.ApplyFormat(ColorViolet, keyStr), formattedType)
				// recursively print the array value itself, same indent level
				d.renderValue(tw, v.MapIndex(key), level, visited)
			}

			if !d.shouldRenderInline(v) {
				fmt.Fprintln(tw)
			} else {
				if i != v.Len()-1 {
					fmt.Fprint(tw, ", ")
				}
			}
		}
		if !d.shouldRenderInline(v) {
			d.renderIndent(tw, level, "")
		}
		fmt.Fprint(tw, "}")
	case reflect.Slice, reflect.Array:
		if d.config.ShowMetaInformation {
			var listLen string
			if v.Kind() == reflect.Array {
				listLen = fmt.Sprintf("[len=%d] ", v.Len())
			} else {
				listLen = fmt.Sprintf("[len=%d cap=%d] ", v.Len(), v.Cap())
			}
			fmt.Fprint(tw, d.ApplyFormat(ColorMutedBlue, listLen))
		}
		fmt.Fprint(tw, "{")
		if !d.shouldRenderInline(v) {
			fmt.Fprintln(tw)
		}
		for i := range v.Len() {
			if i >= d.config.MaxItems {
				d.renderIndent(tw, level+1, d.ApplyFormat(ColorGray, "… (truncated)\n"))
				break
			}
			// print element type signature
			formattedType := d.formatType(v.Index(i), true)
			if !d.shouldRenderInline(v) {
				// indent, render index (and type)
				d.renderIndent(tw, level+1, fmt.Sprintf("%s %s => ", d.ApplyFormat(ColorCyan, fmt.Sprintf("%d", i)), formattedType))
				// recursively print the array value itself, increase indent level
				d.renderValue(tw, v.Index(i), level+1, visited)
			} else {
				// do not indent, render index (and type)
				fmt.Fprintf(tw, "%s %s => ", d.ApplyFormat(ColorCyan, fmt.Sprintf("%d", i)), formattedType)
				// recursively print the array value itself, same indent level
				d.renderValue(tw, v.Index(i), level, visited)
			}

			if !d.shouldRenderInline(v) {
				fmt.Fprintln(tw)
			} else {
				if i != v.Len()-1 {
					fmt.Fprint(tw, ", ")
				}
			}
		}

		if !d.shouldRenderInline(v) {
			d.renderIndent(tw, level, "")
		}

		fmt.Fprint(tw, "}")
	case reflect.Func:
		funName := d.ApplyFormat(ColorBlue, getFunctionName(v))
		fmt.Fprint(tw, funName)
		if d.config.ShowMetaInformation {
			funMeta := d.ApplyFormat(ColorMutedBlue, fmt.Sprintf(" [func@%#x]", v.Pointer()))
			fmt.Fprint(tw, funMeta)
		}
	case reflect.Chan:
		if v.IsNil() {
			fmt.Fprint(tw, d.ApplyFormat(ColorMutedRed, "<nil>"))
		} else {

			fmt.Fprintf(tw, "%s%s", d.ApplyFormat(ColorPink, "chan@"), d.ApplyFormat(ColorTeal, fmt.Sprintf("%#x", v.Pointer())))
			if d.config.ShowMetaInformation {
				bufferStr := d.ApplyFormat(ColorMutedBlue, fmt.Sprintf(" [buf=%d]", v.Cap()))
				fmt.Fprint(tw, bufferStr)
			}

		}
	default:
		// Should be unreachable - all reflect.Kind cases are handled
	}
}

// renderIndent writes indented text to the tabwriter.
func (d *Dumper) renderIndent(tw *tabwriter.Writer, indentLevel int, text string) {
	fmt.Fprint(tw, strings.Repeat(" ", indentLevel*d.config.IndentWidth)+text)
}

func (d *Dumper) renderTypeMethods(tw *tabwriter.Writer, t reflect.Type, level int) {
	for _, m := range findTypeMethods(t) {
		// print visibility and symbol name
		symbol := "⦿ "
		methodType := " " + d.ApplyFormat(ColorGray, m.Func.Type().String())
		d.renderIndent(tw, level, d.ApplyFormat(ColorOrange, symbol)+m.Name+methodType)
		if d.config.ShowMetaInformation {
			fmt.Fprint(tw, d.ApplyFormat(ColorMutedBlue, " [Method]"))
		}
		fmt.Fprintln(tw)
	}
}

// asStringer checks if the value implements fmt.Stringer and returns its string representation.
func (d *Dumper) asStringer(v reflect.Value) string {
	val := v
	if !val.CanInterface() {
		val = forceExported(val)
	}
	if val.CanInterface() {
		if s, ok := val.Interface().(fmt.Stringer); ok {
			rv := reflect.ValueOf(s)
			if rv.Kind() == reflect.Ptr && rv.IsNil() {
				return d.ApplyFormat(ColorMutedRed, "<nil>")
			}
			str := d.stringEscape(s.String())
			str = d.ApplyFormat(ColorOrange, `"`) + d.ApplyFormat(ColorLime, str) + d.ApplyFormat(ColorOrange, `"`)
			return str
		}
	}
	return ""
}

// shouldRenderInline determines if a value should be printed inline.
func (d *Dumper) shouldRenderInline(v reflect.Value) bool {
	// Handle zero or invalid
	if !v.IsValid() {
		return true
	}

	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		return isSimpleCollection(v) && v.Len() <= 10 && d.estimatedInlineLength(v) <= d.config.MaxInlineLength

	case reflect.Map:
		return isSimpleMap(v) && v.Len() <= 10 && d.estimatedInlineLength(v) <= d.config.MaxInlineLength

	case reflect.Struct:
		return d.isSimpleStruct(v) && v.NumField() <= 10 && d.estimatedInlineLength(v) <= d.config.MaxInlineLength

	default:
		// Scalars and other simple types can always be inline
		return true
	}
}

// stringEscape escapes control characters like newline in a string for safe display.
// It also truncates strings that are too long to be pretty
func (d *Dumper) stringEscape(str string) string {
	if utf8.RuneCountInString(str) > d.config.MaxStringLen {
		runes := []rune(str)
		str = string(runes[:d.config.MaxStringLen]) + "…"
	}

	replacer := strings.NewReplacer(
		"\n", `\n`,
		"\t", `\t`,
		"\r", `\r`,
		"\v", `\v`,
		"\f", `\f`,
		"\x1b", `\x1b`,
	)

	return replacer.Replace(str)
}

// isNil checks if the value is nil on any kind of object
// It does not fail even if the value type cannot be nil (bool, etc...)
func isNil(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Interface, reflect.Func, reflect.Chan:
		return v.IsNil()
	default:
		return false
	}
}

func isSimpleValue(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}

	switch v.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int64, reflect.Uint, reflect.Uint64, reflect.Float64, reflect.String:
		return true
	default:
		return false
	}
}

func isSimpleCollection(v reflect.Value) bool {
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		if !isSimpleValue(elem) {
			return false
		}
	}
	return true
}

func isSimpleMap(v reflect.Value) bool {
	for _, key := range v.MapKeys() {
		val := v.MapIndex(key)
		if !isSimpleValue(key) || !isSimpleValue(val) {
			return false
		}
	}
	return true
}

func (d *Dumper) isSimpleStruct(v reflect.Value) bool {
	// Reject if the struct has methods and config says to show them
	if d.config.EmbedTypeMethods && len(findTypeMethods(v.Type())) > 0 {
		return false
	}

	for i := 0; i < v.NumField(); i++ {
		if !isSimpleValue(v.Field(i)) {
			return false
		}
	}
	return true
}

// findCallerInStack finds the first non-govar function call in the call-stack.
func findCallerInStack() (string, int, string) {
	govarFuncName := ""
	for i := 2; i < 15; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		if fn == nil || !strings.Contains(fn.Name(), "/"+PackageName) {
			return file, line, govarFuncName
		}
		tmpNameSliced := strings.Split(fn.Name(), "/")
		govarFuncName = tmpNameSliced[len(tmpNameSliced)-1]
	}
	return "", 0, ""
}

func findTypeMethods(typ reflect.Type) []reflect.Method {
	seen := make(map[string]bool)
	methods := []reflect.Method{}

	// Check value receiver methods
	for i := range typ.NumMethod() {
		m := typ.Method(i)
		if !seen[m.Name] {
			methods = append(methods, m)
			seen[m.Name] = true
		}
	}

	// Check pointer receiver methods if applicable
	if typ.Kind() != reflect.Ptr {
		ptrType := reflect.PointerTo(typ)
		for i := range ptrType.NumMethod() {
			m := ptrType.Method(i)
			if !seen[m.Name] {
				methods = append(methods, m)
				seen[m.Name] = true
			}
		}
	}

	return methods
}

// forceExported returns a value that is guaranteed to be exported, even if it is unexported.
func forceExported(v reflect.Value) reflect.Value {
	if v.CanInterface() {
		return v
	}
	if v.CanAddr() {
		return reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem()
	}
	// Final fallback: return original value, even if unexported
	return v
}

func getFunctionName(v reflect.Value) string {
	return runtime.FuncForPC(v.Pointer()).Name()
}

// makeAddressable ensures the value is addressable, wrapping structs in pointers if necessary.
func makeAddressable(v reflect.Value) reflect.Value {
	// Already addressable? Do nothing
	if v.CanAddr() {
		return v
	}

	// If it's a struct and not addressable, wrap it in a pointer
	if v.Kind() == reflect.Struct {
		ptr := reflect.New(v.Type())
		ptr.Elem().Set(v)
		return ptr.Elem()
	}

	return v
}
