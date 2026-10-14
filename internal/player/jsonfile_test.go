package player

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func novoStore(t *testing.T) *JSONStore {
	t.Helper()
	return NewJSONStore(filepath.Join(t.TempDir(), "players.json"))
}

func TestClaimEGet(t *testing.T) {
	ctx := context.Background()
	s := novoStore(t)

	p, err := s.Claim(ctx, "key:abc", "afonso")
	if err != nil {
		t.Fatalf("Claim: %v", err)
	}
	if p.Name != "afonso" || p.ID != "key:abc" {
		t.Fatalf("Claim devolveu %+v", p)
	}
	if p.FirstSeen.IsZero() || p.LastSeen.IsZero() {
		t.Errorf("Claim não preencheu os timestamps: %+v", p)
	}

	got, ok, err := s.Get(ctx, "key:abc")
	if err != nil || !ok {
		t.Fatalf("Get = %v, %v, %v", got, ok, err)
	}
	if got.Name != "afonso" {
		t.Errorf("Get().Name = %q, quero afonso", got.Name)
	}
}

func TestGetIdentidadeDesconhecida(t *testing.T) {
	_, ok, err := novoStore(t).Get(context.Background(), "key:nao-existe")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if ok {
		t.Error("Get de id desconhecido devolveu ok=true")
	}
}

func TestClaimNomeEmUsoPorOutraIdentidade(t *testing.T) {
	ctx := context.Background()
	s := novoStore(t)

	if _, err := s.Claim(ctx, "key:abc", "afonso"); err != nil {
		t.Fatalf("primeiro Claim: %v", err)
	}
	if _, err := s.Claim(ctx, "key:xyz", "afonso"); !errors.Is(err, ErrNameTaken) {
		t.Errorf("segundo Claim = %v, quero ErrNameTaken", err)
	}
}

func TestClaimDoMesmoIdComMesmoNomeNaoFalha(t *testing.T) {
	ctx := context.Background()
	s := novoStore(t)

	if _, err := s.Claim(ctx, "key:abc", "afonso"); err != nil {
		t.Fatalf("primeiro Claim: %v", err)
	}
	if _, err := s.Claim(ctx, "key:abc", "afonso"); err != nil {
		t.Errorf("reivindicar o próprio nome falhou: %v", err)
	}
}

func TestNameTaken(t *testing.T) {
	ctx := context.Background()
	s := novoStore(t)
	if _, err := s.Claim(ctx, "key:abc", "afonso"); err != nil {
		t.Fatalf("Claim: %v", err)
	}

	taken, err := s.NameTaken(ctx, "afonso")
	if err != nil || !taken {
		t.Errorf("NameTaken(afonso) = %v, %v; quero true, nil", taken, err)
	}
	taken, err = s.NameTaken(ctx, "outro")
	if err != nil || taken {
		t.Errorf("NameTaken(outro) = %v, %v; quero false, nil", taken, err)
	}
}

func TestTouchAtualizaLastSeen(t *testing.T) {
	ctx := context.Background()
	s := novoStore(t)
	p, err := s.Claim(ctx, "key:abc", "afonso")
	if err != nil {
		t.Fatalf("Claim: %v", err)
	}

	if err := s.Touch(ctx, "key:abc"); err != nil {
		t.Fatalf("Touch: %v", err)
	}
	got, _, err := s.Get(ctx, "key:abc")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !got.LastSeen.After(p.LastSeen) {
		t.Errorf("LastSeen nao avancou: %v nao e depois de %v", got.LastSeen, p.LastSeen)
	}
	if !got.FirstSeen.Equal(p.FirstSeen) {
		t.Errorf("Touch mexeu no FirstSeen: %v != %v", got.FirstSeen, p.FirstSeen)
	}
}

func TestArquivoNasceCom0600(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	path := filepath.Join(dir, "players.json")
	s := NewJSONStore(path)
	if _, err := s.Claim(ctx, "key:abc", "afonso"); err != nil {
		t.Fatalf("Claim: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("permissão = %o, quero 600", perm)
	}
}

func TestEscritaConcorrenteNaoCorrompe(t *testing.T) {
	ctx := context.Background()
	s := novoStore(t)

	const n = 20
	done := make(chan error, n)
	for i := range n {
		go func(i int) {
			_, err := s.Claim(ctx, fmt.Sprintf("key:%d", i), fmt.Sprintf("jogador-%d", i))
			done <- err
		}(i)
	}
	for range n {
		if err := <-done; err != nil {
			t.Fatalf("Claim concorrente: %v", err)
		}
	}
	for i := range n {
		if _, ok, err := s.Get(ctx, fmt.Sprintf("key:%d", i)); err != nil || !ok {
			t.Errorf("identidade %d sumiu: ok=%v err=%v", i, ok, err)
		}
	}
}
