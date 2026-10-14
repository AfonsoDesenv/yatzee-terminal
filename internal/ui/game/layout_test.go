package game

import (
	"strings"
	"testing"

	"github.com/afonsodesenv/yatzee-terminal/internal/regras"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

// contando devolve a tela no meio da contagem, com e sem o painel em chamas.
func contando(t *testing.T, alvo int64) Screen {
	t.Helper()
	s := New()
	s.largura, s.altura = 118, 32
	s.partida.Alvo = alvo
	s.partida.Dados = regras.Mao{3, 3, 3, 6, 6}
	s, _ = tecla(s, "enter")
	s, _ = aplica(s, tiquePontuacao{})
	if s.Phase() != PhaseContando {
		t.Fatalf("fase = %v, quero PhaseContando", s.Phase())
	}
	return s
}

// O bug: a chama é mais larga que o painel da meta, e enquanto ela substituía
// o bloco do painel empurrava o resto da linha para a direita a cada quadro.
// Os painéis têm de ficar na mesma coluna com o fogo aceso ou apagado.
func TestOFogoNoPainelNaoMexeNoResto(t *testing.T) {
	semFogo := contando(t, 1_000_000) // longe da meta: nada pega fogo
	comFogo := contando(t, 1)         // a meta cai no primeiro passo

	if semFogo.chama != nil {
		t.Fatal("pegou fogo sem cruzar a meta")
	}
	if comFogo.chama == nil {
		t.Fatal("cruzou a meta e não pegou fogo")
	}

	frio := strings.Split(semFogo.mesa(118, 31), "\n")
	quente := strings.Split(comFogo.mesa(118, 31), "\n")

	if len(frio) != len(quente) {
		t.Fatalf("o fogo mudou a altura da tela: %d linhas viram %d", len(frio), len(quente))
	}
	for i := range frio {
		if a, b := len([]rune(frio[i])), len([]rune(quente[i])); a != b {
			t.Errorf("linha %d: largura %d sem fogo, %d com fogo", i, a, b)
		}
	}

	// As colunas dos painéis que não queimam têm de ser as mesmas.
	for _, marca := range []string{"RUNAS", "NIVEL", "REROLLS"} {
		if a, b := coluna(frio, marca), coluna(quente, marca); a != b {
			t.Errorf("%s na coluna %d sem fogo e %d com fogo", marca, a, b)
		}
	}
}

// coluna acha em que coluna um texto aparece na tela.
func coluna(tela []string, texto string) int {
	for _, l := range tela {
		if i := strings.Index(l, texto); i >= 0 {
			return i
		}
	}
	return -1
}

// A chama transborda o painel, e e a margem invisivel que a segura: ela nao
// pode passar da borda da tela nem comer o painel vizinho.
func TestAChamaCabeDentroDaMargem(t *testing.T) {
	s := contando(t, 1)
	tela := strings.Split(s.mesa(118, 31), "\n")

	for i, l := range tela {
		if n := len([]rune(l)); n != 118 {
			t.Fatalf("linha %d tem %d colunas, a tela tem 118", i, n)
		}
	}

	xRunas := ui.Margem + larguraEsquerda + 2 + espacoPainel
	if c := coluna(tela, "| RUNAS"); c != xRunas {
		t.Errorf("o painel das runas esta na coluna %d, esperava %d", c, xRunas)
	}
}
