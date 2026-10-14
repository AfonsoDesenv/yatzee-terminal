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
	return strings.Join(ui.SobreOpaco(s.chama.Linhas(), modal), "\n")
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

// loja segue o arranjo do Balatro: ações à esquerda, prateleira no meio,
// pacotes embaixo. As runas à venda já funcionam; os pacotes continuam vazios.
func (s Screen) loja(largura, altura int) string {
	const larguraAcoes = 18

	acoes := ui.Painel(larguraAcoes, ui.Centraliza("PROXIMO NIVEL", larguraAcoes), ui.Centraliza("[enter]", larguraAcoes))
	acoes = append(acoes, ui.Painel(larguraAcoes,
		ui.Rotulado("moedas", ui.Numero(int64(s.partida.Moedas)), larguraAcoes),
		ui.Rotulado("runas", fmt.Sprintf("%d/%d", len(s.partida.Runas), regras.MaxRunas), larguraAcoes),
	)...)

	larguraVitrine := largura - larguraAcoes - 8
	if larguraVitrine < 40 {
		larguraVitrine = 40
	}

	// Prateleira: as três runas sorteadas para este nível.
	var vitrines [][]string
	for i, r := range s.vitrine {
		const w = 16
		if r.Letra == "" {
			vitrines = append(vitrines, ui.Painel(w, "", ui.Centraliza("vendida", w), "", ""))
			continue
		}
		f, m := r.Aplica(regras.Mao{1, 2, 3, 4, 5})
		_, _ = f, m
		vitrines = append(vitrines, ui.Painel(w,
			ui.Centraliza(fmt.Sprintf("[%d]  %s", i+1, r.Letra), w),
			ui.Separador,
			ui.Centraliza(fmt.Sprintf("%d moedas", PrecoRuna), w),
		))
	}
	prateleira := ui.Painel(larguraVitrine, ui.Centraliza("A VENDA", larguraVitrine), ui.Separador)
	prateleira = prateleira[:len(prateleira)-1]
	for _, l := range ui.LadoALado(2, vitrines...) {
		prateleira = append(prateleira, "|"+ui.Centraliza(l, larguraVitrine)+"|")
	}
	// Uma linha por runa, dizendo o que ela faz: sem isso a letra não informa
	// nada e a compra vira chute.
	for _, r := range s.vitrine {
		if r.Letra == "" {
			continue
		}
		prateleira = append(prateleira, "|"+ui.Centraliza(r.Letra+" - "+r.Descricao(), larguraVitrine)+"|")
	}
	prateleira = append(prateleira, "'"+strings.Repeat("-", larguraVitrine)+"'")

	var caixas [][]string
	for i := 0; i < 2; i++ {
		caixas = append(caixas, ui.Painel(18, "", "", ""))
	}
	pacotes := ui.Painel(larguraVitrine, ui.Centraliza("PACOTES", larguraVitrine), ui.Separador)
	pacotes = pacotes[:len(pacotes)-1]
	for _, l := range ui.LadoALado(4, caixas...) {
		pacotes = append(pacotes, "|"+ui.Centraliza(l, larguraVitrine)+"|")
	}
	pacotes = append(pacotes, "'"+strings.Repeat("-", larguraVitrine)+"'")

	direita := append([]string{}, prateleira...)
	direita = append(direita, pacotes...)

	bloco := ui.LadoALado(2, acoes, direita)
	return strings.Join(ui.CentralizaBloco(bloco, largura, altura), "\n")
}

func moedas(n int) string {
	if n == 1 {
		return "1 moeda"
	}
	return fmt.Sprintf("%d moedas", n)
}
