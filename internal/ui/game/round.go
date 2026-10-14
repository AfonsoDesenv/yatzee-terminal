package game

import (
	"fmt"
	"strings"

	"github.com/afonsodesenv/yatzee-terminal/internal/regras"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

// passo é uma etapa da contagem. Fichas e mult são ambos somados: a combinação
// entra primeiro e define a base, e tudo depois acrescenta por cima.
type passo struct {
	rotulo string
	fichas int
	mult   int
}

// animacao percorre os passos de uma jogada.
type animacao struct {
	passos []passo
	atual  int
	fichas int
	mult   int
}

// novaAnimacao monta a sequência na ordem do Balatro: a combinação primeiro,
// fixando fichas e multiplicador; depois cada dado somando o próprio valor; e
// por fim as runas que dispararam.
//
// A ordem importa além da estética: com o multiplicador definido logo no
// começo, o total cresce de verdade a cada passo e pode cruzar a meta no meio
// da contagem. Somando os dados antes, a conta ficava rasteira até o último
// passo e a travessia só acontecia no fim.
func novaAnimacao(p regras.Pontuacao, dados regras.Mao, runas []regras.Runa) animacao {
	base, mult := p.Combinacao.Base()
	ps := []passo{{rotulo: p.Combinacao.Nome(), fichas: base, mult: mult}}

	for _, d := range dados {
		ps = append(ps, passo{rotulo: fmt.Sprintf("dado %d", d), fichas: d})
	}
	for _, r := range runas {
		if !r.Dispara(dados) {
			continue
		}
		f, m := r.Aplica(dados)
		ps = append(ps, passo{rotulo: "runa " + r.Letra, fichas: f, mult: m})
	}
	return animacao{passos: ps}
}

func (a animacao) acabou() bool { return a.atual >= len(a.passos) }

func (a animacao) avanca() animacao {
	if a.acabou() {
		return a
	}
	p := a.passos[a.atual]
	a.fichas += p.fichas
	a.mult += p.mult
	a.atual++
	return a
}

func (a animacao) rotulo(padrao string) string {
	if a.atual > 0 && a.atual <= len(a.passos) {
		return a.passos[a.atual-1].rotulo
	}
	return padrao
}

// larguras dos painéis da linha de cima. O do meio estica para ocupar o que
// sobrar da largura da tela.
const (
	larguraEsquerda = 20
	larguraDireita  = 20
	espacoPainel    = 2
)

func (s Screen) painelEsquerdo() []string {
	fichas, mult := 0, 0
	rotulo := regras.Identifica(s.partida.Dados).Nome()
	if s.phase == PhaseContando {
		fichas, mult = s.anim.fichas, s.anim.mult
		rotulo = s.anim.rotulo(rotulo)
	}

	const w = larguraEsquerda
	return ui.Painel(w,
		ui.Rotulado("ALVO", ui.Numero(s.partida.Alvo), w),
		ui.Rotulado("VOCE", ui.Numero(s.parcial()), w),
		ui.Separador,
		ui.Centraliza(fmt.Sprintf("%s  x  %s", ui.Numero(int64(fichas)), ui.Numero(int64(mult))), w),
		ui.Separador,
		ui.Centraliza(rotulo, w),
	)
}

// painelRunas estica para preencher a largura disponível.
func (s Screen) painelRunas(largura int) []string {
	var caixas [][]string
	for i := 0; i < regras.MaxRunas; i++ {
		letra := " "
		if i < len(s.partida.Runas) {
			letra = s.partida.Runas[i].Letra
		}
		caixas = append(caixas, []string{".---.", "| " + letra + " |", "'---'"})
	}

	out := ui.Painel(largura,
		ui.Rotulado("RUNAS", fmt.Sprintf("%d/%d", len(s.partida.Runas), regras.MaxRunas), largura),
		ui.Separador,
	)
	out = out[:len(out)-1]
	for _, l := range ui.LadoALado(2, caixas...) {
		out = append(out, "|"+ui.Centraliza(l, largura)+"|")
	}
	return append(out, "'"+strings.Repeat("-", largura)+"'")
}

func (s Screen) painelDireito() []string {
	const w = larguraDireita
	return ui.Painel(w,
		ui.Rotulado("NIVEL", fmt.Sprintf("%d", s.partida.Nivel), w),
		ui.Separador,
		ui.Rotulado("MAOS", fmt.Sprintf("%d/%d", s.partida.Maos, regras.MaosPorRodada), w),
		ui.Rotulado("REROLLS", fmt.Sprintf("%d/%d", s.partida.Rerolls, regras.RerollsPorRodada), w),
		ui.Separador,
		ui.Rotulado("MOEDAS", ui.Numero(int64(s.partida.Moedas)), w),
	)
}

// mesa desenha a rodada: os painéis ocupam a largura toda no alto, e os dados
// ficam no meio do espaço que sobra — em vez de tudo espremido num bloco
// centralizado, que deixava buracos esquisitos em tela grande.
func (s Screen) mesa(largura, altura int) string {
	conteudo := largura - 2*ui.Margem
	meio := conteudo - larguraEsquerda - larguraDireita - 4 - 2*espacoPainel
	if meio < 20 {
		meio = 20
	}
	// Mesa de dados, centralizada na largura.
	var ds [][]string
	for _, v := range s.partida.Dados {
		ds = append(ds, ui.Dado(v))
	}
	var marcas, rotulos strings.Builder
	for i := range s.partida.Dados {
		marcado := s.partida.Marcados[i]
		marcas.WriteString(ui.MarcaDado(marcado))
		rot := fmt.Sprintf("[%d]", i+1)
		if marcado {
			rot = fmt.Sprintf(">%d<", i+1)
		}
		rotulos.WriteString(ui.Centraliza(rot, ui.LarguraDado))
		if i < len(s.partida.Dados)-1 {
			marcas.WriteString("  ")
			rotulos.WriteString("  ")
		}
	}

	var mesa []string
	for _, l := range ui.LadoALado(2, ds...) {
		mesa = append(mesa, ui.Centraliza(l, conteudo))
	}
	mesa = append(mesa,
		ui.Centraliza(marcas.String(), conteudo),
		ui.Centraliza(rotulos.String(), conteudo),
	)

	// Tudo é composto numa tela de posição fixa: cada painel tem seu lugar,
	// os dados ficam no meio do que sobra. O fogo entra atrás do painel da
	// meta, também em posição fixa — se ele substituísse o bloco do painel, a
	// largura maior empurraria o resto da linha de lugar a cada quadro.
	esquerdo := s.painelEsquerdo()
	xEsq := ui.Margem
	xRunas := xEsq + larguraEsquerda + 2 + espacoPainel
	xDir := xRunas + meio + 2 + espacoPainel

	yTopo := ui.MargemY
	tela := ui.Tela(largura, altura)
	yMesa := yTopo + len(esquerdo) + (altura-yTopo-len(esquerdo)-len(mesa))/2

	if s.chama != nil {
		// A chama é mais larga que o painel e sobra para os lados: à esquerda
		// cabe na margem invisível, à direita some sob o painel das runas, que
		// é composto depois e opaco.
		// Na vertical a base da chama fica na borda de baixo do painel, para
		// o fogo subir atraves dele em vez de brotar solto embaixo.
		xFogo := xEsq + (larguraEsquerda+2-larguraFogoPainel)/2
		yFogo := yTopo + len(esquerdo) - alturaFogoPainel
		tela = ui.Compoe(tela, s.chama.Linhas(), xFogo, yFogo, false)
	}

	// O painel da meta entra transparente para o fogo queimar por dentro dele.
	tela = ui.Compoe(tela, esquerdo, xEsq, yTopo, s.chama != nil)
	tela = ui.Compoe(tela, s.painelRunas(meio), xRunas, yTopo, false)
	tela = ui.Compoe(tela, s.painelDireito(), xDir, yTopo, false)
	tela = ui.Compoe(tela, mesa, xEsq, yMesa, false)
	return strings.Join(tela, "\n")
}
