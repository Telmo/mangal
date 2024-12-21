package style

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/metafates/mangal/color"
)

var (
	Title      = NewColored(color.New("230"), color.New("62")).Padding(0, 1).Render
	ErrorTitle = NewColored(color.New("230"), color.Red).Padding(0, 1).Render
)

// Tag returns a function that colors text with the given foreground and background colors
func Tag(foreground, background lipgloss.Color) func(string) string {
	style := lipgloss.NewStyle().
		Foreground(foreground).
		Background(background).
		Padding(0, 1)
	return func(s string) string {
		return style.Render(s)
	}
}

// Padded returns a function that colors text with the given foreground and background colors and adds padding
func Padded(foreground, background lipgloss.Color) func(string) string {
	style := lipgloss.NewStyle().
		Foreground(foreground).
		Background(background).
		Padding(0, 1)
	return func(s string) string {
		return style.Render(s)
	}
}
