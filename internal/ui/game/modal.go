package game

import (
	"fmt"
	"strings"

	"github.com/afonsodesenv/yatzee-terminal/internal/regras"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

// telaVitoria põe o modal por cima da tela em chamas. O modal é opaco: o fogo
// queima em volta dele, não dentro, senão os números da recompensa ficariam
// ilegíveis justo na hora de lê-los.
func (s Screen) telaVitoria(largura, altura int) string {
	modal := s.blocoVitoria()

	if s.chama == nil {
		return strings.Join(ui.CentralizaBloco(modal, largura, altura), "\n")
	}
	return strings.Join(ui.SobreOpaco(s.chamaPintada(), modal), "\n")
}

func (s Screen) blocoVitoria() []string {
	const w = 32
	p := s.partida

	linhas := []string{
		ui.Centraliza("META BATIDA", w),
		ui.Separador,
		ui.Rotulado("nivel", fmt.Sprintf("%d", p.Nivel), w),
		ui.Rotulado("pontos", ui.Numero(p.Pontos), w),
		ui.Rotulado("alvo", ui.Numero(p.Alvo), w),
		ui.Separador,
		ui.Rotulado("vitoria", moedas(regras.MoedasPorVitoria), w),
	}
	if p.Maos > 0 {
		linhas = append(linhas, ui.Rotulado(
			fmt.Sprintf("%d maos de sobra", p.Maos),
			moedas(p.Maos*regras.MoedasPorMaoSobrando), w))
	}
	if p.Rerolls > 0 {
		linhas = append(linhas, ui.Rotulado(
			fmt.Sprintf("%d rerolls de sobra", p.Rerolls),
			moedas(p.Rerolls*regras.MoedasPorRerollSobrando), w))
	}
	linhas = append(linhas,
		ui.Separador,
		ui.Rotulado("ganhou", moedas(p.Recompensa()), w),
		ui.Rotulado("total", moedas(p.Moedas), w),
	)

	bloco := append([]string{}, ui.Painel(w, linhas...)...)
	return append(bloco, "", ui.Centraliza("[ continuar ]", w+2))
}

// modalDerrota espelha o de vitória, para a partida terminar com um fecho em
// vez de simplesmente voltar ao menu.
func (s Screen) modalDerrota(largura, altura int) string {
	const w = 32
	p := s.partida

	bloco := append([]string{}, ui.Painel(w,
		ui.Centraliza("ACABARAM AS MAOS", w),
		ui.Separador,
		ui.Rotulado("nivel", fmt.Sprintf("%d", p.Nivel), w),
		ui.Rotulado("pontos", ui.Numero(p.Pontos), w),
		ui.Rotulado("alvo", ui.Numero(p.Alvo), w),
		ui.Separador,
		ui.Rotulado("faltou", ui.Numero(p.Alvo-p.Pontos), w),
	)...)
	bloco = append(bloco, "", ui.Centraliza("[ voltar ao menu ]", w+2))
	return strings.Join(ui.CentralizaBloco(bloco, largura, altura), "\n")
}

// Medidas da loja. A prateleira é dimensionada pelas cartas, e não pela tela:
// com a largura flutuando, a loja mudava de forma a cada tamanho de janela e
// as cartas nadavam dentro dela.
const (
	larguraAcoes   = 18
	larguraCarta   = 16
	vaoCarta       = 2
	larguraVitrine = 3*larguraCarta + 2*vaoCarta
)

// loja segue o arranjo do Balatro: ações à esquerda, prateleira à direita,
// pacotes embaixo. O retângulo entre a prateleira e os pacotes descreve o que
// está em foco — sem ele a letra na carta não diria nada e comprar seria
// chute.
func (s Screen) loja(largura, altura int) string {
	bloco := ui.LadoALado(2, s.acoes(), s.prateleira())
	return strings.Join(ui.CentralizaBloco(bloco, largura, altura), "\n")
}

// acoes é a coluna da esquerda: o botão de continuar e o saldo.
func (s Screen) acoes() []string {
	const w = larguraAcoes

	rotulo := "CONTINUAR"
	if s.sel == focoContinuar {
		rotulo = "> " + rotulo + " <"
	}

	out := ui.Painel(w, ui.Centraliza(rotulo, w))
	if s.sel == focoContinuar {
		out = ui.Pinta(s.theme.BotaoEmFoco, out)
	}
	return append(out, ui.Painel(w,
		ui.Rotulado("moedas", ui.Numero(int64(s.partida.Moedas)), w),
		ui.Rotulado("runas", fmt.Sprintf("%d/%d", len(s.partida.Runas), regras.MaxRunas), w),
	)...)
}

// prateleira empilha a vitrine, o retângulo de descrição e os pacotes.
func (s Screen) prateleira() []string {
	var cartas [][]string
	for i := range s.vitrine {
		cartas = append(cartas, s.carta(i))
	}

	out := ui.Painel(larguraVitrine, ui.Centraliza("A VENDA", larguraVitrine), ui.Separador)
	out = out[:len(out)-1]
	for _, l := range ui.LadoALado(vaoCarta, cartas...) {
		out = append(out, "|"+ui.Centraliza(l, larguraVitrine)+"|")
	}
	out = append(out,
		"|"+ui.Centraliza(s.marcaDaVitrine(), larguraVitrine)+"|",
		"'"+strings.Repeat("-", larguraVitrine)+"'",
	)

	out = append(out, s.descricao()...)
	return append(out, pacotes()...)
}

// carta é a runa i na prateleira. Altura fixa: uma carta vendida que
// encolhesse puxaria a prateleira inteira para cima quando alguém comprasse.
func (s Screen) carta(i int) []string {
	const w = larguraCarta - 2 // o painel acrescenta as duas bordas
	r := s.vitrine[i]

	var out []string
	if r.Letra == "" {
		out = ui.Painel(w, ui.Centraliza("-", w), ui.Separador, ui.Centraliza("vendida", w))
	} else {
		out = ui.Painel(w,
			ui.Centraliza(r.Letra, w),
			ui.Separador,
			ui.Centraliza(fmt.Sprintf("%d moedas", PrecoRuna), w),
		)
	}

	if s.sel == focoPrimeiraRuna+i {
		out = ui.Pinta(s.theme.RunaEmFoco, out)
	}
	return out
}

// marcaDaVitrine aponta a carta em foco com o mesmo traço que marca um dado na
// mesa: a tela inteira usa um sinal só para "esta é a escolhida".
func (s Screen) marcaDaVitrine() string {
	var b strings.Builder
	for i := range s.vitrine {
		marca := strings.Repeat(" ", larguraCarta)
		if s.sel == focoPrimeiraRuna+i {
			marca = ui.Centraliza(strings.Repeat("^", larguraCarta-4), larguraCarta)
			marca = s.theme.RunaEmFoco.Render(marca)
		}
		b.WriteString(marca)
		if i < len(s.vitrine)-1 {
			b.WriteString(strings.Repeat(" ", vaoCarta))
		}
	}
	return b.String()
}

// descricao é o retângulo abaixo da prateleira. São sempre três linhas, mesmo
// quando alguma fica vazia, para a loja não mudar de altura conforme o foco
// anda — um texto mais curto encolheria a caixa e empurraria os pacotes.
func (s Screen) descricao() []string {
	const w = larguraVitrine

	titulo, efeito, estado := "CONTINUAR", "comeca o proximo nivel", ""
	if r, ok := s.emFoco(); ok {
		switch {
		case r.Letra == "":
			titulo, efeito, estado = "VENDIDA", "esta runa ja foi levada", ""
		default:
			titulo, efeito, estado = "RUNA "+r.Letra, r.Descricao(), s.preco()
		}
	} else {
		estado = "alvo do proximo nivel: " + ui.Numero(s.partida.ProximoAlvo())
	}

	return ui.Painel(w,
		ui.Centraliza(titulo, w),
		ui.Separador,
		ui.Centraliza(efeito, w),
		ui.Centraliza(estado, w),
	)
}

// preco diz o que a runa em foco custa e, quando a compra não dá, por quê.
// Saber o motivo evita a pessoa apertar espaço de novo achando que falhou.
func (s Screen) preco() string {
	custo := fmt.Sprintf("custa %d moedas", PrecoRuna)

	switch {
	case len(s.partida.Runas) >= regras.MaxRunas:
		return fmt.Sprintf("suas %d runas ja estao ocupadas", regras.MaxRunas)
	case s.partida.Moedas < PrecoRuna:
		return fmt.Sprintf("%s - voce tem so %d", custo, s.partida.Moedas)
	}
	return fmt.Sprintf("%s - voce tem %d", custo, s.partida.Moedas)
}

// pacotes é espaço reservado: o arranjo do Balatro já tem o lugar deles, o
// conteúdo vem depois.
func pacotes() []string {
	var caixas [][]string
	for i := 0; i < 2; i++ {
		caixas = append(caixas, ui.Painel(18, "", "", ""))
	}

	out := ui.Painel(larguraVitrine, ui.Centraliza("PACOTES", larguraVitrine), ui.Separador)
	out = out[:len(out)-1]
	for _, l := range ui.LadoALado(4, caixas...) {
		out = append(out, "|"+ui.Centraliza(l, larguraVitrine)+"|")
	}
	return append(out, "'"+strings.Repeat("-", larguraVitrine)+"'")
}

func moedas(n int) string {
	if n == 1 {
		return "1 moeda"
	}
	return fmt.Sprintf("%d moedas", n)
}
