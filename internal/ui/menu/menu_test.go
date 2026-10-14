package menu

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

func tecla(s Screen, nome string) (Screen, tea.Cmd) {
	next, cmd := s.Update(teclaMsg(nome))
	return next.(Screen), cmd
}

func teclaMsg(nome string) tea.KeyPressMsg {
	switch nome {
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "down":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	case "up":
		return tea.KeyPressMsg{Code: tea.KeyUp}
	default:
		r := rune(nome[0])
		return tea.KeyPressMsg{Code: r, Text: nome}
	}
}

func TestCursorAndaEParaNasPontas(t *testing.T) {
	s := New("afonso")
	if s.Cursor() != 0 {
		t.Fatalf("cursor inicial = %d, quero 0", s.Cursor())
	}

	s, _ = tecla(s, "down")
	if s.Cursor() != 1 {
		t.Errorf("cursor após down = %d, quero 1", s.Cursor())
	}
	s, _ = tecla(s, "j")
	if s.Cursor() != 2 {
		t.Errorf("cursor após j = %d, quero 2", s.Cursor())
	}
	s, _ = tecla(s, "down") // já no fim
	if s.Cursor() != 2 {
		t.Errorf("cursor após down no fim = %d, quero 2", s.Cursor())
	}
	s, _ = tecla(s, "k")
	if s.Cursor() != 1 {
		t.Errorf("cursor após k = %d, quero 1", s.Cursor())
	}

	// "up" literal tem que mover de verdade antes de o teste do limite
	// significar algo: se o mapa de teclas perdesse "up" e ficasse só com
	// "k", um "up" no topo continuaria "parando" em 0 por não fazer nada.
	s, _ = tecla(s, "up")
	if s.Cursor() != 0 {
		t.Errorf("cursor após up = %d, quero 0", s.Cursor())
	}

	s, _ = tecla(s, "up") // agora sim: já no topo, não deve passar do limite
	if s.Cursor() != 0 {
		t.Errorf("cursor após up no topo = %d, quero 0", s.Cursor())
	}
}

func TestEnterEmitAIntencaoDoItemSelecionado(t *testing.T) {
	casos := []struct {
		descidas int
		quero    tea.Msg
	}{
		{0, ui.StartGameMsg{}},
		{1, ui.ShowScoresMsg{}},
		{2, ui.QuitMsg{}},
	}
	for _, c := range casos {
		s := New("afonso")
		for range c.descidas {
			s, _ = tecla(s, "down")
		}
		_, cmd := tecla(s, "enter")
		if cmd == nil {
			t.Fatalf("item %d: enter não devolveu comando", c.descidas)
		}
		if got := cmd(); got != c.quero {
			t.Errorf("item %d emitiu %T, quero %T", c.descidas, got, c.quero)
		}
	}
}

func TestTeclaQSai(t *testing.T) {
	_, cmd := tecla(New("afonso"), "q")
	if cmd == nil {
		t.Fatal("q não devolveu comando")
	}
	if _, ok := cmd().(ui.QuitMsg); !ok {
		t.Errorf("q emitiu %T, quero ui.QuitMsg", cmd())
	}
}

// esc no menu não faz nada: é a base da pilha, e sair exige q ou ctrl+c.
func TestEscNaoEmiteNada(t *testing.T) {
	_, cmd := New("afonso").Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd != nil {
		t.Errorf("esc no menu devolveu comando %v", cmd())
	}
}

func TestViewMostraONomeEOsItens(t *testing.T) {
	v := New("afonso").View()
	for _, termo := range []string{"afonso", "novo jogo", "placar", "sair"} {
		if !strings.Contains(v, termo) {
			t.Errorf("View() não contém %q:\n%s", termo, v)
		}
	}
}
