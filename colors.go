package govar

// ANSI color codes inspired by Go brand colors
const (
	ColorPaleGray  = "\033[38;5;250m" // #B0BEC5
	ColorSlateGray = "\033[38;5;245m" // #A0A8B3
	ColorDimGray   = "\033[38;5;240m" // #5F6368
	ColorDarkGray  = "\033[38;5;238m" // #444444

	ColorLime         = "\033[38;5;120m" // #A8FF80 → brighter lime
	ColorSkyBlue      = "\033[38;5;123m" // #77DDEE → slightly punchier sky blue
	ColorMutedBlue    = "\033[38;5;111m" // #7DCBEB → brighter muted blue
	ColorLightTeal    = "\033[38;5;80m"  // #5CD5D0 → fresher teal
	ColorGoBlue       = "\033[38;5;39m"  // #00CFFF → boosted Go blue
	ColorDarkTeal     = "\033[38;5;30m"  // #005F5F
	ColorDarkGoBlue   = "\033[38;5;25m"  // #0077AF → slightly brighter
	ColorSeafoamGreen = "\033[38;5;86m"  // #70F0E0 → more luminous seafoam
	ColorGreen        = "\033[38;5;40m"  // #00d75f → fresher, still readable
	ColorGoldenrod    = "\033[38;5;227m" // #FFE082 → brighter golden yellow
	ColorCoralRed     = "\033[38;5;203m" // #F46C5E → lighter and warmer coral
	ColorRed          = "\033[38;5;196m" // #FF0000 → vivid red

	ColorPink = "\033[38;5;219m" // #ffbaff → a brighter pink than 218

	ColorReset = "\033[0m"
)

// ColorPaletteHTML maps color codes to HTML colors.
var ColorPaletteHTML = map[string]string{
	ColorPaleGray:  "#B0BEC5", // #B0BEC5
	ColorSlateGray: "#A0A8B3", // #A0A8B3
	ColorDimGray:   "#5F6368", // #5F6368
	ColorDarkGray:  "#444444", // #444444

	ColorLime:         "#A8FF80", // #A8FF80
	ColorSkyBlue:      "#77DDEE", // #77DDEE
	ColorMutedBlue:    "#7DCBEB", // #7DCBEB
	ColorLightTeal:    "#5CD5D0", // #5CD5D0
	ColorGoBlue:       "#00CFFF", // #00CFFF
	ColorDarkTeal:     "#005F5F", // #005F5F
	ColorDarkGoBlue:   "#0077AF", // #0077AF
	ColorSeafoamGreen: "#70F0E0", // #70F0E0
	ColorGreen:        "#00d75f", // #00d75f
	ColorGoldenrod:    "#FFE082", // #FFE082
	ColorCoralRed:     "#FF857F", // #F46C5E
	ColorRed:          "#FF0000", // #FF0000

	ColorPink: "#ffbaff", // #ffbaff
}
