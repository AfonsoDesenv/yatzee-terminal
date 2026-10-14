package scores

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/score"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

// storeFake evita I/O no teste de tela e permite simular falha de leitura.
type storeFake struct {
	entries []score.Entry
	err     error
}

func (f storeFake) Top(context.Context, int) ([]score.Entry, error) {
	return f.entries, f.err
}
func (f storeFake) Save(context.Context, score.Entry) error { return nil }

func TestInitCarregaOPlacar(t *testing.T) {
	base := time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC)
	store := storeFake{entries: []score.Entry{
		{Name: "ana", Points: 300, At: base},
	}}

	cmd := New(store).Init()
	if cmd == nil {
		t.Fatal("Init não devolveu comando de carga")
	}
	msg, ok := cmd().(LoadedMsg)
	if !ok {
		t.Fatalf("Init emitiu %T, quero LoadedMsg", cmd())
	}
	if len(msg.Entries) != 1 || msg.Entries[0].Name != "ana" {
		t.Errorf("LoadedMsg.Entries = %+v", msg.Entries)
	}
}

func TestViewListaAsEntradas(t *testing.T) {
	base := time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC)
	s := New(storeFake{})
	next, _ := s.Update(LoadedMsg{Entries: []score.Entry{
		{Name: "ana", Points: 300, At: base},
		{Name: "bob", Points: 120, At: base},
	}})

	v := next.View()
	for _, termo := range []string{"ana", "300", "bob", "120"} {
		if !strings.Contains(v, termo) {
			t.Errorf("View() não contém %q:\n%s", termo, v)
		}
	}
}

// Sem gameplay ainda não existem partidas: a tela precisa dizer isso em vez de
// mostrar um quadro vazio.
func TestViewComPlacarVazioExplica(t *testing.T) {
	s := New(storeFake{})
	next, _ := s.Update(LoadedMsg{Entries: nil})

	if !strings.Contains(next.View(), "nenhuma partida") {
		t.Errorf("View() vazia não explica o estado:\n%s", next.View())
	}
}

func TestViewComErroDeLeitura(t *testing.T) {
	s := New(storeFake{})
	next, _ := s.Update(LoadedMsg{Err: errors.New("disco pegou fogo")})

	if !strings.Contains(next.View(), "não consegui") {
		t.Errorf("View() não reporta a falha:\n%s", next.View())
	}
}

func TestEscVolta(t *testing.T) {
	_, cmd := New(storeFake{}).Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("esc não devolveu comando")
	}
	if _, ok := cmd().(ui.BackMsg); !ok {
		t.Errorf("esc emitiu %T, quero ui.BackMsg", cmd())
	}
}
