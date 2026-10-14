// Package fogo simula fogo para a comemoração de bater a meta.
//
// É o algoritmo clássico do Doom: a linha de baixo é a fonte de calor, e cada
// célula acima recebe o calor da de baixo menos um decaimento aleatório, com
// deriva lateral. Ninguém desenha quadro nenhum — o movimento emerge da
// propagação, e é por isso que a chama nunca se repete.
package fogo

import (
	"math/rand/v2"
	"strings"
)

// Rampa vai do frio ao brasa. O espaço no início é a ausência de fogo, e é o
// que deixa o fundo da tela aparecer entre as labaredas.
const Rampa = " .:-=+*#%@"

// Calor máximo de uma célula.
var calorMaximo = len(Rampa) - 1

// Fogo é uma fogueira em andamento. O gerador vem de fora para a animação ser
// reproduzível em teste: mesma semente, mesma chama.
type Fogo struct {
	largura, altura int
	calor           []int
	rnd             *rand.Rand
	aceso           bool
}

// Novo acende uma fogueira do tamanho pedido.
func Novo(largura, altura int, rnd *rand.Rand) *Fogo {
	f := &Fogo{
		largura: largura,
		altura:  altura,
		calor:   make([]int, largura*altura),
		rnd:     rnd,
		aceso:   true,
	}
	for x := 0; x < largura; x++ {
		f.calor[(altura-1)*largura+x] = calorMaximo
	}
	return f
}

// Apagar manda a fogueira morrer: a fonte esfria aos poucos e a chama sobe até
// acabar, em vez de sumir de uma vez.
func (f *Fogo) Apagar() { f.aceso = false }

// Vivo diz se ainda há calor em algum lugar. Enquanto for verdadeiro vale
// continuar animando.
func (f *Fogo) Vivo() bool {
	for _, c := range f.calor {
		if c > 0 {
			return true
		}
	}
	return false
}

// Passo propaga uma linha de calor para cima e esfria a fonte, se estiver
// apagando.
func (f *Fogo) Passo() {
	for y := 1; y < f.altura; y++ {
		for x := 0; x < f.largura; x++ {
			decai := f.rnd.IntN(3)
			destino := x - decai + 1
			if destino < 0 {
				destino = 0
			}
			if destino >= f.largura {
				destino = f.largura - 1
			}
			novo := f.calor[y*f.largura+x] - decai
			if novo < 0 {
				novo = 0
			}
			f.calor[(y-1)*f.largura+destino] = novo
		}
	}

	if !f.aceso {
		base := (f.altura - 1) * f.largura
		for x := 0; x < f.largura; x++ {
			if f.calor[base+x] > 0 {
				f.calor[base+x]--
			}
		}
	}
}

// Linhas desenha a fogueira.
func (f *Fogo) Linhas() []string {
	out := make([]string, f.altura)
	for y := 0; y < f.altura; y++ {
		var b strings.Builder
		for x := 0; x < f.largura; x++ {
			b.WriteByte(Rampa[f.calor[y*f.largura+x]])
		}
		out[y] = b.String()
	}
	return out
}

// Sobre escreve linhas por cima do fogo, centralizadas. Espaço no texto deixa
// a chama aparecer, então a moldura de um painel some no meio do fogo mas os
// números continuam legíveis.
func Sobre(fundo, linhas []string) []string {
	out := append([]string(nil), fundo...)
	inicio := (len(fundo) - len(linhas)) / 2
	for i, l := range linhas {
		y := inicio + i
		if y < 0 || y >= len(out) {
			continue
		}
		base := []rune(out[y])
		x := (len(base) - len([]rune(l))) / 2
		for k, ch := range []rune(l) {
			if ch == ' ' {
				continue
			}
			if p := x + k; p >= 0 && p < len(base) {
				base[p] = ch
			}
		}
		out[y] = string(base)
	}
	return out
}
