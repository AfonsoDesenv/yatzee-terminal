package game

import (
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui/fogo"
)

// Os três relógios da jogada. São separados de propósito: o passo da pontuação
// é lento o bastante para a pessoa ler o que cada runa fez, e a chama precisa
// ser rápida o bastante para parecer fogo em vez de slideshow.
const (
	PassoPontuacao = 500 * time.Millisecond
	QuadroFogo     = 50 * time.Millisecond

	// A comemoração escala: o painel pega fogo primeiro e depois as chamas
	// tomam a tela inteira.
	FogoNoPainel = 1200 * time.Millisecond
	FogoNaTela   = 1800 * time.Millisecond
)

// Mensagens da animação. Ficam no pacote porque só esta tela as produz e
// consome; a raiz apenas as repassa para a tela do topo.
type (
	tiquePontuacao struct{}
	tiqueFogo      struct{}
	espalhaFogo    struct{}
	apagaFogo      struct{}
)

// passo é uma etapa da contagem: um dado que pontua ou uma runa que dispara.
type passo struct {
	rotulo string
	fichas int
	mult   int
}

// escopo diz até onde a chama vai.
type escopo int

const (
	semFogo escopo = iota
	fogoPainel
	fogoTela
)

// animacao é o estado da jogada sendo contada. Nil quando a rodada está
// parada esperando o jogador.
type animacao struct {
	passos []passo
	atual  int
	fichas int
	mult   int

	escopo escopo
	chama  *fogo.Fogo
}

// rodada é a fase de jogar os dados. Os valores são fixos nesta entrega: as
// mecânicas vão preencher caixas que já existem.
type rodada struct {
	theme ui.Theme
	keys  ui.Keys
	rnd   *rand.Rand

	alvo, pontos int64
	fichas, mult int
	desenho      string

	dados    []int
	marcados map[int]bool
	maos     [2]int
	rerolls  [2]int
	runas    []string
	maxRunas int

	anim *animacao
}

func novaRodada(theme ui.Theme, keys ui.Keys) rodada {
	var semente [16]byte
	_, _ = crand.Read(semente[:])

	return rodada{
		theme: theme,
		keys:  keys,
		rnd: rand.New(rand.NewPCG(
			binary.LittleEndian.Uint64(semente[0:8]),
			binary.LittleEndian.Uint64(semente[8:16]),
		)),
		alvo:     2000,
		pontos:   1340,
		fichas:   24,
		mult:     8,
		desenho:  "PAR",
		dados:    []int{2, 1, 5, 6, 3},
		marcados: map[int]bool{},
		maos:     [2]int{2, 4},
		rerolls:  [2]int{3, 3},
		runas:    []string{"A", "B", "C"},
		maxRunas: 5,
	}
}

// animando diz se a jogada está sendo contada ou comemorada. Enquanto for
// verdadeiro a tela ignora o teclado: deixar o jogador mexer no meio da
// contagem bagunçaria o estado que ela está percorrendo.
func (r rodada) animando() bool { return r.anim != nil }

// alterna marca ou desmarca um dado para o reroll. A seleção é por dígito, não
// por cursor: digitar 1, 3 e 5 escolhe três dados em três toques.
func (r rodada) alterna(n int) rodada {
	if n < 1 || n > len(r.dados) {
		return r
	}
	copia := make(map[int]bool, len(r.marcados))
	for k, v := range r.marcados {
		copia[k] = v
	}
	i := n - 1
	if copia[i] {
		delete(copia, i)
	} else {
		copia[i] = true
	}
	r.marcados = copia
	return r
}

// envia começa a contagem da jogada. Os passos são fixos por enquanto; com as
// mecânicas, saem dos dados na mesa e das runas compradas.
func (r rodada) envia() (rodada, tea.Cmd) {
	if r.animando() {
		return r, nil
	}
	r.anim = &animacao{
		passos: []passo{
			{"dado 5", 5, 0},
			{"dado 5", 5, 0},
			{"PAR", 10, 2},
			{"runa A", 0, 3},
			{"runa B", 40, 0},
			{"runa C", 0, 3},
		},
		mult: 1,
	}
	r.fichas, r.mult = 0, 1
	return r, tique(PassoPontuacao, tiquePontuacao{})
}

func tique(d time.Duration, msg tea.Msg) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg { return msg })
}

// avanca aplica um passo da contagem e devolve o comando do próximo.
func (r rodada) avanca() (rodada, tea.Cmd) {
	if !r.animando() || r.anim.escopo != semFogo {
		return r, nil
	}
	a := *r.anim

	p := a.passos[a.atual]
	a.fichas += p.fichas
	if p.mult > 0 {
		a.mult *= p.mult
	}
	a.atual++
	r.fichas, r.mult = a.fichas, a.mult

	if a.atual < len(a.passos) {
		r.anim = &a
		return r, tique(PassoPontuacao, tiquePontuacao{})
	}

	// Fim da contagem: a jogada entra no acumulado.
	r.pontos += int64(a.fichas * a.mult)
	if r.pontos < r.alvo {
		r.anim = nil
		return r, nil
	}

	// Meta batida: o painel pega fogo, e daqui a pouco a tela inteira.
	a.escopo = fogoPainel
	a.chama = fogo.Novo(larguraFogoPainel, alturaFogoPainel, r.rnd)
	r.anim = &a
	return r, tea.Batch(
		tique(QuadroFogo, tiqueFogo{}),
		tique(FogoNoPainel, espalhaFogo{}),
	)
}

const (
	larguraFogoPainel = 34
	alturaFogoPainel  = 11
)

// espalha troca a fogueira do painel por uma do tamanho da tela.
func (r rodada) espalha(largura, altura int) (rodada, tea.Cmd) {
	if !r.animando() || r.anim.escopo != fogoPainel {
		return r, nil
	}
	a := *r.anim
	a.escopo = fogoTela
	a.chama = fogo.Novo(largura, altura, r.rnd)
	r.anim = &a
	return r, tique(FogoNaTela, apagaFogo{})
}

// queima avança um quadro da chama e encerra a comemoração quando ela morre.
func (r rodada) queima() (rodada, tea.Cmd) {
	if !r.animando() || r.anim.chama == nil {
		return r, nil
	}
	a := *r.anim
	a.chama.Passo()
	if !a.chama.Vivo() {
		r.anim = nil
		return r, nil
	}
	r.anim = &a
	return r, tique(QuadroFogo, tiqueFogo{})
}

// apaga manda a chama morrer; ela ainda sobe por alguns quadros.
func (r rodada) apaga() rodada {
	if r.animando() && r.anim.chama != nil {
		r.anim.chama.Apagar()
	}
	return r
}

func (r rodada) painelEsquerdo() []string {
	const w = 18
	ultimo := r.desenho
	if r.animando() && r.anim.atual > 0 && r.anim.escopo == semFogo {
		ultimo = r.anim.passos[r.anim.atual-1].rotulo
	}
	return ui.Painel(w,
		ui.Rotulado("ALVO", ui.Numero(r.alvo), w),
		ui.Rotulado("VOCE", ui.Numero(r.pontos), w),
		ui.Separador,
		fmt.Sprintf("%s  x  %s", ui.Numero(int64(r.fichas)), ui.Numero(int64(r.mult))),
		ui.Separador,
		ultimo,
	)
}

func (r rodada) render(largura, altura int) string {
	// Comemoração na tela inteira: a chama cobre tudo e só o painel aparece.
	if r.animando() && r.anim.escopo == fogoTela {
		return strings.Join(fogo.Sobre(r.anim.chama.Linhas(), r.painelEsquerdo()), "\n")
	}

	esquerda := r.painelEsquerdo()
	// Comemoração no painel: a chama fica contida nele.
	if r.animando() && r.anim.escopo == fogoPainel {
		esquerda = fogo.Sobre(r.anim.chama.Linhas(), esquerda)
	}

	var caixas [][]string
	for i := 0; i < r.maxRunas; i++ {
		nome := " "
		if i < len(r.runas) {
			nome = r.runas[i]
		}
		caixas = append(caixas, []string{".---.", "| " + nome + " |", "'---'"})
	}
	centro := ui.Painel(31,
		ui.Rotulado("RUNAS", fmt.Sprintf("%d/%d", len(r.runas), r.maxRunas), 31),
		ui.Separador,
	)
	centro = centro[:len(centro)-1]
	for _, l := range ui.LadoALado(1, caixas...) {
		centro = append(centro, "|"+ui.Centraliza(l, 31)+"|")
	}
	centro = append(centro, "'"+strings.Repeat("-", 31)+"'")

	direita := ui.Painel(16,
		ui.Rotulado("MAOS", fmt.Sprintf("%d/%d", r.maos[0], r.maos[1]), 16),
		ui.Separador,
		ui.Rotulado("REROLLS", fmt.Sprintf("%d/%d", r.rerolls[0], r.rerolls[1]), 16),
	)

	bloco := append([]string{}, ui.LadoALado(2, esquerda, centro, direita)...)
	bloco = append(bloco, "")

	var ds [][]string
	for _, v := range r.dados {
		ds = append(ds, ui.Dado(v))
	}
	var marcas, rotulos strings.Builder
	for i := range r.dados {
		marcas.WriteString(ui.MarcaDado(r.marcados[i]))
		rot := fmt.Sprintf("[%d]", i+1)
		if r.marcados[i] {
			rot = fmt.Sprintf(">%d<", i+1)
		}
		rotulos.WriteString(ui.Centraliza(rot, ui.LarguraDado))
		if i < len(r.dados)-1 {
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

func (r rodada) ajuda() []string {
	if r.animando() {
		if r.anim.escopo != semFogo {
			return []string{"meta batida!"}
		}
		return []string{"contando a jogada..."}
	}
	return []string{"1-5 marca", "r reroll", "enter envia", "tab loja", "esc menu"}
}
