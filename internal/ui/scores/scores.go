// Package scores mostra os melhores jogadores.
package scores

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/score"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

// Quantas linhas o placar mostra.
const topN = 10

// LoadedMsg traz o resultado da leitura do placar.
type LoadedMsg struct {
	Entries []score.Entry
	Err     error
}

// Screen mostra o placar. Recebe só o store de que precisa, não a sessão
// inteira: é o que mantém esta tela testável com um fake de dois métodos.
type Screen struct {
	store   score.Store
	entries []score.Entry
	err     error
	loaded  bool
	theme   ui.Theme
	keys    ui.Keys
}

// New monta a tela.
func New(store score.Store) Screen {
	return Screen{
		store: store,
		theme: ui.NewTheme(),
		keys:  ui.DefaultKeys(),
	}
}

// Title identifica a tela.
func (s Screen) Title() string { return "scores" }

// Init dispara a leitura fora da thread de render: I/O não bloqueia a UI.
func (s Screen) Init() tea.Cmd {
	store := s.store
	return func() tea.Msg {
		entries, err := store.Top(context.Background(), topN)
		return LoadedMsg{Entries: entries, Err: err}
	}
}

// Update recebe o resultado da carga e trata o esc.
func (s Screen) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case LoadedMsg:
		s.entries, s.err, s.loaded = msg.Entries, msg.Err, true
		return s, nil
	case tea.KeyPressMsg:
		if s.keys.Matches(msg.String(), s.keys.Back) {
			return s, ui.Back
		}
	}
	return s, nil
}

// View desenha a lista, o estado vazio ou a falha.
func (s Screen) View() string {
	b := &strings.Builder{}
	b.WriteString(s.theme.Title.Render("placar"))
	b.WriteString("\n\n")

	switch {
	case s.err != nil:
		b.WriteString(s.theme.Error.Render("não consegui ler o placar agora"))
	case !s.loaded:
		b.WriteString(s.theme.Dim.Render("carregando..."))
	case len(s.entries) == 0:
		b.WriteString(s.theme.Dim.Render("nenhuma partida registrada ainda"))
	default:
		for i, e := range s.entries {
			line := fmt.Sprintf("%2d. %-16s %6d", i+1, e.Name, e.Points)
			b.WriteString(s.theme.Item.Render(line))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(s.keys.Help([]string{"esc volta"}))
	return s.theme.Panel.Render(b.String())
}
