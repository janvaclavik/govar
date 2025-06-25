package govar

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"text/tabwriter"
	"unicode/utf8"
)

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
	ShowHexdump         bool
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
			meta := fmt.Sprintf(" |R:%d|", runeCount)
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
			length += 2 + len(name) + 4 + d.estimatedInlineLength(v.Field(i)) // Indicator Name => val
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
	if vKind == reflect.Interface {
		expectedType = d.ApplyFormat(ColorDarkGray, "⧉ "+v.Type().String())
	} else if vKind == reflect.Array || vKind == reflect.Slice || vKind == reflect.Map || vKind == reflect.Struct {
		expectedType = d.ApplyFormat(ColorDarkGray, v.Type().String())
	} else if !isInCollection {
		expectedType = d.ApplyFormat(ColorDarkGray, v.Type().String())
	}

	// if element type is an interface we can show the actual variable type
	actualType := ""
	if vKind == reflect.Interface && !v.IsNil() {
		actualType = d.ApplyFormat(ColorDarkGray, "("+v.Elem().Type().String()+")")
	}
	formattedType := expectedType + actualType

	// Modernize the 'interface {}' to 'any'
	formattedType = strings.ReplaceAll(formattedType, "interface {}", "any")
	return formattedType
}

func (d *Dumper) formatMapKeyAsIndex(k reflect.Value) string {
	var keyFormatted string
	if d.isSimpleMapKey(k) {
		switch k.Kind() {
		case reflect.String:
			keyFormatted = strconv.Quote(k.Interface().(string))
		case reflect.Interface:
			if k.Type().NumMethod() == 0 {
				// If the map key is was an "any" type, but a really a string underneath, format is as a string
				if str, ok := k.Interface().(string); ok {
					keyFormatted = strconv.Quote(str)
				} else {
					// Was any interface, but not a string...
					keyFormatted = fmt.Sprintf("%v", k.Interface())
				}
			} else {
				// Other kinds of interfaces
				keyFormatted = fmt.Sprintf("%v", k.Interface())
			}
		default:
			// Any other types
			keyFormatted = fmt.Sprintf("%v", k.Interface())
		}
	} else {
		// TODO: Here we should use a new summarizeKey(k) for complex, structured map keys
		keyFormatted = fmt.Sprintf("%v", k.Interface())
	}

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

	headerTitle := d.ApplyFormat(ColorGoBlue, "[>] "+govarFuncName)
	headerLocation := d.ApplyFormat(ColorSlateGray, fmt.Sprintf("  ⟵  %s:%d", relPath, line))
	header := headerTitle + headerLocation
	fmt.Fprintln(out, header)
}

func (d *Dumper) metaHint(msg string, ico string) string {
	if ico != "" {
		return d.ApplyFormat(ColorDimGray, fmt.Sprintf("|%s %s| ", ico, msg))
	}
	return d.ApplyFormat(ColorDimGray, fmt.Sprintf("|%s| ", msg))
}

// renderAllValues writes all the values to the tabwriter, handling references and indentation.
func (d *Dumper) renderAllValues(tw *tabwriter.Writer, vs ...any) {
	d.referenceMap = map[uintptr]int{} // reset each time
	visited := map[uintptr]bool{}
	for _, v := range vs {
		rv := reflect.ValueOf(v)
		rv = makeAddressable(rv)

		// Render value's type signature
		fmt.Fprint(tw, d.formatType(rv, false))
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

	if v.Kind() != reflect.Interface {
		// check for concrete interface (std fmt.Stringer) representation
		if str := d.asStringerInterface(v); str != "" {
			if d.config.ShowMetaInformation {
				fmt.Fprint(tw, d.metaHint("as Stringer", ""))
			}
			fmt.Fprint(tw, str+" ")
			return
		}

		// check for concrete interface (std error) representation
		if str := d.asErrorInterface(v); str != "" {
			if d.config.ShowMetaInformation {
				fmt.Fprint(tw, d.metaHint("as error", ""))
			}
			fmt.Fprint(tw, str+" ")
			return
		}
	}

	switch v.Kind() {
	case reflect.Ptr:
		d.renderPointer(tw, v, level, visited)
	case reflect.Interface:
		// TODO: ...
		// Continue with rendering the value that the interface contains
		d.renderValue(tw, v.Elem(), level, visited)
	case reflect.UnsafePointer:
		fmt.Fprint(tw, d.ApplyFormat(ColorSlateGray, fmt.Sprintf("unsafe.Pointer(%#x)", v.Pointer())))
	case reflect.Bool:
		d.renderBool(tw, v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fmt.Fprint(tw, d.ApplyFormat(ColorSkyBlue, fmt.Sprint(v.Int())))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		fmt.Fprint(tw, d.ApplyFormat(ColorSkyBlue, fmt.Sprint(v.Uint())))
	case reflect.Float32, reflect.Float64:
		fmt.Fprint(tw, d.ApplyFormat(ColorSkyBlue, fmt.Sprintf("%f", v.Float())))
	case reflect.Complex64, reflect.Complex128:
		fmt.Fprint(tw, d.ApplyFormat(ColorSkyBlue, fmt.Sprintf("%v", v.Complex())))
	case reflect.String:
		d.renderString(tw, v)
	case reflect.Struct:
		d.renderStruct(tw, v, level, visited)
	case reflect.Map:
		d.renderMap(tw, v, level, visited)
	case reflect.Slice, reflect.Array:
		d.renderArrayOrSlice(tw, v, level, visited)
	case reflect.Func:
		d.renderFunc(tw, v)
	case reflect.Chan:
		d.renderChan(tw, v)
	default:
		// Should be unreachable - all reflect.Kind cases are handled
		fmt.Fprintln(tw, "[WARNING] unknown reflect.Kind, rendering not implemented")
	}
}

func (d *Dumper) renderPointer(tw *tabwriter.Writer, v reflect.Value, level int, visited map[uintptr]bool) {
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
}

func (d *Dumper) renderChan(tw *tabwriter.Writer, v reflect.Value) {
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
			fmt.Fprint(tw, d.metaHint(fmt.Sprintf("B:%d", v.Cap()), ""))
		}
		fmt.Fprintf(tw, "%s %s%s", symbol, d.ApplyFormat(ColorPink, "chan@"), d.ApplyFormat(ColorLightTeal, fmt.Sprintf("%#x", v.Pointer())))
	}
}

func (d *Dumper) renderFunc(tw *tabwriter.Writer, v reflect.Value) {
	funName := d.ApplyFormat(ColorLightTeal, getFunctionName(v))
	if d.config.ShowMetaInformation {
		fmt.Fprint(tw, d.metaHint(fmt.Sprintf("func@%#x", v.Pointer()), ""))
	}
	fmt.Fprint(tw, funName)
}

func (d *Dumper) renderArrayOrSlice(tw *tabwriter.Writer, v reflect.Value, level int, visited map[uintptr]bool) {
	if d.config.ShowMetaInformation {
		var listLen string
		if v.Kind() == reflect.Array {
			listLen = fmt.Sprintf("%d", v.Len())
		} else {
			if v.Len() == v.Cap() {
				listLen = fmt.Sprintf("%d", v.Len())
			} else {
				listLen = fmt.Sprintf("L:%d C:%d", v.Len(), v.Cap())
			}
		}

		fmt.Fprint(tw, d.metaHint(listLen, ""))
	}
	fmt.Fprint(tw, "[")
	if !d.shouldRenderInline(v) {
		fmt.Fprintln(tw)
	}

	if d.config.ShowHexdump && v.Type().Elem().Kind() == reflect.Uint8 {
		d.renderHexdump(tw, v, level)
	} else {

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
				renderIndex := ""
				if formattedType != "" {
					renderIndex = fmt.Sprintf("%s %s\t=> ", indexSymbol, formattedType)
				} else {
					renderIndex = fmt.Sprintf("%s => ", indexSymbol)
				}
				d.renderIndent(tw, level+1, renderIndex)
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
	}

	if !d.shouldRenderInline(v) {
		d.renderIndent(tw, level, "")
	}

	fmt.Fprint(tw, "]")
}

func (d *Dumper) renderHexdump(tw *tabwriter.Writer, v reflect.Value, level int) {
	// using std package hex
	// Safe fallback: Manual conversion to addressable array (cause v.Bytes() might not work)
	content := toAddressableByteSlice(v)
	// fmt.Printf("%s", hex.Dump(content))
	lines := strings.Split(hex.Dump(content), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}
		// Example line:
		// 00000000  48 65 6c 6c 6f 2c 20 57  6f 72 6c 64 21 0a 00 ff  |Hello, World!...|

		// Split into three main parts:
		if len(line) < 10 {
			fmt.Println(line) // fallback
			continue
		}

		offsetPart := line[:10]
		hexPart := line[10:58] // includes 2 spaces between 8-byte blocks
		asciiPart := ""
		if idx := strings.Index(line, "  |"); idx != -1 {
			asciiPart = line[idx:]
		}
		// Print indent
		d.renderIndent(tw, level+1, "")
		// Print with color
		fmt.Fprintf(tw, "%s%s%s\n",
			d.ApplyFormat(ColorDarkTeal, offsetPart),
			d.ApplyFormat(ColorSkyBlue, hexPart),
			d.ApplyFormat(ColorLime, asciiPart),
		)
	}
}

func (d *Dumper) renderStruct(tw *tabwriter.Writer, v reflect.Value, level int, visited map[uintptr]bool) {
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
		formattedType := d.formatType(fieldVal, false)

		if !d.shouldRenderInline(v) {
			fieldRender := fmt.Sprintf("%s\t=>\t", symbol+fieldName)
			if formattedType != "" {
				fieldRender = fmt.Sprintf("%s\t%s\t=>\t", symbol+fieldName, formattedType)
			}
			// print visibility and symbol name, with indent
			d.renderIndent(tw, level+1, fieldRender)
		} else {
			// inline render of the field
			fieldRender := fmt.Sprintf("%s => ", symbol+fieldName)
			if formattedType != "" {
				fieldRender = fmt.Sprintf("%s %s => ", symbol+fieldName, formattedType)
			}
			fmt.Fprint(tw, fieldRender)
		}

		// recursively render the field value itself
		if !d.shouldRenderInline(v) {
			d.renderValue(tw, fieldVal, level+1, visited)
		} else {
			// inline render
			d.renderValue(tw, fieldVal, level, visited)
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
}

func (d *Dumper) renderMap(tw *tabwriter.Writer, v reflect.Value, level int, visited map[uintptr]bool) {
	if d.config.ShowMetaInformation {
		mapLen := fmt.Sprintf("%d", v.Len())
		d.metaHint(mapLen, "")
		fmt.Fprint(tw, d.metaHint(mapLen, ""))
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
			keyRender := fmt.Sprintf("%s\t=> ", d.ApplyFormat(ColorDarkTeal, keyStr))
			if formattedType != "" {
				keyRender = fmt.Sprintf("%s\t%s\t=> ", d.ApplyFormat(ColorDarkTeal, keyStr), formattedType)
			}
			d.renderIndent(tw, level+1, keyRender)
			// recursively print the array value itself, increase indent level
			d.renderValue(tw, v.MapIndex(key), level+1, visited)
		} else {
			// do not indent, render key and type
			fmt.Fprintf(tw, "%s %s => ", d.ApplyFormat(ColorDarkTeal, keyStr), formattedType)
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
}

func (d *Dumper) renderString(tw *tabwriter.Writer, v reflect.Value) {
	strLen := utf8.RuneCountInString(v.String())
	str := d.stringEscape(v.String())
	str = d.ApplyFormat(ColorGoldenrod, `"`) + d.ApplyFormat(ColorLime, str) + d.ApplyFormat(ColorGoldenrod, `"`)
	if d.config.ShowMetaInformation {
		fmt.Fprint(tw, d.metaHint(fmt.Sprintf("R:%d", strLen), ""))
	}
	fmt.Fprint(tw, str)
}

func (d *Dumper) renderBool(tw *tabwriter.Writer, v reflect.Value) {
	if v.Bool() {
		fmt.Fprint(tw, d.ApplyFormat(ColorGreen, "true"))
	} else {
		fmt.Fprint(tw, d.ApplyFormat(ColorCoralRed, "false"))
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
		methodType := d.formatType(m.Func, false)
		renderMethod := fmt.Sprintf("%s%s\t %s", symbol, methodName, methodType)
		if methodType == "" {
			renderMethod = fmt.Sprintf("%s%s", symbol, methodName)
		}
		d.renderIndent(tw, level, renderMethod)
		fmt.Fprintln(tw)
	}
}

// asStringer checks if the value implements the fmt.Stringer and returns its
// string representation.
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

// asErrorInterface checks if the value implements the std error and returns
// its string representation.
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
