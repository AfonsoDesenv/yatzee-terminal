package scores

import (
	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/score"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

// Screen é o placar. Esqueleto: preenchido na Task 9.
type Screen struct{}

// New monta a tela com o store de que ela precisa.
func New(store score.Store) Screen { return Screen{} }

func (s Screen) Init() tea.Cmd                       { return nil }
func (s Screen) Update(tea.Msg) (ui.Screen, tea.Cmd) { return s, nil }
func (s Screen) View() string                        { return "" }
func (s Screen) Title() string                       { return "scores" }
