// Package app monta o programa e roteia as telas.
package app

import (
	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui/game"
)

// Root é o model raiz: dono da pilha de telas e único lugar que conhece o
// grafo de navegação. Telas emitem intenção; a tradução em empilhar ou voltar
// acontece só aqui, e é o que permite nenhuma tela importar outra.
type Root struct {
	stack  []ui.Screen
	keys   ui.Keys
	width  int
	height int
}

// NewRoot monta a raiz com uma tela na base da pilha.
func NewRoot(initial ui.Screen) Root {
	return Root{
		stack:  []ui.Screen{initial},
		keys:   ui.DefaultKeys(),
		width:  ui.DefaultWidth,
		height: ui.DefaultHeight,
	}
}

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
		w, h := ui.Size(msg.Width, msg.Height)
		m.width, m.height = w, h
		if w != msg.Width || h != msg.Height {
			// O renderer só se redimensiona com um WindowSizeMsg que passe pelo
			// laço do Bubble Tea. Um tamanho zerado deixaria o renderer vazio e
			// a tela em branco, então reemitimos o tamanho corrigido. Sem laço:
			// o tamanho normalizado é ponto fixo de ui.Size, então a mensagem
			// reemitida não cai neste ramo de novo — e é ela que chega às telas.
			return m, func() tea.Msg { return tea.WindowSizeMsg{Width: w, Height: h} }
		}
		// Tamanho já normalizado: a tela do topo pode se organizar com ele.
		return m.updateTop(tea.WindowSizeMsg{Width: w, Height: h})

	case tea.KeyPressMsg:
		// Sai de qualquer lugar, sem passar pela tela. Vem do mapa de teclas,
		// não de um literal: keybinding mora num lugar só.
		if m.keys.Matches(msg.String(), m.keys.ForceQuit) {
			return m, tea.Quit
		}

	case ui.QuitMsg:
		return m, tea.Quit

	case ui.StartGameMsg:
		return m.pushWith(game.New())

	case ui.BackMsg:
		return m.pop(), nil
	}

	return m.updateTop(msg)
}

// updateTop entrega a mensagem à tela do topo, sem mexer na pilha anterior.
func (m Root) updateTop(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		return m // base da pilha: voltar não encerra o jogo
	}
	m = m.copyStack()
	m.stack = m.stack[:len(m.stack)-1]
	return m
}

// pushWith empilha e arma a tela. O Bubble Tea chama Init só no model raiz, uma
// vez; sem devolver o Init da tela empilhada aqui, ela nunca carregaria nada —
// é o defeito que deixou a tela de placar presa em "carregando" antes.
func (m Root) pushWith(s ui.Screen) (Root, tea.Cmd) {
	m = m.push(s)
	return m, s.Init()
}
