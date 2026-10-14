package score

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func novoStore(t *testing.T) *JSONStore {
	t.Helper()
	return NewJSONStore(filepath.Join(t.TempDir(), "scores.json"))
}

func TestTopVazio(t *testing.T) {
	got, err := novoStore(t).Top(context.Background(), 10)
	if err != nil {
		t.Fatalf("Top: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("Top em store vazio = %v, quero vazio", got)
	}
}

func TestSaveETopOrdenaPorPontosDesc(t *testing.T) {
	ctx := context.Background()
	s := novoStore(t)
	base := time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC)

	entradas := []Entry{
		{PlayerID: "key:a", Name: "ana", Points: 120, At: base},
		{PlayerID: "key:b", Name: "bob", Points: 300, At: base.Add(time.Minute)},
		{PlayerID: "key:c", Name: "caio", Points: 200, At: base.Add(2 * time.Minute)},
	}
	for _, e := range entradas {
		if err := s.Save(ctx, e); err != nil {
			t.Fatalf("Save(%s): %v", e.Name, err)
		}
	}

	got, err := s.Top(ctx, 10)
	if err != nil {
		t.Fatalf("Top: %v", err)
	}
	quero := []string{"bob", "caio", "ana"}
	if len(got) != len(quero) {
		t.Fatalf("Top devolveu %d entradas, quero %d", len(got), len(quero))
	}
	for i, nome := range quero {
		if got[i].Name != nome {
			t.Errorf("Top[%d].Name = %q, quero %q", i, got[i].Name, nome)
		}
	}
}

func TestTopEmpateDesempatatePeloMaisAntigo(t *testing.T) {
	ctx := context.Background()
	s := novoStore(t)
	base := time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC)

	if err := s.Save(ctx, Entry{PlayerID: "key:b", Name: "bob", Points: 100, At: base.Add(time.Hour)}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := s.Save(ctx, Entry{PlayerID: "key:a", Name: "ana", Points: 100, At: base}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := s.Top(ctx, 10)
	if err != nil {
		t.Fatalf("Top: %v", err)
	}
	if got[0].Name != "ana" {
		t.Errorf("Top[0] = %q, quero ana (chegou primeiro no empate)", got[0].Name)
	}
}

func TestTopRespeitaOLimite(t *testing.T) {
	ctx := context.Background()
	s := novoStore(t)
	for i := range 5 {
		e := Entry{PlayerID: "key:x", Name: "x", Points: i, At: time.Now().UTC()}
		if err := s.Save(ctx, e); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}

	got, err := s.Top(ctx, 3)
	if err != nil {
		t.Fatalf("Top: %v", err)
	}
	if len(got) != 3 {
		t.Errorf("Top(3) devolveu %d entradas", len(got))
	}
}

func TestSaveConcorrenteNaoPerdeEntrada(t *testing.T) {
	ctx := context.Background()
	s := novoStore(t)

	const n = 20
	done := make(chan error, n)
	for i := range n {
		go func(i int) {
			done <- s.Save(ctx, Entry{PlayerID: "key:x", Name: "x", Points: i, At: time.Now().UTC()})
		}(i)
	}
	for range n {
		if err := <-done; err != nil {
			t.Fatalf("Save concorrente: %v", err)
		}
	}
	got, err := s.Top(ctx, 100)
	if err != nil {
		t.Fatalf("Top: %v", err)
	}
	if len(got) != n {
		t.Errorf("guardou %d entradas, quero %d", len(got), n)
	}
}
