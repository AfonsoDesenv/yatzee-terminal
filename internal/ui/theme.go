package ui

import "charm.land/lipgloss/v2"

// Theme concentra todo estilo num lugar, para a aparência mudar sem caçar
// cores espalhadas pelas telas. Sem Renderer: a lipgloss v2 não tem esse
// conceito, e o perfil de cor do cliente é resolvido pelo programa.
type Theme struct {
	Title    lipgloss.Style
	Item     lipgloss.Style
	Selected lipgloss.Style
	Dim      lipgloss.Style
	Error    lipgloss.Style
	Panel    lipgloss.Style

	// As cores de estado. Cada uma acompanha um sinal feito de caractere
	// (^^^^, >  <), e não substitui ele: negrito e cor são um pedido ao
	// terminal, e onde o pedido é ignorado o desenho tem de se explicar só.
	//
	// DadoMarcado é o dado escolhido para o reroll; DadoContando é o que está
	// somando agora. São estados diferentes e pedem cores diferentes, porque
	// vermelho já quer dizer "vai ser descartado".
	DadoMarcado  lipgloss.Style
	DadoContando lipgloss.Style
	// Runa é o roxo da runa, e serve ao foco na loja e à runa que está somando
	// na mesa: roxo quer dizer runa, nos dois lugares.
	Runa        lipgloss.Style
	BotaoEmFoco lipgloss.Style

	// Fogo vai da ponta que apaga até a brasa. Cinco níveis para os dez
	// degraus da rampa de densidade: menos cores significam trechos maiores
	// e menos escape por quadro, e a chama repinta a tela inteira vinte vezes
	// por segundo.
	Fogo []lipgloss.Style
}

// chama monta a paleta do fogo a partir das cores, do frio para o quente.
func chama(cores ...string) []lipgloss.Style {
	out := make([]lipgloss.Style, len(cores))
	for i, c := range cores {
		out[i] = lipgloss.NewStyle().Foreground(lipgloss.Color(c))
	}
	return out
}

// NewTheme devolve o tema padrão.
func NewTheme() Theme {
	return Theme{
		Title:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#b58cff")),
		Item:     lipgloss.NewStyle().Foreground(lipgloss.Color("#d8d8d8")),
		Selected: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ffd866")),
		Dim:      lipgloss.NewStyle().Foreground(lipgloss.Color("#8a8a8a")),
		Error:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ff6b6b")),
		Panel:    lipgloss.NewStyle().Padding(1, 2),

		DadoMarcado:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ff5555")),
		DadoContando: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#4fc3f7")),
		Runa:         lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#b58cff")),
		BotaoEmFoco:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#5af78e")),

		Fogo: chama(
			"#8c2500", // a ponta esfriando
			"#d94e00",
			"#ff7518", // o corpo da chama
			"#ffa500",
			"#ffd166", // a brasa, na base
		),
	}
}
