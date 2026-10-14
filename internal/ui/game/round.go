package game

import (
	"fmt"
	"strings"

	"github.com/afonsodesenv/yatzee-terminal/internal/regras"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui/fogo"
)

// passo é uma etapa da contagem: um dado que soma fichas, ou a combinação, que
// soma a base e aplica o multiplicador.
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

// novaAnimacao monta a sequência: cada dado soma o próprio valor, e a
// combinação entra por último com a base e o multiplicador. É a regra da soma
// visível passo a passo, em vez de um número que muda de uma vez.
func novaAnimacao(p regras.Pontuacao, dados regras.Mao) animacao {
	var ps []passo
	for _, d := range dados {
		ps = append(ps, passo{rotulo: fmt.Sprintf("dado %d", d), fichas: d})
	}
	base, mult := p.Combinacao.Base()
	ps = append(ps, passo{rotulo: p.Combinacao.Nome(), fichas: base, mult: mult})
	return animacao{passos: ps, mult: 1}
}

func (a animacao) acabou() bool { return a.atual >= len(a.passos) }

func (a animacao) avanca() animacao {
	if a.acabou() {
		return a
	}
	p := a.passos[a.atual]
	a.fichas += p.fichas
	if p.mult > 0 {
		a.mult *= p.mult
	}
	a.atual++
	return a
}

// rotulo é o que mostrar na linha do desenho: o passo atual durante a
// contagem, a combinação da mão quando parada.
func (a animacao) rotulo(padrao string) string {
	if a.atual > 0 && a.atual <= len(a.passos) {
		return a.passos[a.atual-1].rotulo
	}
	return padrao
}

// painelEsquerdo é alvo, acumulado, contador e desenho.
func (s Screen) painelEsquerdo() []string {
	const w = 18

	fichas, mult := 0, 1
	rotulo := regras.Identifica(s.partida.Dados).Nome()
	if s.phase == PhaseContando {
		fichas, mult = s.anim.fichas, s.anim.mult
		rotulo = s.anim.rotulo(rotulo)
	}

	return ui.Painel(w,
		ui.Rotulado("ALVO", ui.Numero(s.partida.Alvo), w),
		ui.Rotulado("VOCE", ui.Numero(s.parcial()), w),
		ui.Separador,
		fmt.Sprintf("%s  x  %s", ui.Numero(int64(fichas)), ui.Numero(int64(mult))),
		ui.Separador,
		rotulo,
	)
}

// mesa desenha a tela da rodada: painéis em cima, dados embaixo.
func (s Screen) mesa(largura, altura int) string {
	esquerda := s.painelEsquerdo()
	// O painel da meta queima enquanto a contagem passa do alvo.
	if s.chama != nil && s.phase == PhaseContando {
		esquerda = fogo.Sobre(s.chama.Linhas(), esquerda)
	}

	const larguraRunas = 31
	var caixas [][]string
	for i := 0; i < 5; i++ {
		caixas = append(caixas, []string{".---.", "|   |", "'---'"})
	}
	centro := ui.Painel(larguraRunas,
		ui.Rotulado("RUNAS", "0/5", larguraRunas),
		ui.Separador,
	)
	centro = centro[:len(centro)-1]
	for _, l := range ui.LadoALado(1, caixas...) {
		centro = append(centro, "|"+ui.Centraliza(l, larguraRunas)+"|")
	}
	centro = append(centro, "'"+strings.Repeat("-", larguraRunas)+"'")

	direita := ui.Painel(16,
		ui.Rotulado("NIVEL", fmt.Sprintf("%d", s.partida.Nivel), 16),
		ui.Separador,
		ui.Rotulado("MAOS", fmt.Sprintf("%d/%d", s.partida.Maos, regras.MaosPorRodada), 16),
		ui.Rotulado("REROLLS", fmt.Sprintf("%d/%d", s.partida.Rerolls, regras.RerollsPorRodada), 16),
		ui.Separador,
		ui.Rotulado("MOEDAS", ui.Numero(int64(s.partida.Moedas)), 16),
	)

	bloco := append([]string{}, ui.LadoALado(2, esquerda, centro, direita)...)
	bloco = append(bloco, "")

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
	larguraMesa := ui.Largura(marcas.String())
	for _, l := range ui.LadoALado(2, ds...) {
		bloco = append(bloco, ui.Centraliza(l, larguraMesa))
	}
	bloco = append(bloco, marcas.String(), rotulos.String())

	return strings.Join(ui.CentralizaBloco(bloco, largura, altura), "\n")
}
