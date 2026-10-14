package ui

import (
	"strings"
	"testing"
)

// Um cliente SSH pode anunciar PTY 0x0 — já aconteceu no probe do spec — e com
// largura zero nada é desenhado. O fallback é o que evita tela branca.
func TestSizeFallback(t *testing.T) {
	casos := []struct {
		nome   string
		w, h   int
		queroW int
		queroH int
	}{
		{"zero virou default", 0, 0, DefaultWidth, DefaultHeight},
		{"negativo virou default", -5, -1, DefaultWidth, DefaultHeight},
		{"abaixo do piso sobe pro piso", 10, 2, MinWidth, MinHeight},
		{"tamanho normal passa intacto", 120, 40, 120, 40},
		{"só a largura zerada", 0, 30, DefaultWidth, 30},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			w, h := Size(c.w, c.h)
			if w != c.queroW || h != c.queroH {
				t.Errorf("Size(%d, %d) = (%d, %d), quero (%d, %d)", c.w, c.h, w, h, c.queroW, c.queroH)
			}
		})
	}
}

func TestHelpJuntaAsDicas(t *testing.T) {
	got := DefaultKeys().Help([]string{"enter confirma", "esc volta"})
	if got == "" {
		t.Fatal("Help devolveu vazio")
	}
	for _, termo := range []string{"enter", "esc"} {
		if !strings.Contains(got, termo) {
			t.Errorf("Help() = %q, quero conter %q", got, termo)
		}
	}
}
