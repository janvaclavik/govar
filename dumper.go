// Package govar provides a powerful and highly configurable pretty-printer for Go
// data structures. This file contains the core Dumper struct and its primary
// methods for rendering and formatting values. The complex logic for reference
// tracking is located in `references.go`.
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
	"unsafe"
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
	IgnoreStringer      bool   // Ignores fmt.Stringer and error formatting if true
}

// Dumper is a configurable structure-aware pretty printer for Go values.
type Dumper struct {
	config DumperConfig
	Formatter
	// --- Reference Tracking State ---
	referenceStats     map[canonicalKey]*RefStats       // Statistics for each tracked value.
	referenceIDs       map[canonicalKey]string          // Assigned ID (e.g., "&1") for each root value.
	canonicalRoots     map[canonicalKey]canonicalKey    // Union-find structure to group identical values.
	primitiveInstances map[canonicalKey]any             // Stores instances of primitive values for unification.
	definitionPoints   map[canonicalKey]definitionPoint // The chosen definition point for each ID.
	renderedIDs        map[canonicalKey]bool            // Tracks if an ID has already been printed.
	fakeAddrs          map[any]uintptr                  // Assigns synthetic addresses to non-addressable primitives.
	// --- Simple Cycle Detection State ---
	visitedPointers map[unsafe.Pointer]bool // Used for basic cycle detection when TrackReferences is off.
}

// NewDumper creates a new Dumper with the provided configuration.
func NewDumper(cfg DumperConfig) *Dumper {
	return &Dumper{
		config:             cfg,
		Formatter:          &PlainFormatter{},
		referenceStats:     make(map[canonicalKey]*RefStats),
		referenceIDs:       make(map[canonicalKey]string),
		canonicalRoots:     make(map[canonicalKey]canonicalKey),
		primitiveInstances: make(map[canonicalKey]any),
		definitionPoints:   make(map[canonicalKey]definitionPoint),
		renderedIDs:        make(map[canonicalKey]bool),
		fakeAddrs:          make(map[any]uintptr),
		visitedPointers:    make(map[unsafe.Pointer]bool),
	}
}

// Die dumps the given values and immediately terminates the program.
func (d *Dumper) Die(vs ...any) {
	d.Dump(vs...)
	os.Exit(1)
}

// Dump prints values to stdout using the configured formatting.
func (d *Dumper) Dump(vs ...any) {
	if d.config.UseColors {
		d.Formatter = &ANSIcolorFormatter{}
	} else {
		d.Formatter = &PlainFormatter{}
	}
	sb := &strings.Builder{}
	d.renderHeader(sb)
	d.renderAllValues(sb, vs...)
	fmt.Fprintln(os.Stdout, sb.String())
}

// Fdump writes values to the given io.Writer using the configured formatting.
func (d *Dumper) Fdump(w io.Writer, vs ...any) {
	if d.config.UseColors {
		d.Formatter = &ANSIcolorFormatter{}
	} else {
		d.Formatter = &PlainFormatter{}
	}
	sb := &strings.Builder{}
	d.renderHeader(sb)
	d.renderAllValues(sb, vs...)
	fmt.Fprintln(w, sb.String())
}

// Sdump returns a string containing the formatted values.
func (d *Dumper) Sdump(vs ...any) string {
	if d.config.UseColors {
		d.Formatter = &ANSIcolorFormatter{}
	} else {
		d.Formatter = &PlainFormatter{}
	}
	sb := &strings.Builder{}
	d.renderHeader(sb)
	d.renderAllValues(sb, vs...)
	return sb.String()
}

// SdumpHTML returns an HTML-formatted dump wrapped in a <pre> block.
func (d *Dumper) SdumpHTML(vs ...any) string {
	d.Formatter = &HTMLformatter{HTMLtagToken: d.config.HTMLtagToken, UseColors: d.config.UseColors}

	sb := &strings.Builder{}
	sb.WriteString(fmt.Sprintf(`<%s class="govar" style="background-color:black; color:white; padding:4px; border-radius: 4px">`+"\n", d.config.HTMLtagSection))
	d.renderHeader(sb)
	d.renderAllValues(sb, vs...)
	sb.WriteString(fmt.Sprintf("</%s>", d.config.HTMLtagSection))
	return sb.String()
}

// asStringerInterface checks if a value implements the fmt.Stringer interface.
// If so, it returns the formatted string from its String() method.
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

// asErrorInterface checks if a value implements the error interface. If so, it
// returns the formatted error string. Otherwise, it returns an empty string.
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

// calculateStructPadding determines the maximum key and type string lengths for
// fields within a struct to align them neatly in block mode.
func (d *Dumper) calculateStructPadding(v reflect.Value) (int, int) {
	maxKeyLen, maxTypeLen := 0, 0
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field, fieldVal := t.Field(i), v.Field(i)
		if field.PkgPath != "" {
			fieldVal = tryExport(fieldVal)
		}
		keyLen := utf8.RuneCountInString(field.Name)
		if keyLen > maxKeyLen {
			maxKeyLen = keyLen
		}
		typeName := d.formatTypeNoColors(fieldVal, false)
		if utf8.RuneCountInString(typeName) > maxTypeLen {
			maxTypeLen = utf8.RuneCountInString(typeName)
		}
	}
	if d.config.EmbedTypeMethods {
		for _, m := range findTypeMethods(t) {
			if utf8.RuneCountInString(m.Name) > maxKeyLen {
				maxKeyLen = utf8.RuneCountInString(m.Name)
			}
		}
	}
	maxKeyLen += 2 // for visibility symbol
	return maxKeyLen, maxTypeLen
}

// estimatedInlineLength calculates the approximate string length of a value if it
// were to be rendered on a single line. This is used as a heuristic to decide
// whether to use inline or block formatting.
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

// formatArrayOrSlice formats a slice or an array, deciding between inline and block rendering.
func (d *Dumper) formatArrayOrSlice(v reflect.Value, level int) string {
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
				fmt.Fprint(sb, d.ApplyFormat(ColorSlateGray, "… (truncated)"))
				break
			}
			if i > 0 {
				fmt.Fprint(sb, ", ")
			}
			formattedType := d.formatType(v.Index(i), true)
			indexSymbol := d.ApplyFormat(ColorDarkTeal, fmt.Sprintf("%d", i))

			fmt.Fprintf(sb, "%s%s => ", indexSymbol, formattedType)
			d.renderValue(sb, v.Index(i), level, false)
		}

	} else {
		// BLOCK RENDER
		fmt.Fprintln(sb)
		if d.config.ShowHexdump && v.Type().Elem().Kind() == reflect.Uint8 {
			d.renderHexdump(sb, v, level)
		} else {
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
				formattedType := d.formatType(v.Index(i), true)
				indexSymbol := d.ApplyFormat(ColorDarkTeal, fmt.Sprintf("%d", i))

				renderIndex := ""
				if formattedType != "" {
					unformattedTypeLen := utf8.RuneCountInString(d.formatTypeNoColors(v.Index(i), true))
					paddedType := padRight(formattedType, unformattedTypeLen, maxTypeLen)
					renderIndex = fmt.Sprintf("%s %s => ", indexSymbol, paddedType)
				} else {
					renderIndex = fmt.Sprintf("%s => ", indexSymbol)
				}
				d.renderIndent(sb, level+1, renderIndex)
				d.renderValue(sb, v.Index(i), level+1, false)
				fmt.Fprintln(sb)
			}
		}
		d.renderIndent(sb, level, "")
	}

	fmt.Fprint(sb, "]")

	return sb.String()
}

// formatBool formats a boolean value with color.
func (d *Dumper) formatBool(v reflect.Value) string {
	if v.Bool() {
		return d.ApplyFormat(ColorGreen, "true")
	} else {
		return d.ApplyFormat(ColorCoralRed, "false")
	}
}

// formatChan formats a channel, showing its direction, buffer capacity, and pointer address.
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

// formatFunc formats a function, showing its name and pointer address.
func (d *Dumper) formatFunc(v reflect.Value) string {
	funName := d.ApplyFormat(ColorLightTeal, getFunctionName(v))
	if d.config.ShowMetaInformation {
		funName = fmt.Sprint(d.metaHint(fmt.Sprintf("func@%#x", v.Pointer()), "")) + funName
	}
	return funName
}

// formatMap formats a map, deciding between inline and block rendering.
func (d *Dumper) formatMap(v reflect.Value, level int) string {
	sb := &strings.Builder{}

	if d.config.ShowMetaInformation {
		mapLen := fmt.Sprintf("%d", v.Len())
		fmt.Fprint(sb, d.metaHint(mapLen, ""))
	}

	sortedKeys := sortMapKeys(v)
	fmt.Fprint(sb, "[")

	if d.shouldRenderInline(v) {
		// INLINE RENDER
		for i, key := range sortedKeys {
			if i >= d.config.MaxItems {
				fmt.Fprint(sb, d.ApplyFormat(ColorSlateGray, "… (truncated)"))
				break
			}
			if i > 0 {
				fmt.Fprint(sb, ", ")
			}
			keyStr := d.formatMapKeyAsIndex(key)
			formattedType := d.formatType(v.MapIndex(key), true)
			fmt.Fprintf(sb, "%s %s => ", d.ApplyFormat(ColorDarkTeal, keyStr), formattedType)
			d.renderValue(sb, v.MapIndex(key), level, false)
		}
	} else {
		// BLOCK RENDER
		fmt.Fprintln(sb)
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
				d.renderIndent(sb, level+1, d.ApplyFormat(ColorSlateGray, "… (truncated)\n"))
				break
			}
			keyStr := d.formatMapKeyAsIndex(key)
			formattedType := d.formatType(v.MapIndex(key), true)
			keyRender := ""
			if formattedType != "" {
				unformattedTypeLen := utf8.RuneCountInString(d.formatTypeNoColors(v.MapIndex(key), true))
				paddedKey := padRight(keyStr, utf8.RuneCountInString(keyStr), maxKeyLen)
				paddedType := padRight(formattedType, unformattedTypeLen, maxTypeLen)
				keyRender = fmt.Sprintf("%s  %s => ", d.ApplyFormat(ColorDarkTeal, paddedKey), paddedType)
			} else {
				keyRender = fmt.Sprintf("%s => ", keyStr)
			}
			d.renderIndent(sb, level+1, keyRender)
			d.renderValue(sb, v.MapIndex(key), level+1, false)
			fmt.Fprintln(sb)
		}
		d.renderIndent(sb, level, "")
	}

	fmt.Fprint(sb, "]")
	return sb.String()
}

// formatMapKeyAsIndex formats a map key for display. Simple keys are formatted
// directly, while complex keys are summarized.
func (d *Dumper) formatMapKeyAsIndex(k reflect.Value) string {
	var keyFormatted string
	if d.isSimpleMapKey(k) {
		switch k.Kind() {
		case reflect.String:
			keyFormatted = strconv.Quote(k.Interface().(string))
		case reflect.Interface:
			if k.Type().NumMethod() == 0 {
				if str, ok := k.Interface().(string); ok {
					keyFormatted = strconv.Quote(str)
				} else {
					keyFormatted = fmt.Sprintf("%v", k.Interface())
				}
			} else {
				keyFormatted = fmt.Sprintf("%v", k.Interface())
			}
		default:
			keyFormatted = fmt.Sprintf("%v", k.Interface())
		}
	} else {
		keyFormatted = fmt.Sprintf("%v", k.Interface())
	}

	return keyFormatted
}

// formatString formats a string, escaping special characters, applying color,
// and adding metadata like rune count.
func (d *Dumper) formatString(v reflect.Value) string {
	strLen := utf8.RuneCountInString(v.String())
	str := d.stringEscape(v.String())
	str = d.ApplyFormat(ColorGoldenrod, `"`) + d.ApplyFormat(ColorLime, str) + d.ApplyFormat(ColorGoldenrod, `"`)
	if d.config.ShowMetaInformation {
		str = d.metaHint(fmt.Sprintf("R:%d", strLen), "") + str
	}
	return str
}

// formatType formats the type of a value as a colored string.
func (d *Dumper) formatType(v reflect.Value, isInCollection bool) string {
	if !d.config.ShowTypes {
		return ""
	}
	return d.ApplyFormat(ColorDarkGray, d.formatTypeNoColors(v, isInCollection))
}

// formatTypeNoColors formats the type of a value as a plain string, without colors.
// This is used for padding calculations.
func (d *Dumper) formatTypeNoColors(v reflect.Value, isInCollection bool) string {
	if !d.config.ShowTypes {
		return ""
	}
	if !v.IsValid() {
		return "invalid"
	}
	vKind := v.Kind()
	expectedType := ""
	if vKind == reflect.Interface {
		expectedType = "⧉ " + v.Type().String()
	} else if vKind == reflect.Array || vKind == reflect.Slice || vKind == reflect.Map || vKind == reflect.Struct {
		expectedType = v.Type().String()
	} else if !isInCollection {
		expectedType = v.Type().String()
	}
	actualType := ""
	if vKind == reflect.Interface && !v.IsNil() {
		actualType = "(" + v.Elem().Type().String() + ")"
	}
	formattedType := expectedType + actualType
	formattedType = strings.ReplaceAll(formattedType, "interface {}", "any")
	return formattedType
}

// isSimpleMapKey checks if a map key is a simple primitive that can be rendered inline easily.
func (d *Dumper) isSimpleMapKey(k reflect.Value) bool {
	if isSimpleValue(k) || k.Kind() == reflect.Complex64 || k.Kind() == reflect.Complex128 {
		return true
	} else {
		return d.estimatedInlineLength(k) <= d.config.MaxInlineLength
	}
}

// isSimpleStruct checks if a struct contains only simple primitive fields and has no methods.
func (d *Dumper) isSimpleStruct(v reflect.Value) bool {
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

// metaHint formats a metadata hint (e.g., "|L:5 C:10|") with color.
func (d *Dumper) metaHint(msg string, ico string) string {
	if ico != "" {
		return d.ApplyFormat(ColorDimGray, fmt.Sprintf("|%s %s| ", ico, msg))
	}
	return d.ApplyFormat(ColorDimGray, fmt.Sprintf("|%s| ", msg))
}

// renderAllValues orchestrates the analysis and rendering of all provided values.
func (d *Dumper) renderAllValues(sb *strings.Builder, vs ...any) {
	if len(vs) == 0 {
		return
	}
	addressableVars := make([]reflect.Value, len(vs))
	for i, v := range vs {
		addressableVars[i] = makeAddressable(reflect.ValueOf(v))
	}

	// The analysis pipeline for ID/back-reference tracking.
	if d.config.TrackReferences {
		d.resetState()
		// 1. Traverse the object graph to collect stats on all values.
		for _, v := range addressableVars {
			d.preScanBFS(v)
		}
		// 2. Unify identical values (copies) with their original sources.
		d.unifyAllCopies()
		// 3. Assign IDs (e.g., "&1") to values that are referenced multiple times.
		d.assignReferenceIDs()
		// 4. Determine the best location to print each ID.
		for _, v := range addressableVars {
			d.determineDefinitionPoints(v)
		}
	}

	// Render each top-level value.
	for i, v := range addressableVars {
		if i > 0 {
			sb.WriteString("\n")
		}
		vType, tmpRv := checkNilInterface(vs[i])
		if d.config.ShowTypes {
			if vType != "unknown" {
				vType = d.formatType(v, false)
			} else {
				vType = d.ApplyFormat(ColorDarkGray, vType)
			}
			fmt.Fprint(sb, vType, " => ")
		}

		if tmpRv != "" {
			sb.WriteString(d.ApplyFormat(ColorCoralRed, tmpRv))
		} else {
			d.renderValue(sb, v, 0, false)
		}
		fmt.Fprintln(sb)
	}
}

// renderBackref writes a back-reference symbol "↩︎ &N" to the string builder.
func (d *Dumper) renderBackref(sb *strings.Builder, id string) {
	fmt.Fprint(sb, d.ApplyFormat(ColorPink, "↩︎ "+id))
}

// renderHeader prints the file and line number of the Dump() call.
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

// renderHexdump formats a byte slice as a classic hexdump.
func (d *Dumper) renderHexdump(sb *strings.Builder, v reflect.Value, level int) {
	content := toAddressableByteSlice(v)
	lines := strings.Split(hex.Dump(content), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		if len(line) < 10 {
			fmt.Println(line) // fallback
			continue
		}
		offsetPart := line[:10]
		hexPart := line[10:58]
		asciiPart := ""
		if idx := strings.Index(line, "  |"); idx != -1 {
			asciiPart = line[idx:]
		}
		d.renderIndent(sb, level+1, "")
		fmt.Fprintf(sb, "%s%s%s\n",
			d.ApplyFormat(ColorDarkTeal, offsetPart),
			d.ApplyFormat(ColorSkyBlue, hexPart),
			d.ApplyFormat(ColorLime, asciiPart),
		)
	}
}

// renderID writes an ID symbol "&N" to the string builder.
func (d *Dumper) renderID(sb *strings.Builder, id string) {
	fmt.Fprint(sb, d.ApplyFormat(ColorGoldenrod, id+" "))
}

// renderIndent writes indentation spaces to the string builder.
func (d *Dumper) renderIndent(sb *strings.Builder, indentLevel int, text string) {
	fmt.Fprint(sb, strings.Repeat(" ", indentLevel*d.config.IndentWidth)+text)
}

// renderTypeMethods formats and prints all the embedded methods of a given type.
func (d *Dumper) renderTypeMethods(sb *strings.Builder, t reflect.Type, level int, maxNameLen int) {
	for _, m := range findTypeMethods(t) {
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

// renderPrimitive formats a basic Go type (int, string, bool, etc.) into a string.
func (d *Dumper) renderPrimitive(v reflect.Value) string {
	switch v.Kind() {
	case reflect.Bool:
		return d.formatBool(v)
	case reflect.String:
		return d.formatString(v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return d.ApplyFormat(ColorSkyBlue, fmt.Sprint(v.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return d.ApplyFormat(ColorSkyBlue, fmt.Sprint(v.Uint()))
	case reflect.Float32, reflect.Float64:
		return d.ApplyFormat(ColorSkyBlue, fmt.Sprintf("%f", v.Float()))
	case reflect.Complex64, reflect.Complex128:
		return d.ApplyFormat(ColorSkyBlue, fmt.Sprintf("%v", v.Complex()))
	}
	return "" // Should not be reached
}

// renderStruct formats a struct, deciding between inline and block rendering.
func (d *Dumper) renderStruct(sb *strings.Builder, v reflect.Value, level int) {
	t := v.Type()
	fmt.Fprint(sb, "{")

	if d.shouldRenderInline(v) {
		// --- INLINE RENDER ---
		for i := 0; i < t.NumField(); i++ {
			if i > 0 {
				fmt.Fprint(sb, ", ")
			}
			field, fieldVal := t.Field(i), v.Field(i)
			// Special check for embedded structs that are back-references.
			if d.config.TrackReferences && fieldVal.Kind() == reflect.Struct {
				rawKey, ok := d.getRawKey(fieldVal)
				if ok {
					rootKey := d.findRoot(rawKey)
					if id, hasID := d.referenceIDs[rootKey]; hasID {
						def, defExists := d.definitionPoints[rootKey]
						if defExists && def.isPointerRef && deref(fieldVal).Type() == def.valueType {
							d.renderStructField(sb, field, fieldVal, 0, 0, true)
							d.renderBackref(sb, id)
							continue
						}
					}
				}
			}
			d.renderStructField(sb, field, fieldVal, 0, 0, true)
			d.renderValue(sb, fieldVal, level, false)
		}
	} else {
		// --- BLOCK RENDER ---
		fmt.Fprintln(sb)
		maxKeyLen, maxTypeLen := d.calculateStructPadding(v)

		for i := 0; i < t.NumField(); i++ {
			field, fieldVal := t.Field(i), v.Field(i)

			// Special check for embedded structs that are back-references.
			if d.config.TrackReferences && fieldVal.Kind() == reflect.Struct {
				rawKey, ok := d.getRawKey(fieldVal)
				if ok {
					rootKey := d.findRoot(rawKey)
					if id, hasID := d.referenceIDs[rootKey]; hasID {
						def, defExists := d.definitionPoints[rootKey]
						if defExists && def.isPointerRef && deref(fieldVal).Type() == def.valueType {
							d.renderIndent(sb, level+1, "")
							d.renderStructField(sb, field, fieldVal, maxKeyLen, maxTypeLen, false)
							d.renderBackref(sb, id)
							fmt.Fprintln(sb)
							continue
						}
					}
				}
			}
			d.renderIndent(sb, level+1, "")
			d.renderStructField(sb, field, fieldVal, maxKeyLen, maxTypeLen, false)
			d.renderValue(sb, fieldVal, level+1, false)
			fmt.Fprintln(sb)
		}
		if d.config.EmbedTypeMethods {
			d.renderTypeMethods(sb, t, level+1, maxKeyLen)
		}
		d.renderIndent(sb, level, "")
	}
	fmt.Fprint(sb, "}")
}

// renderStructField is a helper to format the field part of a struct line.
func (d *Dumper) renderStructField(sb *strings.Builder, field reflect.StructField, fieldVal reflect.Value, maxKeyLen, maxTypeLen int, isInline bool) {
	renderVal := fieldVal
	symbol := "⯀ "
	if !field.IsExported() {
		symbol = "🞏 "
		renderVal = tryExport(fieldVal)
	}

	unformattedFieldLen := utf8.RuneCountInString(symbol + field.Name)
	unformattedTypeLen := utf8.RuneCountInString(d.formatTypeNoColors(renderVal, false))
	symbol = d.ApplyFormat(ColorDarkGoBlue, symbol)
	fieldName := d.ApplyFormat(ColorLightTeal, field.Name)
	formattedType := d.formatType(renderVal, false)

	var fieldRender string
	if isInline {
		// INLINE RENDER
		fieldRender = fmt.Sprintf("%s%s", symbol, fieldName)
		if formattedType != "" {
			fieldRender += " " + formattedType
		}
		fieldRender += " => "
	} else {
		// BLOCK RENDER
		fieldRender = padRight(symbol+fieldName, unformattedFieldLen, maxKeyLen)
		if formattedType != "" {
			fieldRender += "  " + padRight(formattedType, unformattedTypeLen, maxTypeLen)
		}
		fieldRender += " => "
	}

	sb.WriteString(fieldRender)
}

// renderValue is the main recursive rendering function. It handles printing a single value,
// including its ID/back-reference if applicable, and then delegates to type-specific formatters.
func (d *Dumper) renderValue(sb *strings.Builder, v reflect.Value, level int, skipRefCheck bool) {
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

	// Handle ID and back-reference printing.
	if d.config.TrackReferences && !skipRefCheck {
		rawKey, keyOK := d.getRawKey(v)
		if keyOK {
			rootKey := d.findRoot(rawKey)
			if id, hasID := d.referenceIDs[rootKey]; hasID {
				def, defExists := d.definitionPoints[rootKey]
				instKey, instKeyOK := d.getInstanceKey(v)
				// Check if the current value is the chosen "definition point".
				isTheChosenDefinition := defExists && instKeyOK && def.instanceKey == instKey

				if isTheChosenDefinition {
					// This is the definition point. If we've already rendered it (e.g., a cycle),
					// print a back-reference. Otherwise, print the ID and render the value.
					if d.renderedIDs[rootKey] {
						d.renderBackref(sb, id)
						return
					}
					d.renderID(sb, id)
					d.renderedIDs[rootKey] = true
				} else {
					// This is not the definition point, so it must be a back-reference.
					d.renderBackref(sb, id)
					return
				}
			}
		}
	}

	// Simple cycle detection for when TrackReferences is false.
	if !d.config.TrackReferences {
		addr := getValPtr(v)
		if addr != nil {
			if d.visitedPointers[addr] {
				sb.WriteString(d.ApplyFormat(ColorSlateGray, "<cycle>"))
				return
			}
			d.visitedPointers[addr] = true
		}
	}

	// Check for fmt.Stringer or error interfaces.
	exportedV := tryExport(v)
	if exportedV.Kind() != reflect.Interface && !d.config.IgnoreStringer {
		if str := d.asStringerInterface(exportedV); str != "" {
			if d.config.ShowMetaInformation {
				fmt.Fprint(sb, d.metaHint("as Stringer", ""))
			}
			fmt.Fprint(sb, str)
			return
		}
		if str := d.asErrorInterface(exportedV); str != "" {
			if d.config.ShowMetaInformation {
				fmt.Fprint(sb, d.metaHint("as error", ""))
			}
			fmt.Fprint(sb, str)
			return
		}
	}

	// Delegate to kind-specific rendering functions.
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		d.renderValue(sb, v.Elem(), level, true) // Dereference and render, skipping the next ref check.
	case reflect.Struct:
		d.renderStruct(sb, v, level)
	case reflect.Slice, reflect.Array:
		renderVal := d.formatArrayOrSlice(v, level)
		d.wrapAndRender(sb, renderVal, v.Type(), level)
	case reflect.Map:
		renderVal := d.formatMap(v, level)
		d.wrapAndRender(sb, renderVal, v.Type(), level)
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.String:
		renderVal := d.renderPrimitive(v)
		d.wrapAndRender(sb, renderVal, v.Type(), level)
	case reflect.UnsafePointer:
		fmt.Fprint(sb, d.ApplyFormat(ColorSlateGray, fmt.Sprintf("unsafe.Pointer(%#x)", v.Pointer())))
	case reflect.Func:
		renderVal := d.formatFunc(v)
		d.wrapAndRender(sb, renderVal, v.Type(), level)
	case reflect.Chan:
		renderVal := d.formatChan(v)
		d.wrapAndRender(sb, renderVal, v.Type(), level)
	default:
		fmt.Fprintln(sb, "[WARNING] unknown reflect.Kind")
	}
}

// shouldRenderInline determines if a value is simple enough to be rendered on a
// single line. The decision is based on its kind, number of elements, and
// estimated inline length.
func (d *Dumper) shouldRenderInline(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		return isSimpleCollection(v) && v.Len() <= 10 && d.estimatedInlineLength(v) <= d.config.MaxInlineLength
	case reflect.Map:
		return isSimpleMap(v) && v.Len() <= 10 && d.estimatedInlineLength(v) <= d.config.MaxInlineLength
	case reflect.Struct:
		if d.config.EmbedTypeMethods && len(findTypeMethods(v.Type())) > 0 {
			return false
		}
		return d.isSimpleStruct(v) && v.NumField() <= 10 && d.estimatedInlineLength(v) <= d.config.MaxInlineLength
	default:
		return true
	}
}

// stringEscape truncates a string if it exceeds MaxStringLen and escapes
// common non-printable characters.
func (d *Dumper) stringEscape(str string) string {
	if utf8.RuneCountInString(str) > d.config.MaxStringLen {
		runes := []rune(str)
		str = string(runes[:d.config.MaxStringLen]) + "…"
	}
	replacer := strings.NewReplacer("\n", `\n`, "\t", `\t`, "\r", `\r`, "\v", `\v`, "\f", `\f`, "\x1b", `\x1b`)
	return replacer.Replace(str)
}

// wrapAndRender prints the rendered value, wrapping it in braces and showing
// its methods if it's a named type with methods.
func (d *Dumper) wrapAndRender(sb *strings.Builder, renderVal string, t reflect.Type, level int) {
	if d.config.EmbedTypeMethods && len(findTypeMethods(t)) > 0 {
		fmt.Fprintln(sb, "{")
		d.renderIndent(sb, level+1, "=> ")
		sb.WriteString(renderVal)
		fmt.Fprintln(sb)
		d.renderTypeMethods(sb, t, level+1, 0)
		d.renderIndent(sb, level, "")
		fmt.Fprint(sb, "}")
	} else {
		fmt.Fprint(sb, renderVal)
	}
}

// padRight adds spaces to the right of a string to reach a minimum width.
// It correctly handles ANSI color codes, using the unformattedWidth for calculation.
func padRight(s string, unformattedWidth int, maxWidth int) string {
	if unformattedWidth >= maxWidth {
		return s
	}
	return s + strings.Repeat(" ", maxWidth-unformattedWidth)
}
