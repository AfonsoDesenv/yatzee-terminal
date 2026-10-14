package game

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

func TestComecaNaRodada(t *testing.T) {
	if got := New().Phase(); got != PhaseRound {
		t.Errorf("fase inicial = %v, quero PhaseRound", got)
	}
}

func TestTabAlternaAsFases(t *testing.T) {
	s := New()

	next, _ := s.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	s = next.(Screen)
	if s.Phase() != PhaseShop {
		t.Fatalf("depois do tab = %v, quero PhaseShop", s.Phase())
	}

	next, _ = s.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	s = next.(Screen)
	if s.Phase() != PhaseRound {
		t.Errorf("depois do segundo tab = %v, quero PhaseRound", s.Phase())
	}
}

func TestViewSegueAFase(t *testing.T) {
	// Afirma texto exclusivo de cada fase, e a ausência do texto da outra. A
	// barra de help diz "tab alterna rodada e loja", então procurar só por
	// "rodada" ou "loja" passaria em qualquer fase — inclusive se o View
	// ignorasse a fase e sempre desenhasse a rodada, que é justamente o
	// comportamento que este teste existe para garantir.
	const textoRodada = "as mecânicas de dados entram aqui"
	const textoLoja = "os itens entram aqui"

	s := New()
	v := s.View()
	if !strings.Contains(v, textoRodada) {
		t.Errorf("view da rodada não mostra o conteúdo da rodada:\n%s", v)
	}
	if strings.Contains(v, textoLoja) {
		t.Errorf("view da rodada mostra conteúdo da loja:\n%s", v)
	}

	next, _ := s.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	v = next.View()
	if !strings.Contains(v, textoLoja) {
		t.Errorf("view da loja não mostra o conteúdo da loja:\n%s", v)
	}
	if strings.Contains(v, textoRodada) {
		t.Errorf("view da loja mostra conteúdo da rodada:\n%s", v)
	}
}

// Mensagem que não é tecla não pode mexer na fase: a raiz manda WindowSizeMsg
// para a tela do topo a cada resize.
func TestMensagemNaoTeclaNaoMudaAFase(t *testing.T) {
	s := New()
	next, cmd := s.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	if got := next.(Screen).Phase(); got != PhaseRound {
		t.Errorf("fase = %v, quero PhaseRound", got)
	}
	if cmd != nil {
		t.Errorf("mensagem não-tecla devolveu comando %v", cmd())
	}
}

func TestEscVoltaProMenu(t *testing.T) {
	_, cmd := New().Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("esc não devolveu comando")
	}
	if _, ok := cmd().(ui.BackMsg); !ok {
		t.Errorf("esc emitiu %T, quero ui.BackMsg", cmd())
	}
}
