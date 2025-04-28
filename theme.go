package main

import (
	"fyne.io/fyne/v2"
	"image/color"
)

type ForcedVariant struct {
	fyne.Theme
	fyne.ThemeVariant
}

func (f *ForcedVariant) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	return f.Theme.Color(name, f.ThemeVariant)
}
