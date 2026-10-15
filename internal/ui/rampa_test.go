package ui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

const rampaTeste = " .:-=+*#%@"

var palhetaTeste = []lipgloss.Style{
	lipgloss.NewStyle().Foreground(lipgloss.Color("#111111")),
	lipgloss.NewStyle().Foreground(lipgloss.Color("#222222")),
	lipgloss.NewStyle().Foreground(lipgloss.Color("#333333")),
}

// A cor sai do calor: um caractere mais denso na rampa cai num nível mais
// quente da paleta. É o que faz a chama ter ponta fria e brasa.
func TestPintaRampaSegueOCalor(t *testing.T) {
	casos := []struct {
		ch    string
		nivel int
	}{
		{".", 0}, // o degrau mais frio
		{":", 0},
		{"=", 1},
		{"*", 1},
		{"%", 2},
		{"@", 2}, // a brasa
	}
	for _, c := range casos {
		out := PintaRampa([]string{c.ch}, rampaTeste, palhetaTeste)[0]
		quero := palhetaTeste[c.nivel].Render(c.ch)
		if out != quero {
			t.Errorf("%q saiu no nível errado: %q, quero %q", c.ch, out, quero)
		}
	}
}

// O primeiro degrau da rampa é a ausência de fogo e fica sem tinta: é por ele
// que o fundo da tela aparece entre as labaredas.
func TestOEspacoNaoLevaTinta(t *testing.T) {
	out := PintaRampa([]string{"   "}, rampaTeste, palhetaTeste)[0]

	if out != "   " {
		t.Errorf("espaço saiu pintado: %q", out)
	}
	// E num trecho misturado, o espaço continua limpo no meio.
	mista := PintaRampa([]string{"@ @"}, rampaTeste, palhetaTeste)[0]
	if cs := celulas(mista); cs[1].estilo != "" {
		t.Error("o espaço entre duas labaredas levou tinta")
	}
}

// Pintar não pode mudar o que o terminal mostra, só como mostra.
func TestPintaRampaNaoMexeNoTexto(t *testing.T) {
	linhas := []string{"  .:-=  ", "@@%%##**", "        "}
	out := PintaRampa(linhas, rampaTeste, palhetaTeste)

	for i, l := range out {
		if got := visivel(l); got != linhas[i] {
			t.Errorf("linha %d: texto visível %q, quero %q", i, got, linhas[i])
		}
		if got := Largura(l); got != len(linhas[i]) {
			t.Errorf("linha %d mede %d colunas, quero %d", i, got, len(linhas[i]))
		}
	}
}

// Vizinhos de mesma cor saem num trecho só. Não é esmero: um escape por
// caractere multiplicaria o tamanho de cada quadro de uma chama que cobre a
// tela toda, e ela repinta vinte vezes por segundo.
func TestPintaRampaJuntaOsVizinhos(t *testing.T) {
	out := PintaRampa([]string{"@@@@@@@@@@"}, rampaTeste, palhetaTeste)[0]

	if got := strings.Count(out, "\x1b["); got != 2 {
		t.Errorf("%d escapes para dez caracteres iguais, quero 2 (abre e fecha)", got)
	}

	// Numa linha de calor variado, ainda tem de haver menos escape que
	// caractere — senão a compressão não está funcionando.
	variada := PintaRampa([]string{"..::==**%%@@"}, rampaTeste, palhetaTeste)[0]
	if got := strings.Count(variada, "\x1b["); got >= 12 {
		t.Errorf("%d escapes para 12 caracteres: a compressão não pegou", got)
	}
}

// Sem paleta, a chama sai como estava: é o que mantém o desenho utilizável num
// terminal sem cor.
func TestSemPalhetaNaoPinta(t *testing.T) {
	linhas := []string{"@@%%"}
	if out := PintaRampa(linhas, rampaTeste, nil); out[0] != linhas[0] {
		t.Errorf("pintou sem paleta: %q", out[0])
	}
}

// Caractere fora da rampa não é fogo e não leva tinta.
func TestForaDaRampaNaoLevaTinta(t *testing.T) {
	out := PintaRampa([]string{"abc"}, rampaTeste, palhetaTeste)[0]
	if out != "abc" {
		t.Errorf("pintou o que não é da rampa: %q", out)
	}
}
