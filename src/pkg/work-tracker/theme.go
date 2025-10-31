package worktracker

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

/* ---- Minimal theme scaler ---- */
type scaledTheme struct {
	base   fyne.Theme
	factor float32
}

func (t scaledTheme) Color(n fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	return t.base.Color(n, theme.VariantLight)
}
func (t scaledTheme) Font(st fyne.TextStyle) fyne.Resource {
	return t.base.Font(st)
}
func (t scaledTheme) Icon(n fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(n)
}
func (t scaledTheme) Size(n fyne.ThemeSizeName) float32 {
	// Scale ALL sizes – simplest way to bump label text size everywhere.
	// If you want only text scaled, we can branch on n.
	return t.base.Size(n) * t.factor
}