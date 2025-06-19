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

const (
	SymbolArrL    = "["
	SymbolArrR    = "]"
	SymbolStructL = "{"
	SymbolStructR = "}"
	SymbolMetaL   = "≪" // ⟪, ‹, ⟦
	SymbolMetaR   = "≫" // ⟫, ›, ⟧
)

// ANSI color codes inspired by Go brand colors
const (
	ColorPaleGray  = "\033[38;5;250m" // #B0BEC5
	ColorSlateGray = "\033[38;5;245m" // #A0A8B3
	ColorDimGray   = "\033[38;5;240m" // #5F6368
	ColorDarkGray  = "\033[38;5;238m" // #444444
	ColorCharcoal  = "\033[38;5;235m" // #262626

	ColorLime       = "\033[38;5;113m" // #80ff80
	ColorSkyBlue    = "\033[38;5;117m" // #55C1E7
	ColorMutedBlue  = "\033[38;5;110m" // #73B3D5
	ColorLightTeal  = "\033[38;5;73m"  // #46A4B0
	ColorBrightCyan = "\033[38;5;51m"  // #00D5EF
	ColorGoBlue     = "\033[38;5;33m"  // #00ADD8
	ColorDarkTeal   = "\033[38;5;30m"  // #005F5F
	ColorDarkGoBlue = "\033[38;5;24m"  // #005F87

	ColorSeafoamGreen = "\033[38;5;79m" // #60DBD4
	ColorGreen        = "\033[38;5;34m" // #00af00

	ColorGoldenrod   = "\033[38;5;220m" // #FFD54F
	ColorCoralRed    = "\033[38;5;203m" // #F46C5E
	ColorRed         = "\033[38;5;160m" // #d70000
	ColorDarkRed     = "\033[38;5;131m" // #aa4444
	ColorGopherBrown = "\033[38;5;95m"  // #85696D

	ColorViolet  = "\033[38;5;135m" // #af5fff
	ColorMagenta = "\033[38;5;201m" // #ff5fff
	ColorPink    = "\033[38;5;218m" // #ffafd7

	ColorReset = "\033[0m"
)

// (OLD) ColorPaletteHTML maps color codes to HTML colors.
var ColorPaletteHTML = map[string]string{
	ColorPaleGray:  "#B0BEC5", // #B0BEC5
	ColorSlateGray: "#A0A8B3", // #A0A8B3
	ColorDimGray:   "#5F6368", // #5F6368
	ColorDarkGray:  "#444444", // #444444
	ColorCharcoal:  "#262626", // #262626

	ColorLime:       "#80ff80", // #80ff80
	ColorSkyBlue:    "#55C1E7", // #55C1E7
	ColorMutedBlue:  "#73B3D5", // #73B3D5
	ColorLightTeal:  "#46A4B0", // #46A4B0
	ColorBrightCyan: "#00D5EF", // #00D5EF
	ColorGoBlue:     "#00ADD8", // #00ADD8
	ColorDarkTeal:   "#005F5F", // #005F5F
	ColorDarkGoBlue: "#005F87", // #005F87

	ColorSeafoamGreen: "#60DBD4", // #60DBD4
	ColorGreen:        "#00af00", // #00af00

	ColorGoldenrod:   "#FFD54F", // #FFD54F
	ColorCoralRed:    "#F46C5E", // #F46C5E
	ColorRed:         "#d70000", // #d70000
	ColorDarkRed:     "#aa4444", // #aa4444
	ColorGopherBrown: "#85696D", // #85696D

	ColorViolet:  "#af5fff", // #af5fff
	ColorMagenta: "#ff5fff", // #ff5fff
	ColorPink:    "#ffafd7", // #ffafd7
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
	ShowTypes           bool
	UseColors           bool
	TrackReferences     bool
	HTMLtagToken        string
	HTMLtagSection      string
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
			meta := fmt.Sprintf(" |R=%d|", runeCount)
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
		length += 2 // braces
		if d.config.ShowMetaInformation {
			if v.Kind() == reflect.Slice && v.Len() != v.Cap() {
				length += len(fmt.Sprintf("|L:%d C:%d| ", v.Len(), v.Cap()))
			} else {
				length += len(fmt.Sprintf("|%d| ", v.Len()))
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
		length += 2 // braces
		if d.config.ShowMetaInformation {
			length += len(fmt.Sprintf("|%d| ", v.Len()))
		}
		for i, key := range v.MapKeys() {
			if i > 0 {
				length += 2 // comma and space
			}
			val := v.MapIndex(key)

			length += d.estimatedInlineLength(key) + 4 + d.estimatedInlineLength(val) // key => val
			if d.config.ShowTypes {
				length += len(val.Type().String()) + 1 // type len + whitespace
			}
		}
		return length

	case reflect.Struct:
		length += 2 // braces
		t := v.Type()
		for i := range v.NumField() {
			if i > 0 {
				length += 2 // comma and space
			}
			name := t.Field(i).Name
			length += len(name) + 4 + d.estimatedInlineLength(v.Field(i)) // Name => val
			if d.config.ShowTypes {
				length += len(v.Field(i).Type().String()) + 1 // type len + whitespace
			}
		}
		return length

	default:
		return 10 // fallback
	}
}

func (d *Dumper) isSimpleMapKey(k reflect.Value) bool {
	// If simple or complex num or if estimated length is small, map key is "simple"
	if isSimpleValue(k) || k.Kind() == reflect.Complex64 || k.Kind() == reflect.Complex128 {
		return true
	} else {
		return d.estimatedInlineLength(k) <= d.config.MaxInlineLength
	}
}

// Returns a string representation for a value type (and handle any type)
func (d *Dumper) formatType(v reflect.Value, isInCollection bool) string {
	if !d.config.ShowTypes {
		return ""
	}

	if !v.IsValid() {
		return d.ApplyFormat(ColorDarkGray, "invalid")
	}

	// print element type signature
	vKind := v.Kind()
	expectedType := ""
	if vKind == reflect.Array || vKind == reflect.Slice || vKind == reflect.Map || vKind == reflect.Struct || vKind == reflect.Interface {
		expectedType = " " + d.ApplyFormat(ColorDarkGray, v.Type().String())
	} else if !isInCollection {
		expectedType = " " + d.ApplyFormat(ColorDarkGray, v.Type().String())
	}

	// if element type is just "any", print the actual variable type
	actualType := ""
	if vKind == reflect.Interface && v.Type().NumMethod() == 0 && !v.IsNil() {
		actualType = d.ApplyFormat(ColorDarkGray, "("+v.Elem().Type().String()+")")
	}
	formattedType := expectedType + actualType

	// Modernize the 'interface {}' to 'any'
	formattedType = strings.ReplaceAll(formattedType, "interface {}", "any")
	return formattedType
}

func (d *Dumper) formatMapKeyAsIndex(k reflect.Value) string {
	if d.isSimpleMapKey(k) {
		switch k.Kind() {
		case reflect.String:
			keyFormatted := strconv.Quote(k.Interface().(string))
			return keyFormatted
		default:
			keyFormatted := fmt.Sprintf("%v", k.Interface())
			return keyFormatted
		}
	}
	// TODO: should be summarizeKey(k)
	keyFormatted := fmt.Sprintf("%v", k.Interface())
	return keyFormatted

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

	headerTitle := d.ApplyFormat(ColorSlateGray, "[>]") + " " + d.ApplyFormat(ColorGoBlue, govarFuncName)
	headerLocation := d.ApplyFormat(ColorSlateGray, fmt.Sprintf("  ⟵  %s:%d", relPath, line))
	header := headerTitle + headerLocation
	fmt.Fprintln(out, header)
}

// renderAllValues writes all the values to the tabwriter, handling references and indentation.
func (d *Dumper) renderAllValues(tw *tabwriter.Writer, vs ...any) {
	d.referenceMap = map[uintptr]int{} // reset each time
	visited := map[uintptr]bool{}
	for _, v := range vs {
		rv := reflect.ValueOf(v)
		rv = makeAddressable(rv)

		// Render value's type signature
		fmt.Fprint(tw, d.ApplyFormat(ColorDarkGray, d.formatType(rv, false)))
		// On the zero level, if types are ON, render the "mapping to" symbol
		if d.config.ShowTypes {
			fmt.Fprint(tw, " => ")
		}
		// Render the value itself
		d.renderValue(tw, rv, 0, visited)

		fmt.Fprintln(tw)
	}
}

// renderValue recursively writes the value with indentation and handles references.
func (d *Dumper) renderValue(tw *tabwriter.Writer, v reflect.Value, level int, visited map[uintptr]bool) {
	if level > d.config.MaxDepth {
		fmt.Fprint(tw, d.ApplyFormat(ColorSlateGray, "… (max depth reached)"))
		return
	}
	if !v.IsValid() {
		fmt.Fprint(tw, d.ApplyFormat(ColorRed, "<invalid>"))
		return
	}

	if isNil(v) {
		fmt.Fprint(tw, d.ApplyFormat(ColorCoralRed, "<nil>"))
		return
	}

	// check for std fmt.Stringer interface representation
	if str := d.asStringerInterface(v); str != "" {
		if d.config.ShowMetaInformation {
			fmt.Fprint(tw, d.ApplyFormat(ColorDimGray, "|⧉ Stringer| "))
		}
		fmt.Fprint(tw, str)
		return
	}

	// check for std error interface representation
	if str := d.asErrorInterface(v); str != "" {
		if d.config.ShowMetaInformation {
			fmt.Fprint(tw, d.ApplyFormat(ColorDimGray, "|⧉ error| "))
		}
		fmt.Fprint(tw, str)
		return
	}

	switch v.Kind() {
	case reflect.Ptr:
		// If a pointer type is addressable and known, show a reference marker
		// If a pointer type is addressable and new, store it in the reference map
		if v.CanAddr() {
			ptr := v.Pointer()
			if id, ok := d.referenceMap[ptr]; ok {
				fmt.Fprintf(tw, d.ApplyFormat(ColorSlateGray, "↩︎ &%d"), id)
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
		fmt.Fprint(tw, d.ApplyFormat(ColorSlateGray, fmt.Sprintf("unsafe.Pointer(%#x)", v.Pointer())))
	case reflect.Bool:
		if v.Bool() {
			fmt.Fprint(tw, d.ApplyFormat(ColorSeafoamGreen, "true"))
		} else {
			fmt.Fprint(tw, d.ApplyFormat(ColorCoralRed, "false"))
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fmt.Fprint(tw, d.ApplyFormat(ColorSkyBlue, fmt.Sprint(v.Int())))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		fmt.Fprint(tw, d.ApplyFormat(ColorSkyBlue, fmt.Sprint(v.Uint())))
	case reflect.Float32, reflect.Float64:
		fmt.Fprint(tw, d.ApplyFormat(ColorSkyBlue, fmt.Sprintf("%f", v.Float())))
	case reflect.Complex64, reflect.Complex128:
		fmt.Fprint(tw, d.ApplyFormat(ColorSkyBlue, fmt.Sprintf("%v", v.Complex())))
	case reflect.String:
		strLen := utf8.RuneCountInString(v.String())
		str := d.stringEscape(v.String())
		str = d.ApplyFormat(ColorGoldenrod, `"`) + d.ApplyFormat(ColorLime, str) + d.ApplyFormat(ColorGoldenrod, `"`)
		if d.config.ShowMetaInformation {
			fmt.Fprint(tw, d.ApplyFormat(ColorDimGray, fmt.Sprintf("|R:%d| ", strLen)))
		}
		fmt.Fprint(tw, str)
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
			symbol = d.ApplyFormat(ColorDarkGoBlue, symbol)
			fieldName := d.ApplyFormat(ColorLightTeal, field.Name)
			if !d.shouldRenderInline(v) {
				// print visibility and symbol name, with indent
				d.renderIndent(tw, level+1, symbol+fieldName)
			} else {
				// inline render of the field
				fmt.Fprintf(tw, symbol+fieldName)
			}
			// print field type signature
			formattedType := d.formatType(fieldVal, false)
			fmt.Fprintf(tw, "	%s	=> ", formattedType)

			// Try the stringer interface on this struct field first
			if str := d.asStringerInterface(fieldVal); str != "" {
				if d.config.ShowMetaInformation {
					fmt.Fprint(tw, d.ApplyFormat(ColorDimGray, "|⧉ Stringer| "))
				}
				fmt.Fprint(tw, str)
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
			mapLen := fmt.Sprintf("|%d| ", v.Len())
			fmt.Fprint(tw, d.ApplyFormat(ColorDimGray, mapLen))
		}

		fmt.Fprint(tw, "[")
		if !d.shouldRenderInline(v) {
			fmt.Fprintln(tw)
		}

		keys := v.MapKeys()
		for i, key := range keys {
			if i >= d.config.MaxItems {
				d.renderIndent(tw, level+1, d.ApplyFormat(ColorSlateGray, "… (truncated)"))
				break
			}

			// keyStr := fmt.Sprintf("%v", key.Interface())
			keyStr := d.formatMapKeyAsIndex(key)

			// print element type signature
			formattedType := d.formatType(v.MapIndex(key), true)

			if !d.shouldRenderInline(v) {
				// indent, render key and type
				d.renderIndent(tw, level+1, fmt.Sprintf("%s%s	=> ", d.ApplyFormat(ColorViolet, keyStr), formattedType))
				// recursively print the array value itself, increase indent level
				d.renderValue(tw, v.MapIndex(key), level+1, visited)
			} else {
				// do not indent, render key and type
				fmt.Fprintf(tw, "%s%s	=> ", d.ApplyFormat(ColorViolet, keyStr), formattedType)
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
		fmt.Fprint(tw, "]")
	case reflect.Slice, reflect.Array:
		if d.config.ShowMetaInformation {
			var listLen string
			if v.Kind() == reflect.Array {
				listLen = fmt.Sprintf("|%d| ", v.Len())
			} else {
				if v.Len() == v.Cap() {
					listLen = fmt.Sprintf("|%d| ", v.Len())
				} else {
					listLen = fmt.Sprintf("|L:%d C:%d| ", v.Len(), v.Cap())
				}

			}
			fmt.Fprint(tw, d.ApplyFormat(ColorDimGray, listLen))
		}
		fmt.Fprint(tw, "[")
		if !d.shouldRenderInline(v) {
			fmt.Fprintln(tw)
		}
		for i := range v.Len() {
			if i >= d.config.MaxItems {
				d.renderIndent(tw, level+1, d.ApplyFormat(ColorSlateGray, "… (truncated)\n"))
				break
			}
			// print element type signature
			formattedType := d.formatType(v.Index(i), true)
			indexSymbol := d.ApplyFormat(ColorDarkTeal, fmt.Sprintf("%d", i))
			if !d.shouldRenderInline(v) {
				// indent, render index (and type)
				d.renderIndent(tw, level+1, fmt.Sprintf("%s%s => ", indexSymbol, formattedType))
				// recursively print the array value itself, increase indent level
				d.renderValue(tw, v.Index(i), level+1, visited)
			} else {
				// do not indent, render index (and type)
				fmt.Fprintf(tw, "%s%s => ", indexSymbol, formattedType)
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

		fmt.Fprint(tw, "]")
	case reflect.Func:
		funName := d.ApplyFormat(ColorLightTeal, getFunctionName(v))
		if d.config.ShowMetaInformation {
			funMeta := d.ApplyFormat(ColorDimGray, fmt.Sprintf("|func@%#x| ", v.Pointer()))
			fmt.Fprint(tw, funMeta)
		}
		fmt.Fprint(tw, funName)
	case reflect.Chan:
		if v.IsNil() {
			fmt.Fprint(tw, d.ApplyFormat(ColorCoralRed, "<nil>"))
		} else {
			symbol := d.ApplyFormat(ColorGoldenrod, "⮁") // ▲ 🠕 ⯭ ▼ ⯯ ▦
			chDir := v.Type().ChanDir().String()
			if chDir == "chan<-" {
				symbol = d.ApplyFormat(ColorGoBlue, "🡹")
			} else if chDir == "<-chan" {
				symbol = d.ApplyFormat(ColorGreen, "🢃")
			}
			if d.config.ShowMetaInformation {
				chBuffStr := d.ApplyFormat(ColorDimGray, fmt.Sprintf("|B:%d| ", v.Cap()))
				fmt.Fprint(tw, chBuffStr)
			}
			fmt.Fprintf(tw, "%s %s%s", symbol, d.ApplyFormat(ColorPink, "chan@"), d.ApplyFormat(ColorLightTeal, fmt.Sprintf("%#x", v.Pointer())))
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
		symbol := d.ApplyFormat(ColorDarkTeal, "⦿ ")
		methodName := d.ApplyFormat(ColorMutedBlue, m.Name)
		methodType := "	" + d.ApplyFormat(ColorDarkGray, m.Func.Type().String())
		d.renderIndent(tw, level, symbol+methodName+methodType)
		if d.config.ShowMetaInformation {
			fmt.Fprint(tw, d.ApplyFormat(ColorDimGray, " |Method|"))
		}
		fmt.Fprintln(tw)
	}
}

// asStringer checks if the value implements fmt.Stringer and returns its string representation.
func (d *Dumper) asStringerInterface(v reflect.Value) string {
	val := v
	if !val.CanInterface() {
		val = forceExported(val)
	}
	if val.CanInterface() {
		if s, ok := val.Interface().(fmt.Stringer); ok {
			rv := reflect.ValueOf(s)
			if rv.Kind() == reflect.Ptr && rv.IsNil() {
				return d.ApplyFormat(ColorCoralRed, "<nil>")
			}
			str := d.stringEscape(s.String())
			str = d.ApplyFormat(ColorGoldenrod, `"`) + d.ApplyFormat(ColorLime, str) + d.ApplyFormat(ColorGoldenrod, `"`)
			return str
		}
	}
	return ""
}

// asErrorer checks if the value implements fmt.Stringer and returns its string representation.
func (d *Dumper) asErrorInterface(v reflect.Value) string {
	val := v
	if !val.CanInterface() {
		val = forceExported(val)
	}
	if val.CanInterface() {
		if e, ok := val.Interface().(error); ok {
			rv := reflect.ValueOf(e)
			if rv.Kind() == reflect.Ptr && rv.IsNil() {
				return d.ApplyFormat(ColorCoralRed, "<nil>")
			}
			str := d.stringEscape(e.Error())
			str = d.ApplyFormat(ColorGoldenrod, `"`) + d.ApplyFormat(ColorCoralRed, str) + d.ApplyFormat(ColorGoldenrod, `"`)
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
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32, reflect.Float64,
		reflect.String:
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
