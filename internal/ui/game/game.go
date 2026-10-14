package game

import (
	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

// Screen é a partida. Esqueleto: preenchido na Task 10.
type Screen struct{}

// New começa uma partida.
func New() Screen { return Screen{} }

func (s Screen) Init() tea.Cmd                       { return nil }
func (s Screen) Update(tea.Msg) (ui.Screen, tea.Cmd) { return s, nil }
func (s Screen) View() string                        { return "" }
func (s Screen) Title() string                       { return "jogo" }
