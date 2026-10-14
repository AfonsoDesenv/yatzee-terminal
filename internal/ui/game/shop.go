package game

import (
	"strings"

	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

// loja é a fase entre rodadas, onde as runas são compradas. Os itens são fixos
// nesta entrega: a tela existe para a estrutura estar de pé.
type loja struct {
	theme  ui.Theme
	moedas int64
	itens  []itemLoja
}

type itemLoja struct {
	nome  string
	preco int64
}

func novaLoja(theme ui.Theme) loja {
	return loja{
		theme:  theme,
		moedas: 12,
		itens: []itemLoja{
			{"RUNA A", 5},
			{"RUNA B", 7},
			{"RUNA C", 4},
		},
	}
}

func (l loja) render(largura, altura int) string {
	topo := ui.Painel(24, ui.Rotulado("MOEDAS", ui.Numero(l.moedas), 24))

	var caixas [][]string
	for _, it := range l.itens {
		caixas = append(caixas, ui.Painel(14, it.nome, ui.Separador, ui.Numero(it.preco)))
	}
	prateleira := ui.LadoALado(3, caixas...)

	bloco := []string{l.theme.Title.Render("L O J A"), ""}
	bloco = append(bloco, topo...)
	bloco = append(bloco, "")
	bloco = append(bloco, prateleira...)
	bloco = append(bloco, "", l.theme.Dim.Render("os efeitos das runas entram aqui"))

	return strings.Join(ui.CentralizaBloco(bloco, largura, altura), "\n")
}

func (l loja) ajuda() []string {
	return []string{"tab volta à rodada", "esc menu"}
}
