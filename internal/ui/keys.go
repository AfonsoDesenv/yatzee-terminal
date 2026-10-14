package ui

import "strings"

// Keys é o mapa de teclas, num lugar só para a barra de help nunca mentir
// sobre o que o jogo aceita.
type Keys struct {
	Up        []string
	Down      []string
	Left      []string
	Right     []string
	Confirm   []string
	Back      []string
	Toggle    []string
	Quit      []string
	ForceQuit []string
	theme     Theme
}

// DefaultKeys devolve o mapa do spec.
func DefaultKeys() Keys {
	return Keys{
		Up:    []string{"up", "k"},
		Down:  []string{"down", "j"},
		Left:  []string{"left", "h"},
		Right: []string{"right", "l"},
		// Espaço é o confirmar do jogo; enter continua valendo porque é o que
		// a mão procura sozinha, e aceitar os dois não custa nada.
		Confirm:   []string{"space", "enter"},
		Back:      []string{"esc"},
		Toggle:    []string{"tab"},
		Quit:      []string{"q"},
		ForceQuit: []string{"ctrl+c"},
		theme:     NewTheme(),
	}
}

// Matches diz se a tecla pressionada está num dos conjuntos.
func (k Keys) Matches(pressed string, set []string) bool {
	for _, s := range set {
		if pressed == s {
			return true
		}
	}
	return false
}

// Help monta a barra de dicas no pé da tela.
func (k Keys) Help(hints []string) string {
	return k.theme.Dim.Render(strings.Join(hints, "  ·  "))
}
