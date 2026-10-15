package game

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/regras"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

// marca devolve o escape que um estilo põe na frente do texto. Procura-se o
// prefixo, e não a linha inteira, porque cada linha é embalada por conta
// própria e o teste não deve depender de onde a cor abre.
func marca(e lipgloss.Style) string {
	pintado := e.Render("x")
	return pintado[:strings.Index(pintado, "x")]
}

func mesaCom(marcados ...int) Screen {
	s := New()
	s.largura, s.altura = 118, 32
	s.partida.Dados = regras.Mao{3, 1, 4, 6, 2}
	for _, i := range marcados {
		s.partida = s.partida.Marca(i)
	}
	return s
}

// O dado escolhido para o reroll sai de vermelho; o que não foi escolhido
// fica como estava.
func TestDadoMarcadoSaiVermelho(t *testing.T) {
	cor := marca(mesaCom().theme.DadoMarcado)

	if v := mesaCom().View(); strings.Contains(v, cor) {
		t.Error("a mesa já tem vermelho sem nenhum dado marcado")
	}
	if v := mesaCom(2).View(); !strings.Contains(v, cor) {
		t.Error("o dado marcado não saiu vermelho")
	}
}

// A cor acompanha o cubo, o traço e o rótulo: os três são um sinal só.
func TestOVermelhoPegaCuboTracoERotulo(t *testing.T) {
	s := mesaCom(2)
	cor := marca(s.theme.DadoMarcado)

	var comCubo, comTraco, comRotulo bool
	for _, l := range strings.Split(s.View(), "\n") {
		if !strings.Contains(l, cor) {
			continue
		}
		switch {
		case strings.Contains(l, ">2<"):
			comRotulo = true
		case strings.Contains(l, "^^^"):
			comTraco = true
		case strings.Contains(l, "---") || strings.Contains(l, "|"):
			comCubo = true
		}
	}
	if !comCubo {
		t.Error("o cubo do dado marcado não está pintado")
	}
	if !comTraco {
		t.Error("o traço do dado marcado não está pintado")
	}
	if !comRotulo {
		t.Error("o rótulo do dado marcado não está pintado")
	}
}

// O ponto de fazer a composição em células: a cor não pode mover nada. Era o
// risco real, porque os bytes do escape ocupariam coluna num []rune.
func TestACorDoDadoNaoMexeNoLayout(t *testing.T) {
	limpa := strings.Split(mesaCom().mesa(118, 31), "\n")
	pintada := strings.Split(mesaCom(1, 3, 5).mesa(118, 31), "\n")

	if len(limpa) != len(pintada) {
		t.Fatalf("a cor mudou a altura: %d linhas viram %d", len(limpa), len(pintada))
	}
	for i := range limpa {
		if a, b := ui.Largura(limpa[i]), ui.Largura(pintada[i]); a != b {
			t.Errorf("linha %d: %d colunas limpa, %d pintada", i, a, b)
		}
		if ui.Largura(pintada[i]) != 118 {
			t.Errorf("linha %d mede %d colunas, a tela tem 118", i, ui.Largura(pintada[i]))
		}
	}
}

// Na loja a runa em foco vai de roxo e o botão de verde, cada cor só no seu
// item: se as duas aparecessem juntas, o foco não diria onde está.
func TestAsCoresDaLojaSeguemOFoco(t *testing.T) {
	s := naLoja(t, 50)
	roxo := marca(s.theme.RunaEmFoco)
	verde := marca(s.theme.BotaoEmFoco)

	s.sel = focoPrimeiraRuna
	naRuna := s.View()
	if !strings.Contains(naRuna, roxo) {
		t.Error("a runa em foco não saiu roxa")
	}
	if strings.Contains(naRuna, verde) {
		t.Error("o botão está verde com o foco numa runa")
	}

	s.sel = focoContinuar
	noBotao := s.View()
	if !strings.Contains(noBotao, verde) {
		t.Error("o botão em foco não saiu verde")
	}
	if strings.Contains(noBotao, roxo) {
		t.Error("alguma runa está roxa com o foco no botão")
	}
}

// Uma runa só de cada vez: duas pintadas deixariam a escolha ambígua.
func TestSoUmaRunaFicaRoxa(t *testing.T) {
	s := naLoja(t, 50)
	roxo := marca(s.theme.RunaEmFoco)

	for i := range s.vitrine {
		s.sel = focoPrimeiraRuna + i

		for j := range s.vitrine {
			pintada := strings.Contains(strings.Join(s.carta(j), ""), roxo)
			if i == j && !pintada {
				t.Errorf("foco na runa %d: a carta dela não está roxa", i)
			}
			if i != j && pintada {
				t.Errorf("foco na runa %d: a carta %d ficou roxa também", i, j)
			}
		}
	}
}

// A cor da loja também não pode empurrar nada de lugar.
func TestACorDaLojaNaoMexeNoLayout(t *testing.T) {
	s := naLoja(t, 50)

	s.sel = -1 // fora do anel: nada em foco, nada pintado
	limpa := strings.Split(s.loja(118, 31), "\n")

	for _, sel := range []int{focoContinuar, focoPrimeiraRuna, focoPrimeiraRuna + 2} {
		s.sel = sel
		pintada := strings.Split(s.loja(118, 31), "\n")
		if len(limpa) != len(pintada) {
			t.Fatalf("foco %d mudou a altura da loja: %d para %d", sel, len(limpa), len(pintada))
		}
		for i := range limpa {
			if a, b := ui.Largura(limpa[i]), ui.Largura(pintada[i]); a != b {
				t.Errorf("foco %d, linha %d: %d colunas limpa, %d pintada", sel, i, a, b)
			}
		}
	}
}

// A chama sai laranja, no painel e na tela cheia — dois caminhos de desenho
// diferentes (Compoe num, SobreOpaco no outro), então os dois são checados.
func TestAChamaSaiPintada(t *testing.T) {
	s := contando(t, 1)
	s, _ = aplica(s, tiqueFogo{})
	if s.chama == nil {
		t.Fatal("sem chama no painel")
	}

	if !temCorDeFogo(s, s.View()) {
		t.Error("a chama do painel saiu sem cor nenhuma")
	}

	v := New()
	v.largura, v.altura = 118, 32
	v = venceAgora(v)
	v, _ = aplica(v, tiqueFogo{})
	if !temCorDeFogo(v, v.View()) {
		t.Error("a chama da tela de vitória saiu sem cor nenhuma")
	}
}

// A paleta vai do frio ao quente, e cores diferentes têm de aparecer juntas na
// mesma chama: uma laranja chapada não teria ponta nem brasa.
func TestAChamaTemGradiente(t *testing.T) {
	s := contando(t, 1)
	for i := 0; i < 16; i++ {
		s, _ = aplica(s, tiqueFogo{})
	}
	tela := s.View()

	var niveis int
	for _, e := range s.theme.Fogo {
		if strings.Contains(tela, marca(e)) {
			niveis++
		}
	}
	if niveis < 3 {
		t.Errorf("a chama usa só %d níveis da paleta, quero pelo menos 3", niveis)
	}
}

// E a cor da chama também não pode mover nada: ela passa por cima do painel.
func TestACorDaChamaNaoMexeNoLayout(t *testing.T) {
	s := contando(t, 1)
	for i := 0; i < 8; i++ {
		s, _ = aplica(s, tiqueFogo{})
	}

	for i, l := range strings.Split(s.mesa(118, 31), "\n") {
		if got := ui.Largura(l); got != 118 {
			t.Errorf("linha %d mede %d colunas com a chama pintada, a tela tem 118", i, got)
		}
	}
}

// temCorDeFogo diz se algum nível da paleta da chama aparece na tela.
func temCorDeFogo(s Screen, tela string) bool {
	for _, e := range s.theme.Fogo {
		if strings.Contains(tela, marca(e)) {
			return true
		}
	}
	return false
}
