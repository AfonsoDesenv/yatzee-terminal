package ui

import "testing"

// Espaço é o confirmar do jogo. Mora no mapa de teclas, e não espalhado pelas
// telas, justamente para valer em todas de uma vez — menu, loja e mesa leem
// daqui.
func TestEspacoConfirma(t *testing.T) {
	k := DefaultKeys()

	if !k.Matches("space", k.Confirm) {
		t.Error("espaço não confirma")
	}
	// Enter segue valendo: é o que a mão procura sozinha, e aceitar os dois
	// não tira nada de ninguém.
	if !k.Matches("enter", k.Confirm) {
		t.Error("enter deixou de confirmar")
	}
}

// As quatro setas têm nome no mapa; sem isso a loja teria de comparar com
// literais e a barra de ajuda poderia mentir sobre o que o jogo aceita.
func TestAsQuatroSetasTemNome(t *testing.T) {
	k := DefaultKeys()

	casos := []struct {
		tecla string
		set   []string
		nome  string
	}{
		{"up", k.Up, "cima"},
		{"down", k.Down, "baixo"},
		{"left", k.Left, "esquerda"},
		{"right", k.Right, "direita"},
	}
	for _, c := range casos {
		if !k.Matches(c.tecla, c.set) {
			t.Errorf("a seta para %s não está no mapa", c.nome)
		}
	}

	// Nenhuma seta pode cair em dois conjuntos: o mesmo toque andando para
	// dois lados seria imprevisível.
	if k.Matches("left", k.Right) || k.Matches("up", k.Down) {
		t.Error("uma seta responde por dois movimentos")
	}
}
