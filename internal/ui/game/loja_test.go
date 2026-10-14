package game

import (
	"strings"
	"testing"

	"github.com/afonsodesenv/yatzee-terminal/internal/regras"
)

// naLoja leva a tela até a loja, com saldo suficiente para comprar.
func naLoja(t *testing.T, moedas int) Screen {
	t.Helper()
	s := venceAgora(New())
	s.partida.Moedas = moedas
	s, _ = tecla(s, "enter") // vitória -> loja
	if s.Phase() != PhaseLoja {
		t.Fatalf("fase = %v, quero PhaseLoja", s.Phase())
	}
	return s
}

func TestALojaOfereceTresRunasDistintas(t *testing.T) {
	s := naLoja(t, 50)

	if got := len(s.vitrine); got != 3 {
		t.Fatalf("a vitrine tem %d runas, quero 3", got)
	}
	vistas := map[string]bool{}
	for _, r := range s.vitrine {
		if r.Letra == "" {
			t.Error("vitrine com espaço vazio logo ao abrir")
		}
		if vistas[r.Letra] {
			t.Errorf("a runa %s apareceu duas vezes na mesma vitrine", r.Letra)
		}
		vistas[r.Letra] = true
	}
}

// A vitrine tem de dizer o que cada runa faz: só a letra deixaria a compra no
// chute.
func TestALojaExplicaCadaRuna(t *testing.T) {
	s := naLoja(t, 50)
	v := s.View()
	for _, r := range s.vitrine {
		if !strings.Contains(v, r.Descricao()) {
			t.Errorf("a loja não explica a runa %s (%q)", r.Letra, r.Descricao())
		}
	}
}

func TestComprarLevaARunaEDescontaOSaldo(t *testing.T) {
	s := naLoja(t, 50)
	escolhida := s.vitrine[0]

	s, _ = tecla(s, "1")

	if got := len(s.Partida().Runas); got != 1 {
		t.Fatalf("o jogador tem %d runas, quero 1", got)
	}
	if got := s.Partida().Runas[0].Letra; got != escolhida.Letra {
		t.Errorf("comprou a runa %s, esperava %s", got, escolhida.Letra)
	}
	if got := s.Partida().Moedas; got != 50-PrecoRuna {
		t.Errorf("saldo = %d, quero %d", got, 50-PrecoRuna)
	}
}

// Comprada, some da prateleira: senão daria para levar a mesma runa duas
// vezes com um toque repetido.
func TestRunaCompradaSaiDaVitrine(t *testing.T) {
	s := naLoja(t, 50)
	s, _ = tecla(s, "1")
	if s.vitrine[0].Letra != "" {
		t.Error("a runa continua à venda depois de comprada")
	}

	s, _ = tecla(s, "1")
	if got := len(s.Partida().Runas); got != 1 {
		t.Errorf("comprou de novo no mesmo slot: %d runas", got)
	}
}

func TestSemSaldoNaoCompra(t *testing.T) {
	s := naLoja(t, PrecoRuna-1)
	s, _ = tecla(s, "1")

	if got := len(s.Partida().Runas); got != 0 {
		t.Errorf("comprou sem saldo: %d runas", got)
	}
	if got := s.Partida().Moedas; got != PrecoRuna-1 {
		t.Errorf("saldo = %d, não devia mudar", got)
	}
}

// Cabem cinco runas; a sexta não entra nem com dinheiro sobrando.
func TestNaoPassaDoLimiteDeRunas(t *testing.T) {
	s := naLoja(t, 500)
	for i := 0; i < regras.MaxRunas; i++ {
		s.partida = s.partida.Compra(regras.Catalogo[i], 0)
	}

	s, _ = tecla(s, "1")
	if got := len(s.Partida().Runas); got != regras.MaxRunas {
		t.Errorf("passou do limite: %d runas, máximo %d", got, regras.MaxRunas)
	}
}

// A runa comprada participa da jogada seguinte.
func TestARunaCompradaEntraNaContagem(t *testing.T) {
	s := naLoja(t, 50)
	c, _ := regras.RunaPorLetra("C")
	s.partida = s.partida.Compra(c, 0)

	s, _ = tecla(s, "enter") // loja -> próximo nível
	s.partida.Dados = regras.Mao{3, 3, 3, 6, 6}
	s, _ = tecla(s, "enter") // envia

	var achou bool
	for _, p := range s.anim.passos {
		if strings.Contains(p.rotulo, "runa C") {
			achou = true
		}
	}
	if !achou {
		t.Error("a runa comprada não virou passo da contagem")
	}
}
