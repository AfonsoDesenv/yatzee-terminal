package game

import (
	"strings"
	"testing"

	"github.com/afonsodesenv/yatzee-terminal/internal/regras"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

// comRunas monta uma mao conhecida com as runas pedidas, pronta para enviar.
func comRunas(letras ...string) Screen {
	s := New()
	s.largura, s.altura = 118, 32
	s.partida.Alvo = 1_000_000 // longe da meta: nada de fogo no meio do caminho
	for _, l := range letras {
		r, _ := regras.RunaPorLetra(l)
		s.partida = s.partida.Compra(r, 0)
	}
	s.partida.Dados = regras.Mao{3, 3, 3, 6, 6} // full house
	return s
}

// contaPassos roda os tiques dos passos, sem o tique que fecha.
func contaPassos(s Screen) Screen {
	for i := 0; i < 40 && !s.anim.acabou(); i++ {
		s, _ = aplica(s, tiquePontuacao{})
	}
	return s
}

// A runa conta um tique por dado que a dispara, e nao um com a soma de todos.
// Com tres dados de 3, a runa C paga tres vezes.
func TestARunaContaUmTiquePorDado(t *testing.T) {
	s, _ := tecla(comRunas("C", "A", "L"), "space")

	var rotulos []string
	for _, p := range s.anim.passos {
		rotulos = append(rotulos, p.rotulo)
	}
	quero := []string{
		"FULL HOUSE",
		"dado 3", "dado 3", "dado 3", "dado 6", "dado 6",
		// C soma 10 por dado 3: tres dados, tres tiques.
		"runa C dado 3", "runa C dado 3", "runa C dado 3",
		// L soma 2 ao mult por dado 6: dois dados, dois tiques. A runa A
		// pede um dado 1 e nao entra.
		"runa L dado 6", "runa L dado 6",
	}
	if got := strings.Join(rotulos, " | "); got != strings.Join(quero, " | ") {
		t.Errorf("passos da contagem:\n  %s\nquero:\n  %s", got, strings.Join(quero, " | "))
	}
}

// Cada tique da runa paga a parcela dela, nao o total: com +2 por dado 6, cada
// tique soma 2.
func TestCadaTiqueDaRunaPagaUmaParcela(t *testing.T) {
	s, _ := tecla(comRunas("L"), "space")

	var somas []int
	for _, p := range s.anim.passos {
		if strings.HasPrefix(p.rotulo, "runa") {
			somas = append(somas, p.mult)
		}
	}
	if len(somas) != 2 {
		t.Fatalf("%d tiques de runa, quero 2 (um por dado 6)", len(somas))
	}
	for i, m := range somas {
		if m != regras.MultDaRuna {
			t.Errorf("tique %d da runa somou %d ao mult, quero %d", i+1, m, regras.MultDaRuna)
		}
	}
}

// O invariante que o fatiamento nao pode furar: a soma dos passos tem de dar
// exatamente a pontuacao da mao. Parcela errada por gatilho passaria batida
// na tela e so apareceria no placar.
func TestASomaDosPassosFechaComAPontuacao(t *testing.T) {
	maos := []regras.Mao{
		{3, 3, 3, 6, 6},
		{6, 6, 6, 6, 6},
		{1, 2, 3, 4, 5},
		{2, 4, 4, 5, 1},
	}
	letras := [][]string{nil, {"C"}, {"L"}, {"C", "L"}, {"F", "L"}, {"A", "B", "C", "D", "E"}}

	for _, m := range maos {
		for _, ls := range letras {
			s := comRunas(ls...)
			s.partida.Dados = m
			s, _ = tecla(s, "space")
			s = contaPassos(s)

			quero := regras.AvaliaCom(m, s.partida.Runas).Total()
			if got := s.anim.fichas * s.anim.mult; got != quero {
				t.Errorf("mao %v com %v: a contagem deu %d (%d x %d), a pontuacao e %d",
					m, ls, got, s.anim.fichas, s.anim.mult, quero)
			}
		}
	}
}

// Num tique de runa disparada por dado, os dois acendem: a runa que somou e o
// dado que a fez somar.
func TestOTiqueDaRunaAcendeARunaEODado(t *testing.T) {
	s, _ := tecla(comRunas("L"), "space")

	// Passa a combinacao e os cinco dados.
	for i := 0; i < 6; i++ {
		s, _ = aplica(s, tiquePontuacao{})
	}

	// Os dois tiques da runa L, nos dados 6 (posicoes 3 e 4 da mao).
	for _, quero := range []int{3, 4} {
		s, _ = aplica(s, tiquePontuacao{})

		dado, runa := s.acende()
		if dado != quero {
			t.Errorf("o tique acendeu o dado %d, quero o %d", dado, quero)
		}
		if runa != 0 {
			t.Errorf("o tique acendeu a runa %d, quero a 0", runa)
		}

		v := s.View()
		if !strings.Contains(v, marca(s.theme.Runa)) {
			t.Error("a runa nao esta roxa no tique dela")
		}
		if !strings.Contains(v, marca(s.theme.DadoContando)) {
			t.Error("o dado que disparou a runa nao esta azul")
		}
	}
}

// E e o dado certo: o que disparou, nao um vizinho.
func TestARunaAcendeODadoQueADisparou(t *testing.T) {
	s, _ := tecla(comRunas("C"), "space") // C dispara nos dados 3, posicoes 0,1,2
	azul := marca(s.theme.DadoContando)

	for i := 0; i < 6; i++ {
		s, _ = aplica(s, tiquePontuacao{})
	}
	for _, quero := range []int{0, 1, 2} {
		s, _ = aplica(s, tiquePontuacao{})

		rotulos := strings.Split(s.mesa(118, 31), "\n")[23]
		if !strings.Contains(rotulos, azul+"   ["+itoa(quero+1)+"]   ") {
			t.Errorf("o tique da runa devia acender o dado %d:\n%q", quero+1, rotulos)
		}
	}
}

func itoa(n int) string { return string(rune('0' + n)) }

// O passo da combinacao nao aponta item nenhum: nao e dado nem runa.
func TestOPassoDaCombinacaoNaoAcendeNada(t *testing.T) {
	s, _ := tecla(comRunas("C"), "space")
	s, _ = aplica(s, tiquePontuacao{})

	if dado, runa := s.acende(); dado != semItem || runa != semItem {
		t.Errorf("a combinacao acendeu dado=%d runa=%d, quero nada", dado, runa)
	}
}

// O passo de dado acende o dado e nenhuma runa.
func TestOPassoDeDadoNaoAcendeRuna(t *testing.T) {
	s, _ := tecla(comRunas("C"), "space")
	s, _ = aplica(s, tiquePontuacao{}) // combinacao

	for i := range s.partida.Dados {
		s, _ = aplica(s, tiquePontuacao{})
		dado, runa := s.acende()
		if dado != i {
			t.Errorf("tique do dado %d acendeu o dado %d", i, dado)
		}
		if runa != semItem {
			t.Errorf("tique do dado %d acendeu a runa %d, quero nenhuma", i, runa)
		}
	}
}

// Era o bug da primeira tentativa: no ultimo passo a fase saia da contagem no
// mesmo tique, e como as runas contam no fim, a runa nunca acendia.
func TestOUltimoPassoFicaAcesoUmTique(t *testing.T) {
	s, _ := tecla(comRunas("L"), "space")
	s = contaPassos(s)

	if s.Phase() != PhaseContando {
		t.Fatalf("fase = %v logo depois do ultimo passo, quero continuar contando", s.Phase())
	}
	if _, runa := s.acende(); runa != 0 {
		t.Errorf("no ultimo passo a runa acesa e %d, quero a 0", runa)
	}
	if !strings.Contains(s.View(), marca(s.theme.Runa)) {
		t.Error("a runa do ultimo passo nao esta pintada na tela")
	}

	s, _ = aplica(s, fechaContagem{})
	if s.Phase() != PhaseRodada {
		t.Errorf("fase = %v depois de fechar, quero PhaseRodada", s.Phase())
	}
}

// Um dado aceso de cada vez, sempre.
func TestSoUmDadoAcendePorTique(t *testing.T) {
	s, _ := tecla(comRunas("C", "L"), "space")
	azul := marca(s.theme.DadoContando)

	for n := 0; !s.anim.acabou(); n++ {
		s, _ = aplica(s, tiquePontuacao{})
		// A cor abre uma vez por linha do dado: o cubo (5), o traco e o
		// rotulo.
		if a := strings.Count(s.View(), azul); a != 0 && a != 7 {
			t.Errorf("tique %d: %d trechos azuis, quero 0 ou 7 (um dado inteiro)", n+1, a)
		}
	}
}

// Fora da contagem nada acende: a mesa parada nao tem item somando.
func TestForaDaContagemNadaAcende(t *testing.T) {
	s := comRunas("C")
	if dado, runa := s.acende(); dado != semItem || runa != semItem {
		t.Errorf("na rodada acende dado=%d runa=%d, quero nada", dado, runa)
	}
	if v := s.View(); strings.Contains(v, marca(s.theme.DadoContando)) {
		t.Error("a mesa parada tem dado aceso")
	}
}

// Dado marcado pro reroll e contando ao mesmo tempo: o azul vence, porque no
// meio da contagem o reroll ja nao vai acontecer.
func TestOAzulDaContagemVenceOVermelhoDoReroll(t *testing.T) {
	s := comRunas()
	s.partida = s.partida.Marca(1) // o primeiro dado, que conta no 2o tique
	s, _ = tecla(s, "space")

	s, _ = aplica(s, tiquePontuacao{}) // combinacao
	s, _ = aplica(s, tiquePontuacao{}) // dado 1

	if dado, _ := s.acende(); dado != 0 {
		t.Fatalf("acende o dado %d, quero o 0", dado)
	}
	primeiro := strings.Split(s.mesa(118, 31), "\n")[17]
	if !strings.Contains(primeiro, marca(s.theme.DadoContando)) {
		t.Error("o dado que conta nao esta azul")
	}
	if strings.Contains(primeiro, marca(s.theme.DadoMarcado)) {
		t.Error("o vermelho do reroll venceu o azul da contagem")
	}
}

// Runa que nao dispara nao vira passo, nem gasta tique.
func TestRunaQueNaoDisparaNaoViraPasso(t *testing.T) {
	s, _ := tecla(comRunas("A"), "space") // A pede um dado 1

	for _, p := range s.anim.passos {
		if strings.Contains(p.rotulo, "runa") {
			t.Errorf("a runa A virou passo (%q) sem disparar", p.rotulo)
		}
	}
}

// E acender nao pode mover nada de lugar.
func TestAcenderNaoMexeNoLayout(t *testing.T) {
	s, _ := tecla(comRunas("C", "L"), "space")

	for n := 0; !s.anim.acabou(); n++ {
		s, _ = aplica(s, tiquePontuacao{})
		for i, l := range strings.Split(s.mesa(118, 31), "\n") {
			if got := ui.Largura(l); got != 118 {
				t.Fatalf("tique %d, linha %d: %d colunas, a tela tem 118", n+1, i, got)
			}
		}
	}
}
