// Package menu tem a tela inicial do jogo.
package menu

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

type item struct {
	label  string
	intent tea.Cmd
}

// Screen é o menu principal.
type Screen struct {
	items           []item
	cursor          int
	largura, altura int
	theme           ui.Theme
	keys            ui.Keys
}

// New monta o menu. A tabela carrega a intenção de cada item, então Update não
// tem switch por índice e acrescentar um item é acrescentar uma linha.
func New() Screen {
	return Screen{
		items: []item{
			{label: "NOVO JOGO", intent: ui.StartGame},
			{label: "DESAFIO", intent: ui.OpenChallenges},
			{label: "SAIR", intent: ui.Quit},
		},
		largura: ui.DefaultWidth,
		altura:  ui.DefaultHeight,
		theme:   ui.NewTheme(),
		keys:    ui.DefaultKeys(),
	}
}

// Cursor é o índice do item selecionado.
func (s Screen) Cursor() int { return s.cursor }

// Title identifica a tela.
func (s Screen) Title() string { return "menu" }

// Init não precisa de nada.
func (s Screen) Init() tea.Cmd { return nil }

// Update move o cursor, guarda o tamanho da tela e emite a intenção escolhida.
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
			if s.cursor < len(s.items)-1 {
				s.cursor++
			}
		case s.keys.Matches(pressed, s.keys.Confirm):
			return s, s.items[s.cursor].intent
		case s.keys.Matches(pressed, s.keys.Quit):
			return s, ui.Quit
		}
	}
	return s, nil
}

// View monta título, itens e arte, centralizados na tela.
func (s Screen) View() string {
	// Todos os itens com a mesma largura: alinhados como placas, em vez de
	// cada caixa acompanhar o tamanho do próprio texto.
	maior := 0
	for _, it := range s.items {
		if l := ui.Largura(it.label); l > maior {
			maior = l
		}
	}
	interna := maior + 6

	var bloco []string
	bloco = append(bloco, ui.Centraliza(s.theme.Title.Render("Y A T Z E E"), interna+4), "")
	for i, it := range s.items {
		caixa := ui.Painel(interna, it.label)
		marca := "   "
		if i == s.cursor {
			marca = " > "
			caixa = ui.Painel(interna, s.theme.Selected.Render(it.label))
		}
		for j, l := range caixa {
			prefixo := "   "
			if j == 1 {
				prefixo = marca
			}
			bloco = append(bloco, prefixo+l)
		}
		bloco = append(bloco, "")
	}

	// A arte só entra quando há largura para ela: num terminal estreito o menu
	// sozinho vale mais que um dado cortado pela metade.
	conteudo := bloco
	arte := ui.DadoArte(5, 3, 2)
	larguraArte := 0
	for _, l := range arte {
		if w := ui.Largura(l); w > larguraArte {
			larguraArte = w
		}
	}
	if s.largura >= interna+larguraArte+14 {
		conteudo = ui.LadoALado(6, bloco, arte)
	}

	tela := ui.CentralizaBloco(conteudo, s.largura, s.altura-1)
	ajuda := ui.Centraliza(s.keys.Help([]string{"setas movem", "espaco escolhe", "q sai"}), s.largura)
	return strings.Join(append(tela, ajuda), "\n")
}
