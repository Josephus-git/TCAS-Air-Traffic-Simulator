package ui

import (
	"image/color" // Required for defining colors

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme" // Used for theme constants and fallback to default theme
)

// CustomDarkTheme implements fyne.Theme to provide a dark mode appearance.
// It's an empty struct, as its purpose is just to attach methods.
type CustomDarkTheme struct{}

// Color returns a theme-specific color for the given FyneColorName.
func (t CustomDarkTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.RGBA{R: 0x21, G: 0x21, B: 0x21, A: 0xFF} // A deep dark grey for backgrounds
	case theme.ColorNameForeground:
		return color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF} // White for foreground text
	case theme.ColorNamePrimary:
		return color.RGBA{R: 0x64, G: 0xB5, B: 0xF6, A: 0xFF} // A soft blue for primary accents/buttons
	case theme.ColorNameInputBackground:
		return color.RGBA{R: 0x42, G: 0x42, B: 0x42, A: 0xFF} // Slightly lighter dark grey for input fields
	case theme.ColorNamePlaceHolder:
		return color.RGBA{R: 0x9E, G: 0x9E, B: 0x9E, A: 0xFF} // Grey for placeholder text
	case theme.ColorNameScrollBar:
		return color.RGBA{R: 0x61, G: 0x61, B: 0x61, A: 0xFF} // Darker grey for scrollbars
	case theme.ColorNameSelection:
		return color.RGBA{R: 0x42, G: 0x42, B: 0x42, A: 0xFF} // Darker grey for selected items
	case theme.ColorNameDisabled:
		return color.RGBA{R: 0x75, G: 0x75, B: 0x75, A: 0xFF} // Medium grey for disabled elements
	case theme.ColorNameError:
		return color.RGBA{R: 0xF4, G: 0x43, B: 0x36, A: 0xFF} // Red for errors
	case theme.ColorNameFocus:
		return color.RGBA{R: 0x21, G: 0x96, B: 0xF3, A: 0xFF} // Bright blue for focus indicators
	case theme.ColorNameHover:
		return color.RGBA{R: 0x30, G: 0x30, B: 0x30, A: 0xFF} // Slightly lighter dark grey on hover
	case theme.ColorNameButton:
		// You can differentiate button colors based on their variant
		if variant == theme.VariantLight {
			return color.RGBA{R: 0x42, G: 0x42, B: 0x42, A: 0xFF} // Example: A lighter grey for light variant buttons
		}
		return color.RGBA{R: 0x30, G: 0x30, B: 0x30, A: 0xFF} // Default dark button color
	case theme.ColorNameSeparator:
		return color.RGBA{R: 0x42, G: 0x42, B: 0x42, A: 0xFF} // Dark grey for separators
	case theme.ColorNameSuccess:
		return color.RGBA{R: 0x4C, G: 0xAF, B: 0x50, A: 0xFF} // Green for success messages
	case theme.ColorNameWarning:
		return color.RGBA{R: 0xFF, G: 0xC1, B: 0x07, A: 0xFF} // Amber for warnings
	}

	// Fallback: If a color name isn't explicitly defined above, use the default Fyne theme's color.
	return theme.DefaultTheme().Color(name, variant)
}

// Font returns a theme-specific font resource. For simplicity, we'll use Fyne's default font.
// You could load a custom .ttf font here if desired.
func (t CustomDarkTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

// Size returns a theme-specific size. We'll use Fyne's default sizes for consistency.
func (t CustomDarkTheme) Size(name fyne.ThemeSizeName) float32 {
	// The theme.Size method returns a floatModifier, which we convert to float32
	return theme.DefaultTheme().Size(name)
}

// Icon returns a theme-specific icon resource. We'll use Fyne's default icons.
// You could embed custom SVG icons here if needed.
func (t CustomDarkTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}
