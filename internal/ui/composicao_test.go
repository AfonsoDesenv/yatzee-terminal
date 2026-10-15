package ui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

var vermelho = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))

// visivel é o que o terminal mostra, sem a marcação.
func visivel(s string) string {
	var b strings.Builder
	for _, c := range celulas(s) {
		b.WriteRune(c.ch)
	}
	return b.String()
}

// Uma linha pintada de ponta a ponta traz o escape uma vez, no começo: o
// estilo tem de valer para todos os caracteres até o reset, e não só para o
// primeiro.
func TestCelulasCarregamOEstiloAteOReset(t *testing.T) {
	cs := celulas("\x1b[31mabc\x1b[0md")

	if len(cs) != 4 {
		t.Fatalf("%d células, quero 4 (os escapes não ocupam coluna)", len(cs))
	}
	for i, c := range cs[:3] {
		if c.estilo != "\x1b[31m" {
			t.Errorf("célula %d (%q) ficou sem estilo", i, string(c.ch))
		}
	}
	if cs[3].estilo != "" {
		t.Errorf("o reset não limpou o estilo: %q continua pintado", string(cs[3].ch))
	}
	if got := visivel("\x1b[31mabc\x1b[0md"); got != "abcd" {
		t.Errorf("texto visível = %q, quero \"abcd\"", got)
	}
}

// O bug que isso evita: contando por rune, os bytes de um escape ocupariam
// coluna e empurrariam todo o resto da linha para a direita.
func TestCompoeNaoDeslocaOQueEstaPintado(t *testing.T) {
	const largura = 20

	semCor := Compoe(Tela(largura, 1), []string{"abc"}, 5, 0, false)
	comCor := Compoe(Tela(largura, 1), []string{vermelho.Render("abc")}, 5, 0, false)

	if a, b := visivel(semCor[0]), visivel(comCor[0]); a != b {
		t.Errorf("a cor mudou o desenho:\nsem: %q\ncom: %q", a, b)
	}
	if got := Largura(comCor[0]); got != largura {
		t.Errorf("a linha pintada mede %d colunas, a tela tem %d", got, largura)
	}
	if !strings.Contains(comCor[0], "\x1b[") {
		t.Error("a cor do bloco não sobreviveu à composição")
	}
}

// Sem fechar a marcação, a cor pintaria o resto da tela a partir dali.
func TestACorNaoVazaParaOResto(t *testing.T) {
	linha := Compoe(Tela(20, 1), []string{vermelho.Render("ab")}, 5, 0, false)[0]

	// O reset vem logo depois do trecho pintado, não no fim da linha: o que
	// sobra à direita é espaço limpo e não precisa de marcação nenhuma.
	if !strings.Contains(linha, reset) {
		t.Errorf("a linha não fecha a marcação: %q", linha)
	}
	// O que vem depois do trecho pintado tem de estar limpo.
	for i, c := range celulas(linha) {
		if (i < 5 || i >= 7) && c.estilo != "" {
			t.Errorf("a coluna %d herdou a cor do vizinho", i)
		}
	}
}

// Compor por cima de uma célula pintada tira a cor dela junto: a tinta
// pertence à célula, não à posição na tela.
func TestComporPorCimaApagaACor(t *testing.T) {
	tela := Compoe(Tela(10, 1), []string{vermelho.Render("abc")}, 0, 0, false)
	tela = Compoe(tela, []string{"x"}, 1, 0, false)

	cs := celulas(tela[0])
	if cs[1].ch != 'x' {
		t.Fatalf("a coluna 1 tem %q, quero 'x'", string(cs[1].ch))
	}
	if cs[1].estilo != "" {
		t.Errorf("o 'x' nasceu pintado com a cor de quem estava embaixo")
	}
	if cs[0].estilo == "" || cs[2].estilo == "" {
		t.Error("os vizinhos perderam a cor que era deles")
	}
}

// Um espaço pintado continua sendo espaço: a transparência olha o caractere,
// senão um fundo colorido taparia o fogo atrás do painel.
func TestTransparenciaIgnoraEspacoPintado(t *testing.T) {
	tela := Compoe(Tela(5, 1), []string{"#####"}, 0, 0, false)
	tela = Compoe(tela, []string{vermelho.Render("a a")}, 0, 0, true)

	if got := visivel(tela[0]); got != "a#a##" {
		t.Errorf("sobreposição transparente = %q, quero \"a#a##\"", got)
	}
}

// A moldura corta o que vaza. Em células, para não partir um escape no meio e
// despejar os bytes dele na tela.
func TestMolduraCortaSemPartirOEscape(t *testing.T) {
	largo := vermelho.Render(strings.Repeat("x", 40))
	out := Moldura(largo, 12, 3)

	for _, l := range strings.Split(out, "\n") {
		if got := Largura(l); got != 12 {
			t.Errorf("linha %q mede %d colunas, quero 12", visivel(l), got)
		}
	}
	if strings.Contains(visivel(out), "[31m") || strings.Contains(visivel(out), "[0") {
		t.Error("o corte partiu um escape e os bytes vazaram para a tela")
	}
}
