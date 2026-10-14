package player

import (
	"errors"
	"strings"
	"testing"
)

func TestCleanNameAceita(t *testing.T) {
	casos := []struct {
		entrada string
		quero   string
	}{
		{"afonso", "afonso"},
		{"afo_nso-1", "afo_nso-1"},
		{"afonso dev", "afonso dev"},
		{"abc", "abc"},
		{"1234567890123456", "1234567890123456"},
		{"  afonso  ", "afonso"}, // trim antes de validar
	}
	for _, c := range casos {
		got, err := CleanName(c.entrada)
		if err != nil {
			t.Errorf("CleanName(%q) = erro %v, quero aceitar", c.entrada, err)
			continue
		}
		if got != c.quero {
			t.Errorf("CleanName(%q) = %q, quero %q", c.entrada, got, c.quero)
		}
	}
}

func TestCleanNameRejeita(t *testing.T) {
	casos := []struct {
		entrada string
		motivo  string
	}{
		{"Afonso", "maiúscula"},
		{"AFONSO", "tudo maiúsculo"},
		{"ab", "curto demais"},
		{"12345678901234567", "longo demais"},
		{"a  b", "dois espaços seguidos"},
		{"a b c", "dois espaços separados"},
		{"joão", "acento"},
		{"afonso!", "pontuação"},
		{"___", "sem alfanumérico"},
		{"---", "sem alfanumérico"},
		{"", "vazio"},
		{"   ", "só espaço"},
		{"af\x1b[31monso", "ANSI embutido"},
		{"afonso\n", "controle"},
	}
	for _, c := range casos {
		if _, err := CleanName(c.entrada); !errors.Is(err, ErrNameInvalid) {
			t.Errorf("CleanName(%q) = %v, quero ErrNameInvalid (%s)", c.entrada, err, c.motivo)
		}
	}
}

func TestNameRuleMencionaOsLimites(t *testing.T) {
	rule := NameRule()
	for _, termo := range []string{"3", "16"} {
		if !strings.Contains(rule, termo) {
			t.Errorf("NameRule() = %q, quero mencionar %q", rule, termo)
		}
	}
}

func TestIdentityPersistent(t *testing.T) {
	casos := []struct {
		kind Kind
		quero bool
	}{
		{KindPubKey, true},
		{KindLocal, true},
		{KindGuest, false},
	}
	for _, c := range casos {
		if got := (Identity{Kind: c.kind}).Persistent(); got != c.quero {
			t.Errorf("Identity{%s}.Persistent() = %v, quero %v", c.kind, got, c.quero)
		}
	}
}
