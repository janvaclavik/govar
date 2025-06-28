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
	"unicode/utf8"
)

// DumperConfig holds configuration parameters for the Dumper.
// These control output formatting, depth, type information, etc.
type DumperConfig struct {
	IndentWidth         int    // Number of spaces to use per indentation level.
	MaxDepth            int    // Maximum levels of nested structures to print.
	MaxItems            int    // Maximum number of items to print per slice/map.
	MaxStringLen        int    // Maximum string length before truncation.
	MaxInlineLength     int    // Maximum inline width before switching to block format.
	ShowTypes           bool   // Whether to show type names.
	UseColors           bool   // Whether to apply ANSI colors to output.
	TrackReferences     bool   // Track shared references to detect cycles.
	HTMLtagToken        string // HTML span tag class used for syntax tokens.
	HTMLtagSection      string // HTML span tag class used for value sections.
	EmbedTypeMethods    bool   // Include exported methods from embedded types.
	ShowMetaInformation bool   // Show metadata such as string lengths or slice capacities.
	ShowHexdump         bool   // Show byte slices as hexdump when applicable.
}

// Dumper is a configurable structure-aware pretty printer for Go values.
//
// It provides colorized and formatted output to stdout, io.Writer, or string formats,
// making it easier to introspect complex values during debugging, logging, or inspection.
// Dumper should be constructed via NewDumper with a DumperConfig. For common use cases,
// high-level helpers like Dump, Sdump, and Fdump are available in the govar API.
type Dumper struct {
	nextRefID       int
	referenceCounts map[uintptr]int
	referenceMap    map[uintptr]int
	config          DumperConfig
	Formatter
}

// NewDumper creates a new Dumper with the provided configuration.
func NewDumper(cfg DumperConfig) *Dumper {
	return &Dumper{nextRefID: 1, referenceMap: map[uintptr]int{}, config: cfg, Formatter: PlainFormatter{}}
}

// Die dumps the given values and immediately terminates the program.
func (d *Dumper) Die(vs ...any) {
	Dump(vs...)
	os.Exit(1)
}

// Dump prints values to stdout using the configured formatting.
func (d *Dumper) Dump(vs ...any) {
	// Enable coloring
	if d.config.UseColors {
		d.Formatter = ANSIcolorFormatter{}
	} else {
		d.Formatter = PlainFormatter{}
	}
	sb := &strings.Builder{}
	d.renderHeader(sb)
	d.renderAllValues(sb, vs...)
	fmt.Fprintln(os.Stdout, sb.String())
}

// Fdump writes values to the given io.Writer using the configured formatting.
func (d *Dumper) Fdump(w io.Writer, vs ...any) {
	// Enable coloring
	if d.config.UseColors {
		d.Formatter = ANSIcolorFormatter{}
	} else {
		d.Formatter = PlainFormatter{}
	}
	sb := &strings.Builder{}
	d.renderHeader(sb)
	d.renderAllValues(sb, vs...)
	fmt.Fprintln(w, sb.String())
}

// Sdump returns a string containing the formatted values.
func (d *Dumper) Sdump(vs ...any) string {
	// Enable coloring
	if d.config.UseColors {
		d.Formatter = ANSIcolorFormatter{}
	} else {
		d.Formatter = PlainFormatter{}
	}
	sb := &strings.Builder{}
	d.renderHeader(sb)
	d.renderAllValues(sb, vs...)
	return sb.String()
}

// SdumpHTML returns an HTML-formatted dump wrapped in a <pre> block.
func (d *Dumper) SdumpHTML(vs ...any) string {
	// Enable HTML coloring
	d.Formatter = HTMLformatter{HTMLtagToken: d.config.HTMLtagToken, UseColors: d.config.UseColors}

	sb := &strings.Builder{}
	sb.WriteString(fmt.Sprintf(`<%s class="govar" style="background-color:black; color:white; padding:4px; border-radius: 4px">`+"\n", d.config.HTMLtagSection))
	d.renderHeader(sb)
	d.renderAllValues(sb, vs...)
	sb.WriteString(fmt.Sprintf("</%s", d.config.HTMLtagSection))
	return sb.String()
}

// asStringerInterface tries to format a value implementing fmt.Stringer.
func (d *Dumper) asStringerInterface(v reflect.Value) string {
	val := v
	if !val.CanInterface() {
		val = tryExport(val)
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

// asErrorInterface tries to format a value implementing the error interface.
func (d *Dumper) asErrorInterface(v reflect.Value) string {
	val := v
	if !val.CanInterface() {
		val = tryExport(val)
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

	for i := range v.NumField() {
		if !isSimpleValue(v.Field(i)) {
			return false
		}
	}
	return true
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

func (d *Dumper) formatArrayOrSlice(v reflect.Value, level int, visited map[uintptr]bool) string {
	sb := &strings.Builder{}

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
		fmt.Fprint(sb, d.metaHint(listLen, ""))
	}

	fmt.Fprint(sb, "[")

	if d.shouldRenderInline(v) {
		// INLINE RENDER
		for i := range v.Len() {
			if i >= d.config.MaxItems {
				d.renderIndent(sb, level+1, d.ApplyFormat(ColorSlateGray, "… (truncated)\n"))
				break
			}
			// print element type signature
			formattedType := d.formatType(v.Index(i), true)
			indexSymbol := d.ApplyFormat(ColorDarkTeal, fmt.Sprintf("%d", i))

			fmt.Fprintf(sb, "%s%s => ", indexSymbol, formattedType)
			// recursively print the array value itself, same indent level
			d.renderValue(sb, v.Index(i), level, visited)

			if i != v.Len()-1 {
				fmt.Fprint(sb, ", ")
			}
		}

	} else {
		// BLOCK RENDER
		fmt.Fprintln(sb)

		// We might render a hexdump table
		if d.config.ShowHexdump && v.Type().Elem().Kind() == reflect.Uint8 {
			d.renderHexdump(sb, v, level)
		} else {
			// Not a hexdump, but a block array/slice

			// First we do a pre-pass and calculate the lengthiest type
			maxTypeLen := 0
			for i := range v.Len() {
				if i >= d.config.MaxItems {
					break
				}
				typeName := d.formatTypeNoColors(v.Index(i), true)
				if utf8.RuneCountInString(typeName) > maxTypeLen {
					maxTypeLen = utf8.RuneCountInString(typeName)
				}
			}

			for i := range v.Len() {
				if i >= d.config.MaxItems {
					d.renderIndent(sb, level+1, d.ApplyFormat(ColorSlateGray, "… (truncated)\n"))
					break
				}
				// print element type signature
				formattedType := d.formatType(v.Index(i), true)
				indexSymbol := d.ApplyFormat(ColorDarkTeal, fmt.Sprintf("%d", i))

				// block render
				renderIndex := ""
				if formattedType != "" {
					unformattedTypeLen := utf8.RuneCountInString(d.formatTypeNoColors(v.Index(i), true))
					paddedType := padRight(formattedType, unformattedTypeLen, maxTypeLen)
					renderIndex = fmt.Sprintf("%s %s => ", indexSymbol, paddedType)
				} else {
					renderIndex = fmt.Sprintf("%s => ", indexSymbol)
				}
				d.renderIndent(sb, level+1, renderIndex)
				// recursively print the array value itself, increase indent level
				d.renderValue(sb, v.Index(i), level+1, visited)

				fmt.Fprintln(sb)
			}
		}
		d.renderIndent(sb, level, "")
	}

	fmt.Fprint(sb, "]")

	return sb.String()
}

func (d *Dumper) formatChan(v reflect.Value) string {
	if v.IsNil() {
		return d.ApplyFormat(ColorCoralRed, "<nil>")
	} else {
		symbol := d.ApplyFormat(ColorGoldenrod, "⮁")
		chDir := v.Type().ChanDir().String()
		if chDir == "chan<-" {
			symbol = d.ApplyFormat(ColorGoBlue, "🡹")
		} else if chDir == "<-chan" {
			symbol = d.ApplyFormat(ColorGreen, "🢃")
		}
		result := ""
		if d.config.ShowMetaInformation {
			result = fmt.Sprint(d.metaHint(fmt.Sprintf("B:%d", v.Cap()), ""))
		}
		result = result + fmt.Sprintf("%s %s%s", symbol, d.ApplyFormat(ColorPink, "chan@"), d.ApplyFormat(ColorLightTeal, fmt.Sprintf("%#x", v.Pointer())))
		return result
	}
}

func (d *Dumper) formatFunc(v reflect.Value) string {
	funName := d.ApplyFormat(ColorLightTeal, getFunctionName(v))
	if d.config.ShowMetaInformation {
		funName = fmt.Sprint(d.metaHint(fmt.Sprintf("func@%#x", v.Pointer()), "")) + funName
	}
	return funName
}

func (d *Dumper) formatBool(v reflect.Value) string {
	if v.Bool() {
		return d.ApplyFormat(ColorGreen, "true")
	} else {
		return d.ApplyFormat(ColorCoralRed, "false")
	}
}

func (d *Dumper) formatMap(v reflect.Value, level int, visited map[uintptr]bool) string {
	sb := &strings.Builder{}

	if d.config.ShowMetaInformation {
		mapLen := fmt.Sprintf("%d", v.Len())
		d.metaHint(mapLen, "")
		fmt.Fprint(sb, d.metaHint(mapLen, ""))
	}

	// keys := v.MapKeys()
	sortedKeys := sortMapKeys(v)

	fmt.Fprint(sb, "[")

	if d.shouldRenderInline(v) {
		// INLINE RENDER
		for i, key := range sortedKeys {
			if i >= d.config.MaxItems {
				d.renderIndent(sb, level+1, d.ApplyFormat(ColorSlateGray, "… (truncated)"))
				break
			}

			keyStr := d.formatMapKeyAsIndex(key)

			// print element type signature
			formattedType := d.formatType(v.MapIndex(key), true)

			// inline render
			fmt.Fprintf(sb, "%s %s => ", d.ApplyFormat(ColorDarkTeal, keyStr), formattedType)
			// recursively print the array value itself, same indent level
			d.renderValue(sb, v.MapIndex(key), level, visited)

			if i != v.Len()-1 {
				fmt.Fprint(sb, ", ")
			}
		}

	} else {
		// BLOCK RENDER
		fmt.Fprintln(sb)

		// First we do a pre-pass and calculate the lengthiest key and type
		maxKeyLen := 0
		maxTypeLen := 0
		for i, key := range sortedKeys {
			if i >= d.config.MaxItems {
				break
			}
			keyStr := d.formatMapKeyAsIndex(key)
			if utf8.RuneCountInString(keyStr) > maxKeyLen {
				maxKeyLen = utf8.RuneCountInString(keyStr)
			}
			typeName := d.formatTypeNoColors(v.MapIndex(key), true)
			if utf8.RuneCountInString(typeName) > maxTypeLen {
				maxTypeLen = utf8.RuneCountInString(typeName)
			}
		}

		for i, key := range sortedKeys {
			if i >= d.config.MaxItems {
				d.renderIndent(sb, level+1, d.ApplyFormat(ColorSlateGray, "… (truncated)"))
				break
			}

			keyStr := d.formatMapKeyAsIndex(key)
			// print element type signature
			formattedType := d.formatType(v.MapIndex(key), true)

			keyRender := fmt.Sprintf("%s => ", padRight(keyStr, utf8.RuneCountInString(keyStr), maxKeyLen))

			if formattedType != "" {
				unformattedTypeLen := utf8.RuneCountInString(d.formatTypeNoColors(v.MapIndex(key), true))
				paddedKey := padRight(keyStr, utf8.RuneCountInString(keyStr), maxKeyLen)
				paddedType := padRight(formattedType, unformattedTypeLen, maxTypeLen)
				keyRender = fmt.Sprintf("%s  %s => ", d.ApplyFormat(ColorDarkTeal, paddedKey), paddedType)
			}

			d.renderIndent(sb, level+1, keyRender)
			// recursively print the array value itself, increase indent level
			d.renderValue(sb, v.MapIndex(key), level+1, visited)

			fmt.Fprintln(sb)
		}

		d.renderIndent(sb, level, "")
	}

	fmt.Fprint(sb, "]")

	return sb.String()
}

func (d *Dumper) formatString(v reflect.Value) string {
	strLen := utf8.RuneCountInString(v.String())
	str := d.stringEscape(v.String())
	str = d.ApplyFormat(ColorGoldenrod, `"`) + d.ApplyFormat(ColorLime, str) + d.ApplyFormat(ColorGoldenrod, `"`)
	if d.config.ShowMetaInformation {
		str = d.metaHint(fmt.Sprintf("R:%d", strLen), "") + str
	}
	return str
}

// Returns a string representation for a value type (and handle any type)
func (d *Dumper) formatType(v reflect.Value, isInCollection bool) string {
	if !d.config.ShowTypes {
		return ""
	}

	return d.ApplyFormat(ColorDarkGray, d.formatTypeNoColors(v, isInCollection))
}

// Returns a string representation for a value type (and handle any type)
func (d *Dumper) formatTypeNoColors(v reflect.Value, isInCollection bool) string {
	if !d.config.ShowTypes {
		return ""
	}

	if !v.IsValid() {
		return "invalid"
	}

	// print element type signature
	vKind := v.Kind()
	expectedType := ""
	if vKind == reflect.Interface {
		expectedType = "⧉ " + v.Type().String()
	} else if vKind == reflect.Array || vKind == reflect.Slice || vKind == reflect.Map || vKind == reflect.Struct {
		expectedType = v.Type().String()
	} else if !isInCollection {
		expectedType = v.Type().String()
	}

	// if element type is an interface we can show the actual variable type
	actualType := ""
	if vKind == reflect.Interface && !v.IsNil() {
		actualType = "(" + v.Elem().Type().String() + ")"
	}
	formattedType := expectedType + actualType

	// Modernize the 'interface {}' to 'any'
	formattedType = strings.ReplaceAll(formattedType, "interface {}", "any")
	return formattedType
}

func (d *Dumper) metaHint(msg string, ico string) string {
	if ico != "" {
		return d.ApplyFormat(ColorDimGray, fmt.Sprintf("|%s %s| ", ico, msg))
	}
	return d.ApplyFormat(ColorDimGray, fmt.Sprintf("|%s| ", msg))
}

// preScan recursively scans all possible references and builds
// the referenceMap
func (d *Dumper) preScan(v reflect.Value, visited map[uintptr]bool) {
	// Handle invalid or zero values early
	if !v.IsValid() {
		return
	}

	// If it's a pointer or interface, we want the underlying value
	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() {
			return
		}

		ptr := v.Pointer()
		d.referenceCounts[ptr]++
		if visited[ptr] {
			return
		}
		visited[ptr] = true
		d.preScan(v.Elem(), visited)
		return
	}

	// Recurse into composite types
	switch v.Kind() {
	case reflect.Interface:
		if v.IsNil() {
			return
		}

		// Unwrap interface first
		elem := v.Elem()
		d.preScan(elem, visited)
		return
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			d.preScan(v.Index(i), visited)
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			d.preScan(key, visited)
			d.preScan(v.MapIndex(key), visited)
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			if field.CanInterface() {
				d.preScan(field, visited)
			}
		}
	}
}

// renderAllValues writes all the values to the stringbuilder, handling references and indentation.
func (d *Dumper) renderAllValues(sb *strings.Builder, vs ...any) {

	if d.config.TrackReferences {
		d.referenceMap = map[uintptr]int{}    // reset each time
		d.referenceCounts = map[uintptr]int{} // reset each time
		d.nextRefID = 1

		// First pass with preScan(): assign reference IDs
		for _, v := range vs {
			rv := reflect.ValueOf(v)
			rv = makeAddressable(rv)
			d.preScan(rv, map[uintptr]bool{}) // fresh map for each top-level value
		}

		// After preScan, loop through referenceCounts and assign refIDs only to those that are shared
		// (more references than 1)
		for ptr, count := range d.referenceCounts {
			if count > 1 {
				d.referenceMap[ptr] = d.nextRefID
				d.nextRefID++
			}
		}
	}
	// Second pass: render the values
	visited := map[uintptr]bool{}
	for _, v := range vs {
		rv := reflect.ValueOf(v)
		rv = makeAddressable(rv)

		// Check for nil interface
		vType, tmpRv := checkNilInterface(v)

		// On the zero level, if types are ON, render the "mapping to" symbol
		if d.config.ShowTypes {
			if vType != "unknown" {
				vType = d.formatType(rv, false)
			} else {
				vType = d.ApplyFormat(ColorDarkGray, vType)
			}

			// Render value's type signature
			fmt.Fprint(sb, vType)
			fmt.Fprint(sb, " => ")
		}

		// Check for nil value at the top
		if tmpRv != "" {
			sb.WriteString(d.ApplyFormat(ColorCoralRed, tmpRv))
		} else {
			// Render the value recursively
			d.renderValue(sb, rv, 0, visited)
		}

		fmt.Fprintln(sb)
	}
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

func (d *Dumper) renderHexdump(sb *strings.Builder, v reflect.Value, level int) {
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
		hexPart := line[10:58] // includes 2 spaces besbeen 8-byte blocks
		asciiPart := ""
		if idx := strings.Index(line, "  |"); idx != -1 {
			asciiPart = line[idx:]
		}
		// Print indent
		d.renderIndent(sb, level+1, "")
		// Print with color
		fmt.Fprintf(sb, "%s%s%s\n",
			d.ApplyFormat(ColorDarkTeal, offsetPart),
			d.ApplyFormat(ColorSkyBlue, hexPart),
			d.ApplyFormat(ColorLime, asciiPart),
		)
	}
}

// renderIndent writes indented text to the stringbuilder.
func (d *Dumper) renderIndent(sb *strings.Builder, indentLevel int, text string) {
	fmt.Fprint(sb, strings.Repeat(" ", indentLevel*d.config.IndentWidth)+text)
}

func (d *Dumper) renderPointer(sb *strings.Builder, v reflect.Value, level int, visited map[uintptr]bool) {
	ptr := v.Pointer()

	if d.config.TrackReferences {
		// If this pointer is known (shared), use ↩︎ or anchor
		if id, ok := d.referenceMap[ptr]; ok {
			if visited[ptr] {
				// Seen already, render as ↩︎
				fmt.Fprintf(sb, d.ApplyFormat(ColorSlateGray, "↩︎ &%d"), id)
				return
			} else {
				// First time seeing it, mark visited and render anchor
				visited[ptr] = true
				fmt.Fprintf(sb, d.ApplyFormat(ColorGoldenrod, "&%d "), id)
			}
		}
	}
	// Continue with rendering the value that the pointer points to
	d.renderValue(sb, v.Elem(), level, visited)
}

func (d *Dumper) renderStruct(sb *strings.Builder, v reflect.Value, level int, visited map[uintptr]bool) {
	t := v.Type()
	visibleFields := reflect.VisibleFields(t)
	fmt.Fprint(sb, "{")

	if d.shouldRenderInline(v) {
		// INLINE RENDER
		for i := range t.NumField() {
			// fieldVal := v.FieldByIndex(field.Index)
			field := t.Field(i)
			fieldVal := v.Field(i)

			symbol := "⯀ "
			if field.PkgPath != "" {
				symbol = "🞏 "
				fieldVal = tryExport(fieldVal)
			}

			symbol = d.ApplyFormat(ColorDarkGoBlue, symbol)
			fieldName := d.ApplyFormat(ColorLightTeal, field.Name)
			formattedType := d.formatType(fieldVal, false)

			// inline render of the field
			fieldRender := fmt.Sprintf("%s => ", symbol+fieldName)
			if formattedType != "" {
				fieldRender = fmt.Sprintf("%s %s => ", symbol+fieldName, formattedType)
			}
			fmt.Fprint(sb, fieldRender)
			d.renderValue(sb, fieldVal, level, visited)

			if i != len(visibleFields)-1 {
				fmt.Fprint(sb, ", ")
			}
		}

	} else {
		// BLOCK RENDER
		fmt.Fprintln(sb)
		// First we do a pre-pass and calculate the lengthiest field and type
		maxKeyLen := 0
		maxTypeLen := 0
		for i := range t.NumField() {
			// fieldVal := v.FieldByIndex(field.Index)
			field := t.Field(i)
			fieldVal := v.Field(i)

			if field.PkgPath != "" {
				fieldVal = tryExport(fieldVal)
			}
			if utf8.RuneCountInString(field.Name) > maxKeyLen {
				maxKeyLen = utf8.RuneCountInString(field.Name)
			}
			typeName := d.formatTypeNoColors(fieldVal, false)
			if utf8.RuneCountInString(typeName) > maxTypeLen {
				maxTypeLen = utf8.RuneCountInString(typeName)
			}
		}
		if d.config.EmbedTypeMethods {
			// If embedded methods are ON, do a pre-pass on them too
			for _, m := range findTypeMethods(t) {
				methodName := m.Name
				if utf8.RuneCountInString(methodName) > maxKeyLen {
					maxKeyLen = utf8.RuneCountInString(methodName)
				}
			}
		}
		maxKeyLen += 2 // for visibility symbol

		// Now can render the fields
		for i := range t.NumField() {
			// fieldVal := v.FieldByIndex(field.Index)
			field := t.Field(i)
			fieldVal := v.Field(i)

			symbol := "⯀ "
			if field.PkgPath != "" {
				symbol = "🞏 "
				fieldVal = tryExport(fieldVal)
			}
			unformattedFieldLen := utf8.RuneCountInString(symbol + field.Name)
			unformattedTypeLen := utf8.RuneCountInString(d.formatTypeNoColors(fieldVal, false))

			symbol = d.ApplyFormat(ColorDarkGoBlue, symbol)
			fieldName := d.ApplyFormat(ColorLightTeal, field.Name)
			formattedType := d.formatType(fieldVal, false)

			// block render of the field
			fieldRender := fmt.Sprintf("%s => ", padRight(symbol+fieldName, unformattedFieldLen, maxKeyLen))
			if formattedType != "" {
				fieldRender = fmt.Sprintf("%s  %s => ", padRight(symbol+fieldName, unformattedFieldLen, maxKeyLen), padRight(formattedType, unformattedTypeLen, maxTypeLen))
			}
			// print visibility and symbol name, with indent
			d.renderIndent(sb, level+1, fieldRender)
			d.renderValue(sb, fieldVal, level+1, visited)
			fmt.Fprintln(sb)
		}
		// print all of struct's type methods (never inline)
		if d.config.EmbedTypeMethods {
			d.renderTypeMethods(sb, t, level+1, maxKeyLen)
		}
		d.renderIndent(sb, level, "")
	}
	fmt.Fprint(sb, "}")
}

func (d *Dumper) renderTypeMethods(sb *strings.Builder, t reflect.Type, level int, maxNameLen int) {
	for _, m := range findTypeMethods(t) {
		// print visibility and symbol name
		unformattedNameLen := utf8.RuneCountInString(m.Name) + 2
		symbol := d.ApplyFormat(ColorDarkTeal, "⦿ ")
		methodName := d.ApplyFormat(ColorMutedBlue, m.Name)
		methodType := d.formatType(m.Func, false)
		renderMethod := fmt.Sprintf("%s  %s", padRight(symbol+methodName, unformattedNameLen, maxNameLen), methodType)
		if methodType == "" {
			renderMethod = symbol + methodName
		}
		d.renderIndent(sb, level, renderMethod)
		fmt.Fprintln(sb)
	}
}

// renderValue recursively writes the value with indentation and handles references.
func (d *Dumper) renderValue(sb *strings.Builder, v reflect.Value, level int, visited map[uintptr]bool) {
	if level > d.config.MaxDepth {
		fmt.Fprint(sb, d.ApplyFormat(ColorSlateGray, "… (max depth reached)"))
		return
	}
	if !v.IsValid() {
		fmt.Fprint(sb, d.ApplyFormat(ColorRed, "<invalid>"))
		return
	}

	if isNil(v) {
		fmt.Fprint(sb, d.ApplyFormat(ColorCoralRed, "<nil>"))
		return
	}

	if v.Kind() != reflect.Interface {
		// check for concrete interface (std fmt.Stringer) representation
		if str := d.asStringerInterface(v); str != "" {
			if d.config.ShowMetaInformation {
				fmt.Fprint(sb, d.metaHint("as Stringer", ""))
			}
			fmt.Fprint(sb, str+" ")
			return
		}

		// check for concrete interface (std error) representation
		if str := d.asErrorInterface(v); str != "" {
			if d.config.ShowMetaInformation {
				fmt.Fprint(sb, d.metaHint("as error", ""))
			}
			fmt.Fprint(sb, str+" ")
			return
		}
	}

	switch v.Kind() {
	case reflect.Ptr:
		d.renderPointer(sb, v, level, visited)
	case reflect.Interface:
		// Continue with rendering the value that the interface contains
		d.renderValue(sb, v.Elem(), level, visited)
	case reflect.UnsafePointer:
		fmt.Fprint(sb, d.ApplyFormat(ColorSlateGray, fmt.Sprintf("unsafe.Pointer(%#x)", v.Pointer())))
	case reflect.Bool:
		renderVal := d.formatBool(v)
		d.wrapAndRender(sb, renderVal, v.Type(), level)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		renderVal := d.ApplyFormat(ColorSkyBlue, fmt.Sprint(v.Int()))
		d.wrapAndRender(sb, renderVal, v.Type(), level)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		renderVal := d.ApplyFormat(ColorSkyBlue, fmt.Sprint(v.Uint()))
		d.wrapAndRender(sb, renderVal, v.Type(), level)
	case reflect.Float32, reflect.Float64:
		renderVal := d.ApplyFormat(ColorSkyBlue, fmt.Sprintf("%f", v.Float()))
		d.wrapAndRender(sb, renderVal, v.Type(), level)
	case reflect.Complex64, reflect.Complex128:
		renderVal := d.ApplyFormat(ColorSkyBlue, fmt.Sprintf("%v", v.Complex()))
		d.wrapAndRender(sb, renderVal, v.Type(), level)
	case reflect.String:
		renderVal := d.formatString(v)
		d.wrapAndRender(sb, renderVal, v.Type(), level)
	case reflect.Struct:
		d.renderStruct(sb, v, level, visited)
	case reflect.Map:
		renderVal := d.formatMap(v, level, visited)
		d.wrapAndRender(sb, renderVal, v.Type(), level)
	case reflect.Slice, reflect.Array:
		renderVal := d.formatArrayOrSlice(v, level, visited)
		d.wrapAndRender(sb, renderVal, v.Type(), level)
	case reflect.Func:
		renderVal := d.formatFunc(v)
		d.wrapAndRender(sb, renderVal, v.Type(), level)
	case reflect.Chan:
		renderVal := d.formatChan(v)
		d.wrapAndRender(sb, renderVal, v.Type(), level)
	default:
		// Should be unreachable - all reflect.Kind cases are handled
		fmt.Fprintln(sb, "[WARNING] unknown reflect.Kind, rendering not implemented")
	}
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
		// If type method embedding is ON and type has methods, struct cannot be inline
		if d.config.EmbedTypeMethods && len(findTypeMethods(v.Type())) > 0 {
			return false
		}
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

func (d *Dumper) wrapAndRender(sb *strings.Builder, renderVal string, t reflect.Type, level int) {
	if d.config.EmbedTypeMethods && len(findTypeMethods(t)) > 0 {
		// There are methods on this type, we need to wrap it
		fmt.Fprintln(sb, "{")
		renderVal := fmt.Sprintf("=> %s\n", renderVal)
		d.renderIndent(sb, level+1, renderVal)
		d.renderTypeMethods(sb, t, level+1, 0)
		d.renderIndent(sb, level, "")
		fmt.Fprint(sb, "}")
	} else {
		// Do not wrap, simply print the value
		fmt.Fprint(sb, renderVal)
	}
}

func padRight(s string, unformattedWidth int, maxWidth int) string {
	if unformattedWidth >= maxWidth {
		return s
	}
	return s + strings.Repeat(" ", maxWidth-unformattedWidth)
}
