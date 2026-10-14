package app

import (
	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui/game"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui/menu"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui/scores"
)

// Root é o model raiz: dono da pilha de telas e único lugar que conhece o
// grafo de navegação. Telas emitem intenção; a tradução em empilhar, voltar ou
// trocar acontece só aqui.
type Root struct {
	sess   *Session
	stack  []ui.Screen
	keys   ui.Keys
	width  int
	height int
}

// NewRoot monta a raiz com uma tela na base da pilha.
func NewRoot(sess *Session, initial ui.Screen) Root {
	return Root{
		sess:   sess,
		stack:  []ui.Screen{initial},
		keys:   ui.DefaultKeys(),
		width:  ui.DefaultWidth,
		height: ui.DefaultHeight,
	}
}

// Session dá acesso à sessão desta conexão.
func (m Root) Session() *Session { return m.sess }

// Top é a tela visível.
func (m Root) Top() ui.Screen { return m.stack[len(m.stack)-1] }

// Depth é o tamanho da pilha.
func (m Root) Depth() int { return len(m.stack) }

// Size devolve o tamanho já normalizado.
func (m Root) Size() (int, int) { return m.width, m.height }

// Init arranca a tela da base.
func (m Root) Init() tea.Cmd { return m.Top().Init() }

// Update roteia. A ordem importa: tamanho e saída global primeiro, intenções
// de navegação depois, e só então a mensagem desce para a tela do topo.
func (m Root) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = ui.Size(msg.Width, msg.Height)
		return m, nil

	case tea.KeyPressMsg:
		// Sai de qualquer lugar, sem passar pela tela. Vem do mapa de teclas,
		// não de um literal: keybinding mora num lugar só.
		if m.keys.Matches(msg.String(), m.keys.ForceQuit) {
			return m, tea.Quit
		}

	case ui.QuitMsg:
		return m, tea.Quit

	case ui.ShowScoresMsg:
		return m.push(scores.New(m.sess.Scores)), nil

	case ui.StartGameMsg:
		return m.push(game.New()), nil

	case ui.BackMsg:
		return m.pop(), nil

	case ui.NameSetMsg:
		// Troca em vez de empilhar: depois de se apresentar, esc não pode
		// levar de volta pra tela de nome.
		m.sess.Identity.Name = msg.Name
		return m.replace(menu.New(m.sess.Identity.Name)), nil
	}

	top := len(m.stack) - 1
	next, cmd := m.stack[top].Update(msg)
	m = m.copyStack()
	m.stack[top] = next
	return m, cmd
}

// View é o único lugar que lida com tea.View.
func (m Root) View() tea.View {
	v := tea.NewView(m.Top().View())
	v.AltScreen = true
	return v
}

// copyStack copia a pilha antes de modificar. Root é passado por valor, então
// sem a cópia duas versões do model compartilhariam o mesmo array de trás e
// uma escrita numa apareceria na outra.
func (m Root) copyStack() Root {
	next := make([]ui.Screen, len(m.stack))
	copy(next, m.stack)
	m.stack = next
	return m
}

func (m Root) push(s ui.Screen) Root {
	m = m.copyStack()
	m.stack = append(m.stack, s)
	return m
}

func (m Root) pop() Root {
	if len(m.stack) <= 1 {
		return m // base da pilha: voltar não encerra a sessão
	}
	m = m.copyStack()
	m.stack = m.stack[:len(m.stack)-1]
	return m
}

func (m Root) replace(s ui.Screen) Root {
	m = m.copyStack()
	m.stack[len(m.stack)-1] = s
	return m
}
