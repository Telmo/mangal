package style

import "github.com/charmbracelet/lipgloss"

// New returns a new lipgloss style
func New() lipgloss.Style {
	return lipgloss.NewStyle()
}

// NewColored returns a new lipgloss style with foreground and background colors
func NewColored(foreground, background lipgloss.Color) lipgloss.Style {
	return New().Foreground(foreground).Background(background)
}

// Fg returns a function that colors text with the given foreground color
func Fg(color lipgloss.Color) func(string) string {
	style := lipgloss.NewStyle().Foreground(color)
	return func(s string) string {
		return style.Render(s)
	}
}

// Bg returns a function that colors text with the given background color
func Bg(color lipgloss.Color) func(string) string {
	style := lipgloss.NewStyle().Background(color)
	return func(s string) string {
		return style.Render(s)
	}
}

// Width returns a function that truncates text to the given width
func Width(max int) func(string) string {
	style := lipgloss.NewStyle().Width(max)
	return func(s string) string {
		return style.Render(s)
	}
}

// Truncate returns a function that truncates text to the given width
func Truncate(max int) func(string) string {
	return func(s string) string {
		if max <= 0 {
			return s
		}
		if len(s) <= max {
			return s
		}
		return s[:max-3] + "..."
	}
}

// Faint returns faint text
func Faint(s string) string {
	return lipgloss.NewStyle().Faint(true).Render(s)
}

var (
	Bold      = lipgloss.NewStyle().Bold(true).Render
	Italic    = lipgloss.NewStyle().Italic(true).Render
	Underline = lipgloss.NewStyle().Underline(true).Render
)
