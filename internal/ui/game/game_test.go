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
	s := New()
	if !strings.Contains(s.View(), "rodada") {
		t.Errorf("view da rodada não menciona rodada:\n%s", s.View())
	}

	next, _ := s.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if !strings.Contains(next.View(), "loja") {
		t.Errorf("view da loja não menciona loja:\n%s", next.View())
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
