// Package regras tem as regras do jogo em Go puro: sem I/O, sem terminal.
// É o que permite testar pontuação sem montar tela nenhuma.
package regras

import "sort"

// DadosPorMao é quantos dados uma jogada usa.
const DadosPorMao = 5

// Mao é uma jogada: cinco dados de 1 a 6.
type Mao [DadosPorMao]int

// Combinacao é o desenho que a mão forma.
type Combinacao int

const (
	// Nada é a mão que não forma nem um par. Ainda pontua, porque a soma dos
	// dados entra em toda jogada.
	Nada Combinacao = iota
	Par
	Trinca
	Sequencia
	FullHouse
	Quadra
	Yatzee
)

// Nome é como a combinação aparece na tela.
func (c Combinacao) Nome() string {
	switch c {
	case Par:
		return "PAR"
	case Trinca:
		return "TRINCA"
	case Sequencia:
		return "SEQUENCIA"
	case FullHouse:
		return "FULL HOUSE"
	case Quadra:
		return "QUADRA"
	case Yatzee:
		return "YATZEE"
	default:
		return "NADA"
	}
}

// Base são as fichas e o multiplicador da combinação, antes de somar os dados.
func (c Combinacao) Base() (fichas, mult int) {
	switch c {
	case Par:
		return 4, 2
	case Trinca:
		return 5, 4
	case Sequencia:
		return 10, 4
	case FullHouse:
		return 5, 5
	case Quadra:
		return 9, 5
	case Yatzee:
		return 10, 5
	default:
		return 0, 1
	}
}

// Identifica diz qual combinação a mão forma. A ordem dos testes importa: uma
// quadra contém uma trinca, e um full house também, então o mais forte tem de
// ser verificado primeiro ou toda mão grande viraria trinca.
func Identifica(m Mao) Combinacao {
	contagem := map[int]int{}
	for _, d := range m {
		contagem[d]++
	}

	var repeticoes []int
	for _, n := range contagem {
		repeticoes = append(repeticoes, n)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(repeticoes)))

	switch {
	case repeticoes[0] == 5:
		return Yatzee
	case repeticoes[0] == 4:
		return Quadra
	case repeticoes[0] == 3 && len(repeticoes) > 1 && repeticoes[1] == 2:
		return FullHouse
	case sequencia(m):
		return Sequencia
	case repeticoes[0] == 3:
		return Trinca
	case repeticoes[0] == 2:
		return Par
	default:
		return Nada
	}
}

// sequencia reconhece as duas únicas corridas possíveis com cinco dados:
// 1-2-3-4-5 e 2-3-4-5-6.
func sequencia(m Mao) bool {
	visto := map[int]bool{}
	for _, d := range m {
		if visto[d] {
			return false // repetido não forma corrida
		}
		visto[d] = true
	}
	return (visto[1] && visto[2] && visto[3] && visto[4] && visto[5]) ||
		(visto[2] && visto[3] && visto[4] && visto[5] && visto[6])
}

// Soma é o total dos dados da mão.
func Soma(m Mao) int {
	total := 0
	for _, d := range m {
		total += d
	}
	return total
}

// Pontuacao é o resultado de avaliar uma mão.
type Pontuacao struct {
	Combinacao Combinacao
	Fichas     int // base da combinação mais a soma dos dados
	Mult       int
}

// Total é quanto a jogada vale.
func (p Pontuacao) Total() int { return p.Fichas * p.Mult }

// Avalia pontua uma mão. A soma dos dados entra nas fichas de toda jogada, o
// que faz o valor dos dados importar: um Yatzee de seis vale muito mais que um
// de um, mesmo sendo a mesma combinação.
func Avalia(m Mao) Pontuacao {
	c := Identifica(m)
	fichas, mult := c.Base()
	return Pontuacao{
		Combinacao: c,
		Fichas:     fichas + Soma(m),
		Mult:       mult,
	}
}
