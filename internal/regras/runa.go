package regras

import "fmt"

// MaxRunas é quantas runas cabem numa partida.
const MaxRunas = 5

// efeito diz como a runa mexe na pontuação.
type efeito int

const (
	// fichasPorDado soma fichas uma vez por dado com o valor da runa.
	fichasPorDado efeito = iota
	// multPorDado soma ao multiplicador uma vez por dado com o valor da runa.
	// Conta por dado, não por mão: três dados com o valor somam três vezes.
	multPorDado
)

// Valores que as runas básicas somam.
const (
	FichasPorDadoDaRuna = 10
	MultDaRuna          = 2
)

// Runa é um coringa: dispara depois dos dados e mexe em fichas ou no
// multiplicador. Por enquanto só existe a família básica, uma por valor de
// dado, em duas variações.
type Runa struct {
	Letra  string
	efeito efeito
	valor  int // o valor de dado que a runa observa
}

// Descricao é o que a runa faz, em uma linha.
func (r Runa) Descricao() string {
	if r.efeito == fichasPorDado {
		return fmt.Sprintf("cada dado %d: +%d fichas", r.valor, FichasPorDadoDaRuna)
	}
	return fmt.Sprintf("cada dado %d: +%d mult", r.valor, MultDaRuna)
}

// Aplica devolve o que a runa acrescenta a esta mão. Fichas e multiplicador
// são somados, nunca multiplicados: é o que mantém a conta previsível quando
// várias runas disparam juntas.
//
// As duas famílias contam por dado, e não por mão: com três dados do valor da
// runa, ela soma três vezes. Vale para fichas e para multiplicador — é o que
// faz um yatzee render bem mais que um par com a mesma runa.
func (r Runa) Aplica(m Mao) (fichas, mult int) {
	f, mu := r.PorGatilho()
	vezes := len(r.Gatilhos(m))
	return vezes * f, vezes * mu
}

// Gatilhos são as posições dos dados que fazem a runa somar. Serve à tela, que
// conta um passo por gatilho e acende o dado de cada um; fatia vazia quer
// dizer que nenhum dado a dispara.
func (r Runa) Gatilhos(m Mao) []int {
	var out []int
	for i, d := range m {
		if d == r.valor {
			out = append(out, i)
		}
	}
	return out
}

// PorGatilho é o que a runa soma a cada dado que a dispara. Aplica é isto
// vezes o número de gatilhos, então a conta da tela, somada passo a passo,
// fecha com a da pontuação.
func (r Runa) PorGatilho() (fichas, mult int) {
	if r.efeito == fichasPorDado {
		return FichasPorDadoDaRuna, 0
	}
	return 0, MultDaRuna
}

// Dispara diz se a runa tem efeito nesta mão. Serve à tela: uma runa que não
// faz nada não vira passo na contagem.
func (r Runa) Dispara(m Mao) bool {
	f, mu := r.Aplica(m)
	return f > 0 || mu > 0
}

// Catalogo são as runas que existem no jogo. A ordem é a das letras, e é por
// ela que a loja sorteia.
var Catalogo = montaCatalogo()

func montaCatalogo() []Runa {
	letras := "ABCDEF"
	var out []Runa
	for i := 0; i < 6; i++ {
		out = append(out, Runa{Letra: string(letras[i]), efeito: fichasPorDado, valor: i + 1})
	}
	letras = "GHIJKL"
	for i := 0; i < 6; i++ {
		out = append(out, Runa{Letra: string(letras[i]), efeito: multPorDado, valor: i + 1})
	}
	return out
}

// RunaPorLetra acha uma runa do catálogo.
func RunaPorLetra(letra string) (Runa, bool) {
	for _, r := range Catalogo {
		if r.Letra == letra {
			return r, true
		}
	}
	return Runa{}, false
}
