// Package ui tem o contrato de tela, o tema e as telas do jogo. Não conhece
// persistência: cada tela recebe só a dependência que usa.
package ui

import tea "charm.land/bubbletea/v2"

// Screen é uma tela do jogo. Devolve string, não tea.View: só o model raiz
// lida com tea.View, então as telas ficam triviais de testar e de compor.
type Screen interface {
	Init() tea.Cmd
	Update(tea.Msg) (Screen, tea.Cmd)
	View() string
	Title() string
}

// Mensagens de intenção. Uma tela diz o que quer; quem decide para onde ir é o
// model raiz, único dono do grafo de navegação. É o que permite nenhuma tela
// importar outra.
type (
	// StartGameMsg pede uma partida nova.
	StartGameMsg struct{}
	// BackMsg sai da tela atual.
	BackMsg struct{}
	// QuitMsg encerra o jogo.
	QuitMsg struct{}
)

// Comandos prontos, para a tela não precisar escrever closure na mão.
func StartGame() tea.Msg { return StartGameMsg{} }
func Back() tea.Msg      { return BackMsg{} }
func Quit() tea.Msg      { return QuitMsg{} }
