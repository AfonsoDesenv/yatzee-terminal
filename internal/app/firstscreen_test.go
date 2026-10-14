package app

import (
	"testing"

	"github.com/afonsodesenv/yatzee-terminal/internal/player"
)

// Identidade sem nome cai na tela de nome; com nome vai direto pro menu. É o
// comportamento que faz a segunda execução não perguntar nada.
func TestFirstScreen(t *testing.T) {
	casos := []struct {
		nome     string
		identity player.Identity
		quero    string
	}{
		{"sem nome pergunta", player.Identity{ID: "key:a", Kind: player.KindPubKey}, "nome"},
		{"com nome vai pro menu", player.Identity{ID: "key:a", Kind: player.KindPubKey, Name: "afonso"}, "menu"},
		{"convidado sem nome pergunta", player.Identity{ID: "guest:1", Kind: player.KindGuest}, "nome"},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			sess := sessaoDeTeste(t)
			sess.Identity = c.identity
			if got := FirstScreen(sess).Title(); got != c.quero {
				t.Errorf("FirstScreen().Title() = %q, quero %q", got, c.quero)
			}
		})
	}
}

func TestLocalIdentityLembraONomeGravado(t *testing.T) {
	sess := sessaoDeTeste(t)
	if _, err := sess.Players.Claim(ctx(), "local:fulano", "afonso"); err != nil {
		t.Fatalf("Claim: %v", err)
	}

	id, err := localIdentityFor(ctx(), "fulano", sess.Players)
	if err != nil {
		t.Fatalf("localIdentityFor: %v", err)
	}
	if id.ID != "local:fulano" {
		t.Errorf("ID = %q, quero local:fulano", id.ID)
	}
	if id.Kind != player.KindLocal {
		t.Errorf("Kind = %q, quero local", id.Kind)
	}
	if id.Name != "afonso" {
		t.Errorf("Name = %q, quero afonso (nome gravado)", id.Name)
	}
}

func TestLocalIdentitySemRegistroVemSemNome(t *testing.T) {
	sess := sessaoDeTeste(t)
	id, err := localIdentityFor(ctx(), "novato", sess.Players)
	if err != nil {
		t.Fatalf("localIdentityFor: %v", err)
	}
	if id.Name != "" {
		t.Errorf("Name = %q, quero vazio", id.Name)
	}
}
