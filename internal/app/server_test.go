package app

import (
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/afonsodesenv/yatzee-terminal/internal/player"
	gossh "golang.org/x/crypto/ssh"
)

// Chave ed25519 fixa, só para o teste ter um fingerprint estável.
const chaveDeTeste = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEnTeDU2pMM3YIoG1dLS1hmVoVVfp9AX8tSCVIxH1xH1 teste@yatzee"

func TestIdentityFromKeyUsaOFingerprint(t *testing.T) {
	key, _, _, _, err := gossh.ParseAuthorizedKey([]byte(chaveDeTeste))
	if err != nil {
		t.Fatalf("parse da chave de teste: %v", err)
	}

	id := IdentityFromKey(key)
	if id.Kind != player.KindPubKey {
		t.Errorf("Kind = %q, quero pubkey", id.Kind)
	}
	if !strings.HasPrefix(id.ID, "key:SHA256:") {
		t.Errorf("ID = %q, quero prefixo key:SHA256:", id.ID)
	}
	if id.Name != "" {
		t.Errorf("Name = %q, quero vazio (o nome vem do store)", id.Name)
	}

	// Mesma chave, mesma identidade: é o que faz a pessoa ser reconhecida.
	if outro := IdentityFromKey(key); outro.ID != id.ID {
		t.Errorf("mesma chave deu IDs diferentes: %q e %q", id.ID, outro.ID)
	}
}

func TestGuestIdentityNaoPersisteENaoRepete(t *testing.T) {
	rnd := rand.New(rand.NewPCG(1, 2))

	a := GuestIdentity(rnd)
	b := GuestIdentity(rnd)

	if a.Kind != player.KindGuest {
		t.Errorf("Kind = %q, quero guest", a.Kind)
	}
	if a.Persistent() {
		t.Error("convidado marcado como persistente")
	}
	if !strings.HasPrefix(a.ID, "guest:") {
		t.Errorf("ID = %q, quero prefixo guest:", a.ID)
	}
	if a.ID == b.ID {
		t.Errorf("dois convidados com o mesmo ID: %q", a.ID)
	}
}

func TestConvidadosDeSessoesDiferentesNaoColidem(t *testing.T) {
	a := sessaoDeTeste(t)
	b := sessaoDeTeste(t)

	// O que importa não é o gerador avançar, e sim cada sessão ter o seu
	// próprio: com um gerador global compartilhado, duas sessões poderiam
	// produzir o mesmo id de convidado.
	if GuestIdentity(a.Rand).ID == GuestIdentity(b.Rand).ID {
		t.Error("duas sessões geraram o mesmo id de convidado")
	}
}
