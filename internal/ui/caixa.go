package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// Separador, no lugar de uma linha de painel, vira um traço de ponta a ponta.
const Separador = "\x00sep"

// Largura mede em colunas de terminal, não em bytes: acento e caractere largo
// ocupam o que ocupam na tela, e é por isso que len() não serve aqui.
func Largura(s string) int { return lipgloss.Width(s) }

// Centraliza põe s no meio de uma faixa, cortando se não couber.
func Centraliza(s string, largura int) string {
	l := Largura(s)
	if l >= largura {
		r := []rune(s)
		if len(r) > largura {
			return string(r[:largura])
		}
		return s
	}
	esq := (largura - l) / 2
	return strings.Repeat(" ", esq) + s + strings.Repeat(" ", largura-l-esq)
}

// Painel desenha uma caixa com as linhas dentro, centralizadas. O vocabulário
// de cantos (.---. e '---') é o mesmo em toda a interface, para as telas
// parecerem de uma peça só.
func Painel(largura int, linhas ...string) []string {
	out := []string{"." + strings.Repeat("-", largura) + "."}
	for _, l := range linhas {
		if l == Separador {
			out = append(out, "|"+strings.Repeat("-", largura)+"|")
			continue
		}
		out = append(out, "|"+Centraliza(l, largura)+"|")
	}
	return append(out, "'"+strings.Repeat("-", largura)+"'")
}

// Rotulado formata "rótulo    valor" numa linha de painel, com o valor colado
// à direita. Serve aos pares ALVO/VOCE, MAOS/REROLLS e afins.
func Rotulado(rotulo, valor string, largura int) string {
	meio := largura - Largura(rotulo) - Largura(valor) - 2
	if meio < 1 {
		meio = 1
	}
	return " " + rotulo + strings.Repeat(" ", meio) + valor + " "
}

// LadoALado junta blocos horizontalmente, alinhados pelo topo e separados por
// espaco colunas. Blocos de alturas diferentes são completados com vazio.
func LadoALado(espaco int, blocos ...[]string) []string {
	altura := 0
	larguras := make([]int, len(blocos))
	for i, b := range blocos {
		if len(b) > altura {
			altura = len(b)
		}
		for _, l := range b {
			if w := Largura(l); w > larguras[i] {
				larguras[i] = w
			}
		}
	}

	out := make([]string, altura)
	for y := 0; y < altura; y++ {
		var sb strings.Builder
		for i, b := range blocos {
			s := ""
			if y < len(b) {
				s = b[y]
			}
			sb.WriteString(s + strings.Repeat(" ", larguras[i]-Largura(s)))
			if i < len(blocos)-1 {
				sb.WriteString(strings.Repeat(" ", espaco))
			}
		}
		out[y] = strings.TrimRight(sb.String(), " ")
	}
	return out
}

// CentralizaBloco põe um bloco no meio de uma tela de dadas dimensões. Linhas
// que sobram embaixo ficam em branco, para a tela não "pular" conforme o
// conteúdo muda de altura.
func CentralizaBloco(bloco []string, largura, altura int) []string {
	larguraBloco := 0
	for _, l := range bloco {
		if w := Largura(l); w > larguraBloco {
			larguraBloco = w
		}
	}
	recuo := (largura - larguraBloco) / 2
	if recuo < 0 {
		recuo = 0
	}
	acima := (altura - len(bloco)) / 2
	if acima < 0 {
		acima = 0
	}

	out := make([]string, 0, altura)
	for i := 0; i < acima; i++ {
		out = append(out, "")
	}
	for _, l := range bloco {
		out = append(out, strings.Repeat(" ", recuo)+l)
	}
	for len(out) < altura {
		out = append(out, "")
	}
	return out
}

// Moldura desenha uma borda em volta do conteúdo, do tamanho exato da tela.
// Serve para o jogador ver onde a tela acaba, em vez de o conteúdo se perder
// no fundo do terminal.
func Moldura(conteudo string, largura, altura int) string {
	interno := largura - 2
	linhas := strings.Split(conteudo, "\n")

	var b strings.Builder
	b.WriteString("." + strings.Repeat("-", interno) + ".\n")
	for i := 0; i < altura-2; i++ {
		l := ""
		if i < len(linhas) {
			l = linhas[i]
		}
		// Corta o que passar da largura: conteúdo vazando quebraria a borda.
		if Largura(l) > interno {
			r := []rune(l)
			l = string(r[:interno])
		}
		b.WriteString("|" + l + strings.Repeat(" ", interno-Largura(l)) + "|\n")
	}
	b.WriteString("'" + strings.Repeat("-", interno) + "'")
	return b.String()
}

// SobreOpaco escreve um bloco por cima do fundo apagando o que havia atrás —
// ao contrário de fogo.Sobre, que deixa o fundo aparecer nos espaços. É o que
// mantém um modal legível sobre a tela em chamas.
func SobreOpaco(fundo, bloco []string) []string {
	out := append([]string(nil), fundo...)
	larguraBloco := 0
	for _, l := range bloco {
		if w := Largura(l); w > larguraBloco {
			larguraBloco = w
		}
	}
	inicio := (len(fundo) - len(bloco)) / 2

	for i, l := range bloco {
		y := inicio + i
		if y < 0 || y >= len(out) {
			continue
		}
		base := []rune(out[y])
		x := (len(base) - larguraBloco) / 2
		preenchido := l + strings.Repeat(" ", larguraBloco-Largura(l))
		for k, ch := range []rune(preenchido) {
			if p := x + k; p >= 0 && p < len(base) {
				base[p] = ch
			}
		}
		out[y] = string(base)
	}
	return out
}
