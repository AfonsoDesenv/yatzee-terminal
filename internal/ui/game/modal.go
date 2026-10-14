package game

import (
	"fmt"
	"strings"

	"github.com/afonsodesenv/yatzee-terminal/internal/regras"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

// modalVitoria mostra o que a vitória pagou, item por item. O detalhamento é
// o que ensina o jogador a terminar a rodada com folga: sobra vira moeda.
func (s Screen) modalVitoria(largura, altura int) string {
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
		ui.Rotulado("total", moedas(p.Recompensa()), w),
	)

	bloco := append([]string{}, ui.Painel(w, linhas...)...)
	bloco = append(bloco, "", ui.Centraliza("[ continuar ]", w+2))
	return strings.Join(ui.CentralizaBloco(bloco, largura, altura), "\n")
}

// modalDerrota espelha o de vitória, para a partida terminar com um fecho em
// vez de simplesmente voltar ao menu.
func (s Screen) modalDerrota(largura, altura int) string {
	const w = 32
	p := s.partida

	faltou := p.Alvo - p.Pontos
	bloco := append([]string{}, ui.Painel(w,
		ui.Centraliza("ACABARAM AS MAOS", w),
		ui.Separador,
		ui.Rotulado("nivel", fmt.Sprintf("%d", p.Nivel), w),
		ui.Rotulado("pontos", ui.Numero(p.Pontos), w),
		ui.Rotulado("alvo", ui.Numero(p.Alvo), w),
		ui.Separador,
		ui.Rotulado("faltou", ui.Numero(faltou), w),
	)...)
	bloco = append(bloco, "", ui.Centraliza("[ voltar ao menu ]", w+2))
	return strings.Join(ui.CentralizaBloco(bloco, largura, altura), "\n")
}

// loja é o intervalo entre níveis. Vazia por enquanto: as runas entram depois,
// e o botão de continuar já fecha o ciclo.
func (s Screen) loja(largura, altura int) string {
	const w = 32
	bloco := append([]string{}, ui.Painel(w,
		ui.Centraliza("L O J A", w),
		ui.Separador,
		ui.Rotulado("moedas", ui.Numero(int64(s.partida.Moedas)), w),
		ui.Separador,
		ui.Centraliza("as runas entram aqui", w),
	)...)
	bloco = append(bloco, "", ui.Centraliza(
		fmt.Sprintf("[ continuar para o nivel %d ]", s.partida.Nivel+1), w+2))
	return strings.Join(ui.CentralizaBloco(bloco, largura, altura), "\n")
}

func moedas(n int) string {
	if n == 1 {
		return "1 moeda"
	}
	return fmt.Sprintf("%d moedas", n)
}
