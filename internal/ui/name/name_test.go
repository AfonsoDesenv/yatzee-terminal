package name

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/player"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

func storeDeTeste(t *testing.T) player.Store {
	t.Helper()
	return player.NewJSONStore(filepath.Join(t.TempDir(), "players.json"))
}

func telaDeTeste(t *testing.T, store player.Store) Screen {
	t.Helper()
	return New(player.Identity{ID: "key:abc", Kind: player.KindPubKey}, store)
}

// Digita texto tecla por tecla, como o runtime faria.
func digita(s Screen, texto string) Screen {
	for _, r := range texto {
		next, _ := s.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		s = next.(Screen)
	}
	return s
}

func TestForcaMinusculoNaDigitacao(t *testing.T) {
	s := digita(telaDeTeste(t, storeDeTeste(t)), "AfOnSo")
	if got := s.Value(); got != "afonso" {
		t.Errorf("Value() = %q, quero afonso", got)
	}
}

func TestEnterComNomeValidoAnunciaONome(t *testing.T) {
	store := storeDeTeste(t)
	s := digita(telaDeTeste(t, store), "afonso")

	next, cmd := s.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter com nome válido não devolveu comando")
	}
	msg := cmd()
	set, ok := msg.(ui.NameSetMsg)
	if !ok {
		t.Fatalf("comando devolveu %T, quero ui.NameSetMsg", msg)
	}
	if set.Name != "afonso" {
		t.Errorf("NameSetMsg.Name = %q, quero afonso", set.Name)
	}
	if e := next.(Screen).Error(); e != "" {
		t.Errorf("erro = %q, quero vazio", e)
	}

	// O nome tem que estar gravado, não só anunciado.
	p, ok, err := store.Get(context.Background(), "key:abc")
	if err != nil || !ok {
		t.Fatalf("Get = %v, %v, %v", p, ok, err)
	}
	if p.Name != "afonso" {
		t.Errorf("nome gravado = %q, quero afonso", p.Name)
	}
}

func TestEnterComNomeInvalidoMostraErroEPreservaOTexto(t *testing.T) {
	s := digita(telaDeTeste(t, storeDeTeste(t)), "ab")

	next, cmd := s.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd != nil {
		if msg := cmd(); msg != nil {
			if _, ok := msg.(ui.NameSetMsg); ok {
				t.Fatal("nome inválido foi aceito")
			}
		}
	}
	tela := next.(Screen)
	if tela.Error() == "" {
		t.Error("nome inválido não gerou mensagem de erro")
	}
	if tela.Value() != "ab" {
		t.Errorf("Value() = %q, quero preservar ab", tela.Value())
	}
}

func TestEnterComNomeEmUsoMostraErro(t *testing.T) {
	store := storeDeTeste(t)
	if _, err := store.Claim(context.Background(), "key:outro", "afonso"); err != nil {
		t.Fatalf("Claim de preparação: %v", err)
	}

	s := digita(telaDeTeste(t, store), "afonso")
	next, _ := s.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	tela := next.(Screen)

	if !strings.Contains(tela.Error(), "uso") {
		t.Errorf("Error() = %q, quero mencionar que está em uso", tela.Error())
	}
	if tela.Value() != "afonso" {
		t.Errorf("Value() = %q, quero preservar o texto", tela.Value())
	}
}

// Convidado não reserva nome, mas também não pode se passar por registrado.
func TestConvidadoComNomeDeRegistradoERecusado(t *testing.T) {
	store := storeDeTeste(t)
	if _, err := store.Claim(context.Background(), "key:outro", "afonso"); err != nil {
		t.Fatalf("Claim de preparação: %v", err)
	}

	s := New(player.Identity{ID: "guest:1", Kind: player.KindGuest}, store)
	s = digita(s, "afonso")
	next, _ := s.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if next.(Screen).Error() == "" {
		t.Error("convidado usou nome de jogador registrado sem erro")
	}
}

func TestConvidadoNaoEPersistido(t *testing.T) {
	store := storeDeTeste(t)
	s := New(player.Identity{ID: "guest:1", Kind: player.KindGuest}, store)
	s = digita(s, "visitante")

	next, cmd := s.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter não devolveu comando")
	}
	if _, ok := cmd().(ui.NameSetMsg); !ok {
		t.Fatal("convidado com nome livre deveria ser aceito")
	}
	_ = next

	if _, ok, err := store.Get(context.Background(), "guest:1"); err != nil || ok {
		t.Errorf("convidado foi persistido: ok=%v err=%v", ok, err)
	}
}

func TestViewMostraARegra(t *testing.T) {
	v := telaDeTeste(t, storeDeTeste(t)).View()
	if !strings.Contains(v, "16") {
		t.Errorf("View() não mostra a regra do nome:\n%s", v)
	}
}
