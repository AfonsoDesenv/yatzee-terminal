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
	phase           Phase
	round           rodada
	shop            loja
	largura, altura int
	theme           ui.Theme
	keys            ui.Keys
}

// New começa uma partida na fase de rodada.
func New() Screen {
	theme := ui.NewTheme()
	keys := ui.DefaultKeys()
	return Screen{
		phase:   PhaseRound,
		round:   novaRodada(theme, keys),
		shop:    novaLoja(theme),
		largura: ui.DefaultWidth,
		altura:  ui.DefaultHeight,
		theme:   theme,
		keys:    keys,
	}
}

// Phase é a fase atual.
func (s Screen) Phase() Phase { return s.phase }

// Title identifica a tela.
func (s Screen) Title() string { return "jogo" }

// Init não precisa de nada ainda.
func (s Screen) Init() tea.Cmd { return nil }

// Update alterna as fases, marca dados e trata a volta.
func (s Screen) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.largura, s.altura = msg.Width, msg.Height
		return s, nil

	case tea.KeyPressMsg:
		pressed := msg.String()
		switch {
		case s.keys.Matches(pressed, s.keys.Toggle):
			if s.phase == PhaseRound {
				s.phase = PhaseShop
			} else {
				s.phase = PhaseRound
			}
		case s.keys.Matches(pressed, s.keys.Back):
			return s, ui.Back
		default:
			// Dígito marca o dado para rerolar: a seleção é por número, não
			// por cursor, então dá para escolher 1, 3 e 5 em três toques.
			if s.phase == PhaseRound && len(pressed) == 1 && pressed[0] >= '1' && pressed[0] <= '9' {
				s.round = s.round.alterna(int(pressed[0] - '0'))
			}
		}
	}
	return s, nil
}

// View desenha a fase atual com a barra de ajuda no rodapé.
func (s Screen) View() string {
	var corpo string
	var dicas []string
	if s.phase == PhaseRound {
		corpo = s.round.render(s.largura, s.altura-1)
		dicas = s.round.ajuda()
	} else {
		corpo = s.shop.render(s.largura, s.altura-1)
		dicas = s.shop.ajuda()
	}
	ajuda := ui.Centraliza(s.keys.Help(dicas), s.largura)
	return strings.Join([]string{corpo, ajuda}, "\n")
}
