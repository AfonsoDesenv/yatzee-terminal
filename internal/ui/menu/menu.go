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
	items  []item
	cursor int
	theme  ui.Theme
	keys   ui.Keys
}

// New monta o menu. A tabela carrega a intenção de cada item, então Update não
// tem switch por índice e acrescentar um item é acrescentar uma linha.
func New() Screen {
	return Screen{
		items: []item{
			{label: "novo jogo", intent: ui.StartGame},
			{label: "sair", intent: ui.Quit},
		},
		theme: ui.NewTheme(),
		keys:  ui.DefaultKeys(),
	}
}

// Cursor é o índice do item selecionado.
func (s Screen) Cursor() int { return s.cursor }

// Title identifica a tela.
func (s Screen) Title() string { return "menu" }

// Init não precisa de nada.
func (s Screen) Init() tea.Cmd { return nil }

// Update move o cursor e emite a intenção do item escolhido.
func (s Screen) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return s, nil
	}
	pressed := key.String()

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
	return s, nil
}

// View desenha o título, os itens e o help.
func (s Screen) View() string {
	b := &strings.Builder{}
	b.WriteString(s.theme.Title.Render("yatzee"))
	b.WriteString("\n\n")

	for i, it := range s.items {
		if i == s.cursor {
			b.WriteString(s.theme.Selected.Render("› " + it.label))
		} else {
			b.WriteString(s.theme.Item.Render("  " + it.label))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(s.keys.Help([]string{"↑↓ move", "enter escolhe", "q sai"}))
	return s.theme.Panel.Render(b.String())
}
