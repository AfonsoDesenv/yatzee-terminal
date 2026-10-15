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
		// Corta em células: contando por rune, os bytes de um escape ANSI
		// entrariam na conta e o corte cairia no meio da marcação, estragando
		// a linha inteira em vez de encurtá-la.
		if cs := celulas(s); len(cs) > largura {
			return junta(cs[:largura])
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
		// O corte é em células, para não partir um escape ANSI no meio e
		// despejar os bytes dele na tela.
		if Largura(l) > interno {
			l = junta(celulas(l)[:interno])
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
		// Em células, como em Compoe: é o que deixa o modal e o fundo serem
		// pintados sem a marcação ocupar coluna e empurrar o desenho.
		base := celulas(out[y])
		x := (len(base) - larguraBloco) / 2
		preenchido := l + strings.Repeat(" ", larguraBloco-Largura(l))
		for k, c := range celulas(preenchido) {
			if p := x + k; p >= 0 && p < len(base) {
				base[p] = c
			}
		}
		out[y] = junta(base)
	}
	return out
}

// Margem é o respiro invisível entre a borda da tela e o conteúdo. Além de
// afastar o desenho da moldura, é o espaço onde o fogo do painel cabe sem
// empurrar nada de lugar.
const Margem = 6

// MargemY e o mesmo respiro na vertical. E o que da altura para a chama subir
// acima do painel: sem ele o fogo comecaria na primeira linha e nao teria para
// onde ir.
const MargemY = 2

// Tela devolve uma superfície em branco para compor.
func Tela(largura, altura int) []string {
	linhas := make([]string, altura)
	vazia := strings.Repeat(" ", largura)
	for i := range linhas {
		linhas[i] = vazia
	}
	return linhas
}

// Compoe desenha um bloco sobre a tela numa posição fixa. Com transparente, os
// espaços do bloco deixam ver o que está atrás — é o que permite o fogo
// aparecer por dentro de um painel sem apagar suas bordas.
//
// Compor em posição fixa, em vez de substituir o bloco, é o que mantém o
// layout parado: um fogo mais largo que o painel mudaria a largura do bloco e
// empurraria todo o resto da linha.
func Compoe(tela []string, bloco []string, x, y int, transparente bool) []string {
	out := append([]string(nil), tela...)
	for i, l := range bloco {
		ly := y + i
		if ly < 0 || ly >= len(out) {
			continue
		}
		base := celulas(out[ly])
		for k, c := range celulas(l) {
			// A transparência olha o caractere, não a célula inteira: um
			// espaço pintado continua sendo espaço, e o que está atrás dele
			// tem de aparecer igual.
			if transparente && c.ch == ' ' {
				continue
			}
			if p := x + k; p >= 0 && p < len(base) {
				base[p] = c
			}
		}
		out[ly] = junta(base)
	}
	return out
}

// Pinta aplica um estilo a cada linha de um bloco, de uma vez por linha. Por
// linha e não por caractere de propósito: um escape por caractere multiplicaria
// o tamanho do quadro e cortaria qualquer busca por texto na tela.
func Pinta(e lipgloss.Style, bloco []string) []string {
	out := make([]string, len(bloco))
	for i, l := range bloco {
		out[i] = e.Render(l)
	}
	return out
}

// reset é o escape que desliga toda a marcação.
const reset = "\x1b[0m"

// celula é um caractere e a marcação que o pinta. Compor em células, e não em
// runes, é o que mantém o desenho no lugar: os bytes de um escape ANSI não
// ocupam coluna na tela, mas ocupariam uma posição num []rune e empurrariam
// todo o resto da linha.
type celula struct {
	estilo string
	ch     rune
}

// celulas parte uma linha em células de uma coluna cada, carregando em cada
// uma o estilo que valia ali. Guarda o estado, e não só o escape vizinho: numa
// linha pintada de ponta a ponta o escape aparece uma vez, no começo, e vale
// para todos os caracteres até o reset.
func celulas(s string) []celula {
	var out []celula
	estilo := ""
	r := []rune(s)

	for i := 0; i < len(r); i++ {
		if r[i] == '\x1b' {
			fim := fimDoEscape(r, i)
			if e := string(r[i:fim]); e == reset || e == "\x1b[m" {
				estilo = ""
			} else {
				estilo += e
			}
			i = fim - 1
			continue
		}
		out = append(out, celula{estilo: estilo, ch: r[i]})
	}
	return out
}

// fimDoEscape devolve o índice logo depois da sequência que começa em i. Um
// escape CSI vai até a primeira letra ou símbolo da faixa @ a ~; sem achar o
// fim, engole o resto da linha, que é melhor do que devolver o escape como se
// fosse texto para a tela.
func fimDoEscape(r []rune, i int) int {
	for j := i + 1; j < len(r); j++ {
		if r[j] >= '@' && r[j] <= '~' && r[j] != '[' {
			return j + 1
		}
	}
	return len(r)
}

// junta remonta a linha, emitindo a marcação só quando ela muda e fechando com
// um reset se a linha acabar pintada — sem o reset a cor vazaria para o resto
// da tela.
func junta(cs []celula) string {
	var b strings.Builder
	atual := ""

	for _, c := range cs {
		if c.estilo != atual {
			if atual != "" {
				b.WriteString(reset)
			}
			b.WriteString(c.estilo)
			atual = c.estilo
		}
		b.WriteRune(c.ch)
	}
	if atual != "" {
		b.WriteString(reset)
	}
	return b.String()
}
