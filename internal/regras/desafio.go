package regras

// Ajuste é uma condição que um desafio impõe à partida, aplicada antes da
// primeira rolagem.
//
// Um desafio é uma lista de ajustes, e não um punhado de campos opcionais:
// combinar condições passa a ser juntar ajustes, e um tipo de condição que
// ainda não existe é um Ajuste novo, sem nenhum switch para atualizar.
type Ajuste func(Partida) Partida

// ComMoedas começa a partida com um saldo.
func ComMoedas(n int) Ajuste {
	return func(p Partida) Partida {
		p.Moedas = n
		return p
	}
}

// ComRunas começa a partida com estas runas, na ordem dada. Letra fora do
// catálogo é ignorada, e o limite de runas continua valendo: um desafio mal
// escrito não deve derrubar o jogo nem furar a regra.
func ComRunas(letras ...string) Ajuste {
	return func(p Partida) Partida {
		runas := append([]Runa{}, p.Runas...)
		for _, l := range letras {
			r, ok := RunaPorLetra(l)
			if !ok || len(runas) >= MaxRunas {
				continue
			}
			runas = append(runas, r)
		}
		p.Runas = runas
		return p
	}
}

// SoComFaces limita os valores que um dado pode cair. Vale para a partida
// inteira — a primeira mão, os rerolls e os níveis seguintes —, porque é
// condição da partida e não da rodada.
func SoComFaces(faces ...int) Ajuste {
	return func(p Partida) Partida {
		p.faces = append([]int{}, faces...)
		return p
	}
}

// Desafio é uma partida com as condições mexidas.
type Desafio struct {
	Nome   string
	Resumo string
	// Ajustes são as condições, aplicadas na ordem.
	Ajustes []Ajuste
}

// Aplica impõe as condições do desafio a uma partida.
func (d Desafio) Aplica(p Partida) Partida {
	for _, a := range d.Ajustes {
		p = a(p)
	}
	return p
}

// Desafios é o catálogo. Acrescentar um desafio é acrescentar uma entrada
// aqui; a tela que os lista e a raiz que os inicia leem desta fatia, então
// nenhuma das duas precisa mudar.
var Desafios = []Desafio{
	{
		Nome:    "BOLSO CHEIO",
		Resumo:  "comeca a partida com 15 moedas",
		Ajustes: []Ajuste{ComMoedas(15)},
	},
	{
		Nome:    "HERANCA",
		Resumo:  "comeca com as runas F e L ja em jogo",
		Ajustes: []Ajuste{ComRunas("F", "L")},
	},
	{
		Nome:    "DADOS VICIADOS",
		Resumo:  "os dados so caem 4, 5 ou 6 - sequencia fica impossivel",
		Ajustes: []Ajuste{SoComFaces(4, 5, 6)},
	},
}
