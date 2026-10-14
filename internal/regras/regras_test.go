package regras

import "testing"

func TestIdentifica(t *testing.T) {
	casos := []struct {
		nome  string
		mao   Mao
		quero Combinacao
	}{
		{"cinco iguais", Mao{4, 4, 4, 4, 4}, Yatzee},
		{"quatro iguais", Mao{2, 2, 2, 2, 5}, Quadra},
		{"trinca mais par", Mao{3, 3, 3, 6, 6}, FullHouse},
		{"corrida de baixo", Mao{1, 2, 3, 4, 5}, Sequencia},
		{"corrida de cima", Mao{2, 3, 4, 5, 6}, Sequencia},
		{"corrida fora de ordem", Mao{5, 3, 1, 4, 2}, Sequencia},
		{"tres iguais", Mao{6, 6, 6, 1, 3}, Trinca},
		{"dois iguais", Mao{5, 5, 1, 2, 3}, Par},
		{"dois pares vira par", Mao{1, 1, 2, 2, 5}, Par},
		{"nada", Mao{1, 2, 3, 4, 6}, Nada},
	}
	for _, c := range casos {
		if got := Identifica(c.mao); got != c.quero {
			t.Errorf("%s %v: identificou %s, quero %s", c.nome, c.mao, got.Nome(), c.quero.Nome())
		}
	}
}

// A ordem dos testes de combinação é o que impede uma quadra de ser lida como
// trinca, e um full house também. Sem ela toda mão grande viraria a menor.
func TestCombinacaoMaisForteGanha(t *testing.T) {
	if got := Identifica(Mao{2, 2, 2, 2, 5}); got == Trinca {
		t.Error("quadra foi lida como trinca: a ordem dos testes está errada")
	}
	if got := Identifica(Mao{3, 3, 3, 6, 6}); got == Trinca {
		t.Error("full house foi lido como trinca")
	}
	if got := Identifica(Mao{4, 4, 4, 4, 4}); got == Quadra {
		t.Error("yatzee foi lido como quadra")
	}
}

// Uma corrida não pode ter dado repetido, senão 1-2-3-4-4 passaria.
func TestSequenciaNaoAceitaRepetido(t *testing.T) {
	naoSao := []Mao{
		{1, 2, 3, 4, 4},
		{2, 3, 4, 5, 5},
		{1, 2, 3, 5, 6},
		{1, 3, 4, 5, 6},
	}
	for _, m := range naoSao {
		if Identifica(m) == Sequencia {
			t.Errorf("%v foi lida como sequência", m)
		}
	}
}

func TestBaseDeCadaCombinacao(t *testing.T) {
	casos := []struct {
		c      Combinacao
		fichas int
		mult   int
	}{
		{Par, 4, 2},
		{Trinca, 5, 4},
		{Sequencia, 10, 4},
		{FullHouse, 5, 5},
		{Quadra, 9, 5},
		{Yatzee, 10, 5},
		{Nada, 0, 1},
	}
	for _, c := range casos {
		f, m := c.c.Base()
		if f != c.fichas || m != c.mult {
			t.Errorf("%s: base %d x %d, quero %d x %d", c.c.Nome(), f, m, c.fichas, c.mult)
		}
	}
}

// A soma dos dados entra nas fichas de toda jogada: é o que faz o valor dos
// dados importar, e não só a combinação.
func TestAvaliaSomaOsDadosNasFichas(t *testing.T) {
	casos := []struct {
		nome       string
		mao        Mao
		quemFichas int
		queroMult  int
		queroTotal int
	}{
		{"yatzee de seis", Mao{6, 6, 6, 6, 6}, 10 + 30, 5, 200},
		{"yatzee de um", Mao{1, 1, 1, 1, 1}, 10 + 5, 5, 75},
		{"trinca de tres", Mao{3, 3, 3, 2, 6}, 5 + 17, 4, 88},
		{"par de um", Mao{1, 1, 2, 3, 4}, 4 + 11, 2, 30},
		{"corrida de baixo", Mao{1, 2, 3, 4, 5}, 10 + 15, 4, 100},
		{"nada", Mao{1, 2, 3, 4, 6}, 0 + 16, 1, 16},
	}
	for _, c := range casos {
		p := Avalia(c.mao)
		if p.Fichas != c.quemFichas {
			t.Errorf("%s: fichas %d, quero %d", c.nome, p.Fichas, c.quemFichas)
		}
		if p.Mult != c.queroMult {
			t.Errorf("%s: mult %d, quero %d", c.nome, p.Mult, c.queroMult)
		}
		if got := p.Total(); got != c.queroTotal {
			t.Errorf("%s: total %d, quero %d", c.nome, got, c.queroTotal)
		}
	}
}

// Mesma combinação com dados maiores tem de valer mais: é a consequência da
// regra da soma, e o que diferencia um yatzee de seis de um de um.
func TestDadosMaioresValemMais(t *testing.T) {
	baixo := Avalia(Mao{2, 2, 2, 2, 2}).Total()
	alto := Avalia(Mao{5, 5, 5, 5, 5}).Total()
	if alto <= baixo {
		t.Errorf("yatzee de cinco (%d) não vale mais que de dois (%d)", alto, baixo)
	}
}

func TestSoma(t *testing.T) {
	if got := Soma(Mao{1, 2, 3, 4, 5}); got != 15 {
		t.Errorf("Soma = %d, quero 15", got)
	}
}

func TestNomeDeTodasAsCombinacoes(t *testing.T) {
	for c := Nada; c <= Yatzee; c++ {
		if c.Nome() == "" {
			t.Errorf("combinação %d sem nome", c)
		}
	}
}
