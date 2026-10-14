package game

import (
	"fmt"
	"strings"

	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

// rodada é a fase de jogar os dados. Os valores são fixos nesta entrega: a
// tela existe para a estrutura estar de pé, e as mecânicas vão preencher
// caixas que já existem em vez de nascerem junto com decisões de layout.
type rodada struct {
	theme ui.Theme
	keys  ui.Keys

	alvo, pontos int64
	fichas, mult int
	desenho      string

	dados    []int
	marcados map[int]bool
	maos     [2]int // usadas, total
	rerolls  [2]int
	runas    []string
	maxRunas int
}

func novaRodada(theme ui.Theme, keys ui.Keys) rodada {
	return rodada{
		theme:    theme,
		keys:     keys,
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

// alterna marca ou desmarca um dado para o reroll. A seleção é por dígito, não
// por cursor: digitar 1, 3 e 5 escolhe três dados de uma vez.
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

func (r rodada) render(largura, altura int) string {
	const larguraPainel = 18

	esquerda := ui.Painel(larguraPainel,
		ui.Rotulado("ALVO", ui.Numero(r.alvo), larguraPainel),
		ui.Rotulado("VOCE", ui.Numero(r.pontos), larguraPainel),
		ui.Separador,
		fmt.Sprintf("%s  x  %s", ui.Numero(int64(r.fichas)), ui.Numero(int64(r.mult))),
		ui.Separador,
		r.desenho,
	)

	var slots []string
	for i := 0; i < r.maxRunas; i++ {
		if i < len(r.runas) {
			slots = append(slots, r.runas[i])
			continue
		}
		slots = append(slots, " ")
	}
	var caixas [][]string
	for _, s := range slots {
		caixas = append(caixas, []string{".---.", "| " + s + " |", "'---'"})
	}
	linhasRunas := ui.LadoALado(1, caixas...)
	centro := ui.Painel(31,
		ui.Rotulado("RUNAS", fmt.Sprintf("%d/%d", len(r.runas), r.maxRunas), 31),
		ui.Separador,
	)
	centro = centro[:len(centro)-1] // tira a borda para encaixar as runas
	for _, l := range linhasRunas {
		centro = append(centro, "|"+ui.Centraliza(l, 31)+"|")
	}
	centro = append(centro, "'"+strings.Repeat("-", 31)+"'")

	direita := ui.Painel(16,
		ui.Rotulado("MAOS", fmt.Sprintf("%d/%d", r.maos[0], r.maos[1]), 16),
		ui.Separador,
		ui.Rotulado("REROLLS", fmt.Sprintf("%d/%d", r.rerolls[0], r.rerolls[1]), 16),
	)

	topo := ui.LadoALado(2, esquerda, centro, direita)

	// mesa
	var ds [][]string
	for _, v := range r.dados {
		ds = append(ds, ui.Dado(v))
	}
	mesa := ui.LadoALado(2, ds...)

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

	bloco := append([]string{}, topo...)
	bloco = append(bloco, "")
	larguraMesa := ui.Largura(marcas.String())
	for _, l := range mesa {
		bloco = append(bloco, ui.Centraliza(l, larguraMesa))
	}
	bloco = append(bloco, marcas.String(), rotulos.String())

	return strings.Join(ui.CentralizaBloco(bloco, largura, altura), "\n")
}

func (r rodada) ajuda() []string {
	return []string{"1-5 marca", "r reroll", "enter envia", "tab loja", "esc menu"}
}
