package ui

import "charm.land/lipgloss/v2"

// Theme concentra todo estilo num lugar, para a aparência mudar sem caçar
// cores espalhadas pelas telas. Sem Renderer: a lipgloss v2 não tem esse
// conceito, e o perfil de cor do cliente é resolvido pelo programa.
type Theme struct {
	Title    lipgloss.Style
	Item     lipgloss.Style
	Selected lipgloss.Style
	Dim      lipgloss.Style
	Error    lipgloss.Style
	Panel    lipgloss.Style
}

// NewTheme devolve o tema padrão.
func NewTheme() Theme {
	return Theme{
		Title:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#b58cff")),
		Item:     lipgloss.NewStyle().Foreground(lipgloss.Color("#d8d8d8")),
		Selected: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ffd866")),
		Dim:      lipgloss.NewStyle().Foreground(lipgloss.Color("#8a8a8a")),
		Error:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ff6b6b")),
		Panel:    lipgloss.NewStyle().Padding(1, 2),
	}
}
