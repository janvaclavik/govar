package govar

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
