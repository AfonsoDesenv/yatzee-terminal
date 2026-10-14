package regras

import (
	"math/rand/v2"
	"testing"
)

func rnd() *rand.Rand { return rand.New(rand.NewPCG(7, 13)) }

func TestNovaPartidaComecaCheia(t *testing.T) {
	p := NovaPartida(rnd())

	if p.Nivel != 1 {
		t.Errorf("nível = %d, quero 1", p.Nivel)
	}
	if p.Alvo != AlvoInicial {
		t.Errorf("alvo = %d, quero %d", p.Alvo, AlvoInicial)
	}
	if p.Maos != MaosPorRodada || p.Rerolls != RerollsPorRodada {
		t.Errorf("mãos/rerolls = %d/%d, quero %d/%d", p.Maos, p.Rerolls, MaosPorRodada, RerollsPorRodada)
	}
	for i, d := range p.Dados {
		if d < 1 || d > 6 {
			t.Errorf("dado %d fora de 1..6: %d", i, d)
		}
	}
}

func TestMarcaAlterna(t *testing.T) {
	p := NovaPartida(rnd())
	if p.AlgumMarcado() {
		t.Fatal("começou com dado marcado")
	}

	p = p.Marca(3)
	if !p.Marcados[2] {
		t.Error("marcar 3 não marcou o terceiro dado")
	}
	p = p.Marca(3)
	if p.Marcados[2] {
		t.Error("marcar de novo não desmarcou")
	}
}

func TestMarcaForaDoIntervaloNaoFazNada(t *testing.T) {
	p := NovaPartida(rnd())
	for _, n := range []int{0, -1, 6, 99} {
		if p.Marca(n).AlgumMarcado() {
			t.Errorf("Marca(%d) marcou algo", n)
		}
	}
}

// Gastar um reroll sem escolher dado nenhum desperdiçaria o recurso.
func TestNaoRerolaSemDadoMarcado(t *testing.T) {
	p := NovaPartida(rnd())
	if p.PodeRerolar() {
		t.Error("pode rerolar sem nada marcado")
	}
	antes := p.Rerolls
	if got := p.Rerola(rnd()).Rerolls; got != antes {
		t.Errorf("rerolls = %d depois de rerolar sem marcar, quero %d", got, antes)
	}
}

func TestRerolaTrocaSoOsMarcados(t *testing.T) {
	p := NovaPartida(rnd())
	antes := p.Dados

	p = p.Marca(1).Marca(2)
	// Semente diferente, para os dados novos terem chance real de mudar.
	p = p.Rerola(rand.New(rand.NewPCG(99, 3)))

	for i := 2; i < DadosPorMao; i++ {
		if p.Dados[i] != antes[i] {
			t.Errorf("dado %d mudou sem estar marcado: %d viva %d", i+1, antes[i], p.Dados[i])
		}
	}
	if p.Rerolls != RerollsPorRodada-1 {
		t.Errorf("rerolls = %d, quero %d", p.Rerolls, RerollsPorRodada-1)
	}
	if p.AlgumMarcado() {
		t.Error("as marcas não foram limpas depois do reroll")
	}
}

func TestRerolaAcabaESeRecusa(t *testing.T) {
	p := NovaPartida(rnd())
	for i := 0; i < RerollsPorRodada; i++ {
		p = p.Marca(1).Rerola(rnd())
	}
	if p.Rerolls != 0 {
		t.Fatalf("rerolls = %d depois de gastar todos", p.Rerolls)
	}
	if p.Marca(1).PodeRerolar() {
		t.Error("ainda deixa rerolar com o contador zerado")
	}
}

func TestEnviaPontuaEGastaUmaMao(t *testing.T) {
	p := NovaPartida(rnd())
	p.Dados = Mao{3, 3, 3, 6, 6} // full house, soma 21

	depois, pont := p.Envia()

	if pont.Combinacao != FullHouse {
		t.Errorf("combinação = %s, quero FULL HOUSE", pont.Combinacao.Nome())
	}
	if quero := int64((5 + 21) * 5); depois.Pontos != quero {
		t.Errorf("pontos = %d, quero %d", depois.Pontos, quero)
	}
	if depois.Maos != MaosPorRodada-1 {
		t.Errorf("mãos = %d, quero %d", depois.Maos, MaosPorRodada-1)
	}
}

func TestEnviaAcumulaEntreMaos(t *testing.T) {
	p := NovaPartida(rnd())
	p.Dados = Mao{1, 1, 2, 3, 4} // par, (4+11)*2 = 30

	p, _ = p.Envia()
	primeiro := p.Pontos
	p, _ = p.Envia()

	if p.Pontos != primeiro*2 {
		t.Errorf("pontos = %d depois de duas mãos iguais, quero %d", p.Pontos, primeiro*2)
	}
}

func TestNaoEnviaSemMao(t *testing.T) {
	p := NovaPartida(rnd())
	p.Maos = 0
	if p.PodeEnviar() {
		t.Error("deixa enviar sem mão disponível")
	}
	depois, _ := p.Envia()
	if depois.Pontos != p.Pontos {
		t.Error("enviar sem mão pontuou")
	}
}

func TestVenceuEPerdeu(t *testing.T) {
	casos := []struct {
		nome           string
		pontos         int64
		maos           int
		venceu, perdeu bool
	}{
		{"bateu a meta", AlvoInicial, 2, true, false},
		{"passou da meta", AlvoInicial + 500, 0, true, false},
		{"ainda tem mão", AlvoInicial - 1, 1, false, false},
		{"acabaram as mãos", AlvoInicial - 1, 0, false, true},
	}
	for _, c := range casos {
		p := Partida{Alvo: AlvoInicial, Pontos: c.pontos, Maos: c.maos}
		if got := p.Venceu(); got != c.venceu {
			t.Errorf("%s: Venceu = %v, quero %v", c.nome, got, c.venceu)
		}
		if got := p.Perdeu(); got != c.perdeu {
			t.Errorf("%s: Perdeu = %v, quero %v", c.nome, got, c.perdeu)
		}
	}
}

// Vencer e perder são exclusivos: bater a meta na última mão é vitória.
func TestVitoriaNaUltimaMaoNaoEDerrota(t *testing.T) {
	p := Partida{Alvo: 100, Pontos: 100, Maos: 0}
	if !p.Venceu() || p.Perdeu() {
		t.Error("bater a meta na última mão foi lido como derrota")
	}
}

func TestRecompensaPagaOQueSobrou(t *testing.T) {
	casos := []struct {
		maos, rerolls, quero int
	}{
		{0, 0, MoedasPorVitoria},
		{5, 3, MoedasPorVitoria + 5 + 3},
		{2, 1, MoedasPorVitoria + 2 + 1},
	}
	for _, c := range casos {
		p := Partida{Maos: c.maos, Rerolls: c.rerolls}
		if got := p.Recompensa(); got != c.quero {
			t.Errorf("Recompensa com %d mãos e %d rerolls = %d, quero %d", c.maos, c.rerolls, got, c.quero)
		}
	}
}

func TestProximoNivelDobraAMetaEEnchiTudo(t *testing.T) {
	p := NovaPartida(rnd())
	p.Pontos = p.Alvo
	p.Maos = 2
	p.Rerolls = 1
	alvoAntes := p.Alvo
	recompensa := p.Recompensa()

	p = p.Premia().ProximoNivel(rnd())

	if p.Nivel != 2 {
		t.Errorf("nível = %d, quero 2", p.Nivel)
	}
	if p.Alvo != alvoAntes*2 {
		t.Errorf("alvo = %d, quero %d", p.Alvo, alvoAntes*2)
	}
	if p.Pontos != 0 {
		t.Errorf("pontos = %d, quero zerar entre níveis", p.Pontos)
	}
	if p.Maos != MaosPorRodada || p.Rerolls != RerollsPorRodada {
		t.Errorf("mãos/rerolls = %d/%d, quero cheios", p.Maos, p.Rerolls)
	}
	if p.Moedas != recompensa {
		t.Errorf("moedas = %d, quero %d", p.Moedas, recompensa)
	}
}

// As moedas se acumulam entre níveis, senão a loja nunca teria o que vender.
// Premia credita; ProximoNivel nao mexe em moeda. Separados porque o saldo tem
// de estar certo quando a loja abre, e nao so depois dela.
func TestProximoNivelNaoCreditaSozinho(t *testing.T) {
	p := NovaPartida(rnd())
	p.Pontos = p.Alvo
	antes := p.Moedas
	if got := p.ProximoNivel(rnd()).Moedas; got != antes {
		t.Errorf("ProximoNivel creditou sozinho: moedas %d, quero %d", got, antes)
	}
	if got := p.Premia().Moedas; got == antes {
		t.Error("Premia nao creditou nada")
	}
}

func TestMoedasAcumulamEntreNiveis(t *testing.T) {
	p := NovaPartida(rnd())
	p.Pontos = p.Alvo
	p = p.Premia().ProximoNivel(rnd())
	primeiras := p.Moedas

	p.Pontos = p.Alvo
	p = p.Premia().ProximoNivel(rnd())

	if p.Moedas <= primeiras {
		t.Errorf("moedas = %d depois do segundo nível, não acumulou sobre %d", p.Moedas, primeiras)
	}
}
