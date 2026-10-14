package menu

import (
	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

// Screen é o menu. Esqueleto: preenchido na Task 8.
type Screen struct{}

// New monta o menu sabendo quem joga.
func New(playerName string) Screen { return Screen{} }

func (s Screen) Init() tea.Cmd                       { return nil }
func (s Screen) Update(tea.Msg) (ui.Screen, tea.Cmd) { return s, nil }
func (s Screen) View() string                        { return "" }
func (s Screen) Title() string                       { return "menu" }
