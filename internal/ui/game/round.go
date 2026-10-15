package game

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/regras"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

// passo é uma etapa da contagem. Fichas e mult são ambos somados: a combinação
// entra primeiro e define a base, e tudo depois acrescenta por cima.
type passo struct {
	rotulo string
	fichas int
	mult   int
	// dado e runa são as posições que o passo acende na tela, ou -1 para
	// nenhuma. Os dois juntos quando uma runa soma por causa de um dado: a
	// runa que somou e o dado que a fez somar acendem no mesmo tique.
	//
	// São posições na mesa e no painel, não a ordem do passo, porque é o item
	// na tela que tem de acender.
	dado int
	runa int
}

// semItem é o passo que não aponta nada na tela, como o da combinação.
const semItem = -1

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
	ps := []passo{{
		rotulo: p.Combinacao.Nome(),
		fichas: base,
		mult:   mult,
		dado:   semItem,
		runa:   semItem,
	}}

	for i, d := range dados {
		ps = append(ps, passo{
			rotulo: fmt.Sprintf("dado %d", d),
			fichas: d,
			dado:   i,
			runa:   semItem,
		})
	}

	// Cada runa vira um passo por dado que a dispara, e não um passo só com a
	// soma de todos. Com três dados de 6, a runa L soma +2 três vezes, e cada
	// tique mostra qual dado pagou aquele +2 — somado de uma vez, o salto
	// aparecia sem dizer de onde veio.
	for i, r := range runas {
		if !r.Dispara(dados) {
			continue
		}

		gatilhos := r.Gatilhos(dados)
		if len(gatilhos) == 0 {
			// Runa que soma sem depender de dado: um passo, nenhum dado aceso.
			f, m := r.Aplica(dados)
			ps = append(ps, passo{
				rotulo: "runa " + r.Letra,
				fichas: f,
				mult:   m,
				dado:   semItem,
				runa:   i,
			})
			continue
		}

		f, m := r.PorGatilho()
		for _, d := range gatilhos {
			ps = append(ps, passo{
				rotulo: fmt.Sprintf("runa %s dado %d", r.Letra, dados[d]),
				fichas: f,
				mult:   m,
				dado:   d,
				runa:   i,
			})
		}
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

// causa devolve o dado e a runa que o passo recém-entrado acende, ou semItem
// para cada um que ele não aponta. Ficam acesos até o próximo tique trocar o
// passo — o que dá exatamente a duração de um tique.
func (a animacao) causa() (dado, runa int) {
	if a.atual <= 0 || a.atual > len(a.passos) {
		return semItem, semItem
	}
	p := a.passos[a.atual-1]
	return p.dado, p.runa
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

// painelRunas estica para preencher a largura disponível. A runa que está
// somando no tique acende de roxo — sem isso o passo dela passava batido, com
// o nome aparecendo no painel da esquerda e nada apontando qual era.
func (s Screen) painelRunas(largura int) []string {
	_, aceso := s.acende()

	var caixas [][]string
	for i := 0; i < regras.MaxRunas; i++ {
		letra := " "
		if i < len(s.partida.Runas) {
			letra = s.partida.Runas[i].Letra
		}
		caixa := []string{".---.", "| " + letra + " |", "'---'"}
		if i == aceso {
			caixa = ui.Pinta(s.theme.Runa, caixa)
		}
		caixas = append(caixas, caixa)
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
	// Mesa de dados, centralizada na largura. O dado escolhido para o reroll
	// vai de vermelho inteiro — cubo, traço e rótulo —, porque os três são um
	// sinal só: pintar apenas um pedaço deixaria a escolha pela metade.
	//
	// Quem está somando no tique ganha o azul, e ele vence o vermelho: no meio
	// da contagem o reroll já não vai acontecer, e o que importa ver é de onde
	// vieram os pontos que acabaram de entrar.
	aceso, _ := s.acende()
	corDo := func(i int) (lipgloss.Style, bool) {
		if i == aceso {
			return s.theme.DadoContando, true
		}
		if s.partida.Marcados[i] {
			return s.theme.DadoMarcado, true
		}
		return lipgloss.Style{}, false
	}

	var ds [][]string
	for i, v := range s.partida.Dados {
		d := ui.Dado(v)
		if cor, pinta := corDo(i); pinta {
			d = ui.Pinta(cor, d)
		}
		ds = append(ds, d)
	}

	var marcas, rotulos strings.Builder
	for i := range s.partida.Dados {
		marcado := s.partida.Marcados[i]

		marca := ui.MarcaDado(marcado)
		rot := fmt.Sprintf("[%d]", i+1)
		if marcado {
			rot = fmt.Sprintf(">%d<", i+1)
		}
		rot = ui.Centraliza(rot, ui.LarguraDado)

		if cor, pinta := corDo(i); pinta {
			marca = cor.Render(marca)
			rot = cor.Render(rot)
		}
		marcas.WriteString(marca)
		rotulos.WriteString(rot)

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
		tela = ui.Compoe(tela, s.chamaPintada(), xFogo, yFogo, false)
	}

	// O painel da meta entra transparente para o fogo queimar por dentro dele.
	tela = ui.Compoe(tela, esquerdo, xEsq, yTopo, s.chama != nil)
	tela = ui.Compoe(tela, s.painelRunas(meio), xRunas, yTopo, false)
	tela = ui.Compoe(tela, s.painelDireito(), xDir, yTopo, false)
	tela = ui.Compoe(tela, mesa, xEsq, yMesa, false)
	return strings.Join(tela, "\n")
}
