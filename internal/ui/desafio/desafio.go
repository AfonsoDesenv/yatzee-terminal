// Package desafio tem a tela que lista os desafios.
package desafio

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/regras"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

// Screen é a lista de desafios. Não guarda cópia nenhuma do catálogo: lê
// regras.Desafios na hora, então um desafio novo aparece aqui sem esta tela
// precisar mudar.
type Screen struct {
	cursor          int
	largura, altura int
	theme           ui.Theme
	keys            ui.Keys
}

// New monta a tela.
func New() Screen {
	return Screen{
		largura: ui.DefaultWidth,
		altura:  ui.DefaultHeight,
		theme:   ui.NewTheme(),
		keys:    ui.DefaultKeys(),
	}
}

// Cursor é o índice do desafio selecionado.
func (s Screen) Cursor() int { return s.cursor }

// Title identifica a tela.
func (s Screen) Title() string { return "desafios" }

// Init não precisa de nada.
func (s Screen) Init() tea.Cmd { return nil }

// Update move o cursor e emite o desafio escolhido.
func (s Screen) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.largura, s.altura = msg.Width, msg.Height
		return s, nil

	case tea.KeyPressMsg:
		pressed := msg.String()
		switch {
		case s.keys.Matches(pressed, s.keys.Up):
			if s.cursor > 0 {
				s.cursor--
			}
		case s.keys.Matches(pressed, s.keys.Down):
			if s.cursor < len(regras.Desafios)-1 {
				s.cursor++
			}
		case s.keys.Matches(pressed, s.keys.Confirm):
			if len(regras.Desafios) == 0 {
				return s, nil
			}
			return s, ui.StartChallenge(s.cursor)
		case s.keys.Matches(pressed, s.keys.Back):
			return s, ui.Back
		}
	}
	return s, nil
}

// View lista os desafios como placas e descreve o selecionado na caixa de
// baixo, do mesmo jeito que a loja: a descrição acompanha o foco em vez de
// todas ficarem na tela de uma vez.
func (s Screen) View() string {
	const recuo = 3 // o espaço onde o ">" do foco entra

	interna := larguraInterna()
	// A faixa que as placas ocupam. O título se centraliza nela, e não na
	// largura da caixa, senão ficava fora de prumo com a lista.
	faixa := recuo + interna + 2

	var bloco []string
	bloco = append(bloco, ui.Centraliza(s.theme.Title.Render("D E S A F I O S"), faixa), "")

	for i, d := range regras.Desafios {
		caixa := ui.Painel(interna, d.Nome)
		marca := strings.Repeat(" ", recuo)
		if i == s.cursor {
			marca = " > "
			caixa = ui.Pinta(s.theme.Selected, ui.Painel(interna, d.Nome))
		}
		for j, l := range caixa {
			prefixo := strings.Repeat(" ", recuo)
			if j == 1 {
				prefixo = marca
			}
			bloco = append(bloco, prefixo+l)
		}
	}

	recuado := strings.Repeat(" ", recuo)
	for _, l := range append([]string{""}, ui.Painel(interna, s.resumo())...) {
		bloco = append(bloco, recuado+l)
	}

	tela := ui.CentralizaBloco(bloco, s.largura, s.altura-1)
	ajuda := ui.Centraliza(s.keys.Help([]string{"setas movem", "espaco comeca", "esc volta"}), s.largura)
	return strings.Join(append(tela, ajuda), "\n")
}

// larguraInterna acompanha o resumo mais longo do catálogo: um desafio novo
// com explicação comprida alarga a caixa, em vez de ter o texto cortado pela
// metade por uma largura fixa escolhida hoje.
func larguraInterna() int {
	maior := 30
	for _, d := range regras.Desafios {
		if l := ui.Largura(d.Resumo); l > maior {
			maior = l
		}
	}
	return maior + 4
}

// resumo é o que o desafio selecionado faz.
func (s Screen) resumo() string {
	if s.cursor < 0 || s.cursor >= len(regras.Desafios) {
		return "nenhum desafio"
	}
	return regras.Desafios[s.cursor].Resumo
}
