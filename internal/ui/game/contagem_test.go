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

// Cada runa que dispara conta num passo seu, depois dos dados, e a que nao
// dispara nao conta passo nenhum. Sem isso o ganho dela apareceria embutido
// no total, sem dar para ver de onde veio.
func TestCadaRunaContaNumPassoSeu(t *testing.T) {
	// C soma fichas por dado 3; L soma mult por ter um 6; A pede um dado 1,
	// que esta mao nao tem.
	s, _ := tecla(comRunas("C", "A", "L"), "space")

	var rotulos []string
	for _, p := range s.anim.passos {
		rotulos = append(rotulos, p.rotulo)
	}
	quero := []string{"FULL HOUSE", "dado 3", "dado 3", "dado 3", "dado 6", "dado 6", "runa C", "runa L"}
	if got := strings.Join(rotulos, " | "); got != strings.Join(quero, " | ") {
		t.Errorf("passos da contagem:\n  %s\nquero:\n  %s", got, strings.Join(quero, " | "))
	}
}

// A runa que nao dispara nao entra, nem gasta tique.
func TestRunaQueNaoDisparaNaoVirapasso(t *testing.T) {
	s, _ := tecla(comRunas("A"), "space") // A pede um dado 1

	for _, p := range s.anim.passos {
		if strings.Contains(p.rotulo, "runa") {
			t.Errorf("a runa A virou passo (%q) sem disparar", p.rotulo)
		}
	}
}

// O item que causou o passo acende, um por tique, na ordem.
func TestOItemQueContaAcende(t *testing.T) {
	s, _ := tecla(comRunas("C", "L"), "space")

	quero := []struct {
		de     origem
		indice int
	}{
		{origemCombinacao, 0}, // a combinacao nao aponta item na mesa
		{origemDado, 0}, {origemDado, 1}, {origemDado, 2}, {origemDado, 3}, {origemDado, 4},
		{origemRuna, 0}, {origemRuna, 1},
	}
	for i, q := range quero {
		s, _ = aplica(s, tiquePontuacao{})
		de, indice := s.acesa()
		if de != q.de || (q.de != origemCombinacao && indice != q.indice) {
			t.Errorf("tique %d: acende (origem=%d indice=%d), quero (origem=%d indice=%d)",
				i+1, de, indice, q.de, q.indice)
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
	de, i := s.acesa()
	if de != origemRuna || i != 0 {
		t.Errorf("no ultimo passo acende (origem=%d indice=%d), quero a runa 0", de, i)
	}
	if !strings.Contains(s.View(), marca(s.theme.Runa)) {
		t.Error("a runa do ultimo passo nao esta pintada na tela")
	}

	// E so entao a contagem fecha.
	s, _ = aplica(s, fechaContagem{})
	if s.Phase() != PhaseRodada {
		t.Errorf("fase = %v depois de fechar, quero PhaseRodada", s.Phase())
	}
}

// contaPassos roda os tiques dos passos, sem o tique que fecha.
func contaPassos(s Screen) Screen {
	for i := 0; i < 30 && !s.anim.acabou(); i++ {
		s, _ = aplica(s, tiquePontuacao{})
	}
	return s
}

// Um item aceso de cada vez: dois acesos nao diriam de onde vieram os pontos.
func TestSoUmItemAcendePorTique(t *testing.T) {
	s, _ := tecla(comRunas("C", "L"), "space")
	azul, roxo := marca(s.theme.DadoContando), marca(s.theme.Runa)

	for n := 0; !s.anim.acabou(); n++ {
		s, _ = aplica(s, tiquePontuacao{})
		v := s.View()

		if a, r := strings.Count(v, azul), strings.Count(v, roxo); a > 0 && r > 0 {
			t.Errorf("tique %d: dado e runa acesos ao mesmo tempo", n+1)
		}
		// O dado aceso e um so: o cubo (5 linhas), o traco e o rotulo.
		if a := strings.Count(v, azul); a != 0 && a != 7 {
			t.Errorf("tique %d: %d trechos azuis, quero 0 ou 7 (um dado inteiro)", n+1, a)
		}
	}
}

// Fora da contagem nada acende: a mesa parada nao tem item somando.
func TestForaDaContagemNadaAcende(t *testing.T) {
	s := comRunas("C")
	if de, i := s.acesa(); de != origemCombinacao || i != -1 {
		t.Errorf("na rodada acende (origem=%d indice=%d), quero nada", de, i)
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

	if de, i := s.acesa(); de != origemDado || i != 0 {
		t.Fatalf("acende (origem=%d indice=%d), quero o dado 0", de, i)
	}
	primeiro := strings.Split(s.mesa(118, 31), "\n")[17]
	if !strings.Contains(primeiro, marca(s.theme.DadoContando)) {
		t.Error("o dado que conta nao esta azul")
	}
	if strings.Contains(primeiro, marca(s.theme.DadoMarcado)) {
		t.Error("o vermelho do reroll venceu o azul da contagem")
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

// Nao basta haver roxo na tela: tem de ser a caixa da runa que esta contando.
// Sem esta checagem, acender a vizinha passaria batido.
func TestAcendeACaixaDaRunaCerta(t *testing.T) {
	s, _ := tecla(comRunas("C", "L"), "space")
	roxo := marca(s.theme.Runa)

	// Passa a combinacao e os cinco dados; sobram os dois passos de runa.
	for i := 0; i < 6; i++ {
		s, _ = aplica(s, tiquePontuacao{})
	}

	for _, esperada := range []string{"C", "L"} {
		s, _ = aplica(s, tiquePontuacao{})
		v := s.View()

		if !strings.Contains(v, roxo+"| "+esperada+" |") {
			t.Errorf("contando a runa %s, a caixa dela nao esta roxa", esperada)
		}
		for _, outra := range []string{"C", "L", " "} {
			if outra == esperada {
				continue
			}
			if strings.Contains(v, roxo+"| "+outra+" |") {
				t.Errorf("contando a runa %s, a caixa de %q acendeu no lugar", esperada, outra)
			}
		}
	}
}

// E o dado aceso tem de ser o da posicao certa, nao um vizinho.
func TestAcendeODadoCerto(t *testing.T) {
	s, _ := tecla(comRunas(), "space")
	azul := marca(s.theme.DadoContando)
	s, _ = aplica(s, tiquePontuacao{}) // combinacao

	for i := range s.partida.Dados {
		s, _ = aplica(s, tiquePontuacao{})
		rotulos := strings.Split(s.mesa(118, 31), "\n")[23]

		if !strings.Contains(rotulos, azul+"   ["+itoa(i+1)+"]   ") {
			t.Errorf("contando o dado %d, o rotulo dele nao esta azul:\n%q", i+1, rotulos)
		}
	}
}

func itoa(n int) string { return string(rune('0' + n)) }
