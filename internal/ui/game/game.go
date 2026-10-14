// Package game tem a tela de partida, dona das duas fases.
package game

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

// Phase é a fase da partida.
type Phase int

const (
	// PhaseRound é a fase de jogar os dados.
	PhaseRound Phase = iota
	// PhaseShop é a loja entre rodadas.
	PhaseShop
)

// Screen é a partida. Rodada e loja não são telas independentes na pilha: são
// fases, e esta tela é dona das duas, o que mantém o estado da partida num
// lugar só quando as mecânicas chegarem.
type Screen struct {
	phase Phase
	round roundView
	shop  shopView
	theme ui.Theme
	keys  ui.Keys
}

// New começa uma partida na fase de rodada.
func New() Screen {
	theme := ui.NewTheme()
	return Screen{
		phase: PhaseRound,
		round: roundView{theme: theme},
		shop:  shopView{theme: theme},
		theme: theme,
		keys:  ui.DefaultKeys(),
	}
}

// Phase é a fase atual.
func (s Screen) Phase() Phase { return s.phase }

// Title identifica a tela.
func (s Screen) Title() string { return "jogo" }

// Init não precisa de nada ainda.
func (s Screen) Init() tea.Cmd { return nil }

// Update alterna as fases e trata a volta.
func (s Screen) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return s, nil
	}
	pressed := key.String()

	switch {
	case s.keys.Matches(pressed, s.keys.Toggle):
		if s.phase == PhaseRound {
			s.phase = PhaseShop
		} else {
			s.phase = PhaseRound
		}
	case s.keys.Matches(pressed, s.keys.Back):
		return s, ui.Back
	}
	return s, nil
}

// View desenha a fase atual.
func (s Screen) View() string {
	b := &strings.Builder{}
	if s.phase == PhaseRound {
		b.WriteString(s.round.render())
	} else {
		b.WriteString(s.shop.render())
	}
	b.WriteString("\n")
	b.WriteString(s.keys.Help([]string{"tab alterna rodada e loja", "esc volta ao menu"}))
	return s.theme.Panel.Render(b.String())
}
