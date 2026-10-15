package regras

import "math/rand/v2"

// Padrões de uma rodada.
const (
	MaosPorRodada    = 5
	RerollsPorRodada = 3

	// AlvoInicial é a meta do primeiro nível; cada nível dobra.
	AlvoInicial = 300

	// Recompensas de vitória, em moedas.
	MoedasPorVitoria        = 3
	MoedasPorMaoSobrando    = 1
	MoedasPorRerollSobrando = 1
)

// Partida é o estado de uma rodada em andamento. É um valor: cada método
// devolve a partida seguinte em vez de mexer na atual, o que deixa a tela
// guardar o estado anterior enquanto anima a transição.
type Partida struct {
	Nivel   int
	Alvo    int64
	Pontos  int64
	Maos    int // quantas ainda pode enviar
	Rerolls int
	Moedas  int

	Dados    Mao
	Marcados [DadosPorMao]bool

	// Runas em jogo, na ordem em que disparam.
	Runas []Runa

	// faces são os valores que um dado pode cair. Vazia quer dizer de 1 a 6;
	// quem estreita é um desafio, e a restrição acompanha a partida inteira
	// em vez de valer só na primeira mão.
	faces []int
}

// NovaPartida começa no primeiro nível, com os dados já rolados.
func NovaPartida(rnd *rand.Rand) Partida { return base().rolaTodos(rnd) }

// NovaPartidaDesafio começa uma partida com as condições de um desafio. Os
// ajustes entram antes da primeira rolagem: um desafio que estreita os dados
// tem de valer já na mão com que a partida abre.
func NovaPartidaDesafio(d Desafio, rnd *rand.Rand) Partida {
	return d.Aplica(base()).rolaTodos(rnd)
}

// base é a partida do primeiro nível, ainda sem dados. Separada da rolagem
// para o desafio poder entrar no meio.
func base() Partida {
	return Partida{
		Nivel:   1,
		Alvo:    AlvoInicial,
		Maos:    MaosPorRodada,
		Rerolls: RerollsPorRodada,
	}
}

// rola sorteia um dado entre as faces permitidas.
func (p Partida) rola(rnd *rand.Rand) int {
	if len(p.faces) == 0 {
		return rnd.IntN(6) + 1
	}
	return p.faces[rnd.IntN(len(p.faces))]
}

func (p Partida) rolaTodos(rnd *rand.Rand) Partida {
	for i := range p.Dados {
		p.Dados[i] = p.rola(rnd)
	}
	p.Marcados = [DadosPorMao]bool{}
	return p
}

// Marca alterna um dado para o reroll. A seleção é por número: o jogador
// digita 1, 3 e 5 e rerola os três de uma vez.
func (p Partida) Marca(n int) Partida {
	if n < 1 || n > DadosPorMao {
		return p
	}
	p.Marcados[n-1] = !p.Marcados[n-1]
	return p
}

// AlgumMarcado diz se há dado escolhido para rerolar.
func (p Partida) AlgumMarcado() bool {
	for _, m := range p.Marcados {
		if m {
			return true
		}
	}
	return false
}

// PodeRerolar exige reroll disponível e pelo menos um dado escolhido: gastar um
// reroll sem trocar nada só desperdiçaria o recurso.
func (p Partida) PodeRerolar() bool {
	return p.Rerolls > 0 && p.AlgumMarcado()
}

// Rerola troca os dados marcados e gasta um reroll.
func (p Partida) Rerola(rnd *rand.Rand) Partida {
	if !p.PodeRerolar() {
		return p
	}
	for i, marcado := range p.Marcados {
		if marcado {
			p.Dados[i] = p.rola(rnd)
		}
	}
	p.Marcados = [DadosPorMao]bool{}
	p.Rerolls--
	return p
}

// PodeEnviar exige mão disponível.
func (p Partida) PodeEnviar() bool { return p.Maos > 0 }

// Envia pontua a mão atual, gasta uma mão e devolve a partida seguinte junto
// com o que a jogada valeu. Os dados não são rolados de novo aqui: quem decide
// isso é a tela, depois de mostrar a contagem.
func (p Partida) Envia() (Partida, Pontuacao) {
	if !p.PodeEnviar() {
		return p, Pontuacao{}
	}
	pont := AvaliaCom(p.Dados, p.Runas)
	p.Pontos += int64(pont.Total())
	p.Maos--
	return p, pont
}

// ProximaMao rola os dados de novo para a mão seguinte.
func (p Partida) ProximaMao(rnd *rand.Rand) Partida { return p.rolaTodos(rnd) }

// Venceu diz se a meta foi batida.
func (p Partida) Venceu() bool { return p.Pontos >= p.Alvo }

// Perdeu diz se acabaram as mãos sem bater a meta.
func (p Partida) Perdeu() bool { return !p.Venceu() && p.Maos <= 0 }

// Recompensa é o que a vitória paga: um valor fixo mais o que sobrou de mãos e
// rerolls, o que premia quem bate a meta com folga.
func (p Partida) Recompensa() int {
	return MoedasPorVitoria +
		p.Maos*MoedasPorMaoSobrando +
		p.Rerolls*MoedasPorRerollSobrando
}

// Premia credita a recompensa da vitória. Fica separado de ProximoNivel
// porque as moedas entram quando a vitória é anunciada, e não quando o jogador
// sai da loja: é o saldo que a loja usa.
func (p Partida) Premia() Partida {
	p.Moedas += p.Recompensa()
	return p
}

// PodeComprar diz se cabe mais uma runa e o saldo cobre o preço.
func (p Partida) PodeComprar(preco int) bool {
	return len(p.Runas) < MaxRunas && p.Moedas >= preco
}

// Compra guarda a runa e desconta o preço.
func (p Partida) Compra(r Runa, preco int) Partida {
	if !p.PodeComprar(preco) {
		return p
	}
	p.Moedas -= preco
	p.Runas = append(append([]Runa{}, p.Runas...), r)
	return p
}

// ProximoAlvo é a meta do nível seguinte. Existe para a loja poder avisar o
// que vem pela frente sem repetir a regra do dobro por conta própria.
func (p Partida) ProximoAlvo() int64 { return p.Alvo * 2 }

// ProximoNivel dobra a meta e devolve tudo cheio de novo. Não mexe nas moedas:
// quem credita é Premia, na vitória.
func (p Partida) ProximoNivel(rnd *rand.Rand) Partida {
	p.Nivel++
	p.Alvo = p.ProximoAlvo()
	p.Pontos = 0
	p.Maos = MaosPorRodada
	p.Rerolls = RerollsPorRodada
	return p.rolaTodos(rnd)
}
