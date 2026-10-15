package regras

import (
	"math/rand/v2"
	"testing"
)

func gerador() *rand.Rand { return rand.New(rand.NewPCG(1, 2)) }

// porLetra acha um desafio do catálogo pelo nome.
func desafioPorNome(t *testing.T, nome string) Desafio {
	t.Helper()
	for _, d := range Desafios {
		if d.Nome == nome {
			return d
		}
	}
	t.Fatalf("desafio %q não existe no catálogo", nome)
	return Desafio{}
}

// Sem desafio, a partida é a de sempre: nenhum ajuste vaza para o jogo normal.
func TestPartidaNormalNaoTemCondicaoNenhuma(t *testing.T) {
	p := NovaPartida(gerador())

	if p.Moedas != 0 {
		t.Errorf("moedas = %d, quero 0", p.Moedas)
	}
	if len(p.Runas) != 0 {
		t.Errorf("%d runas, quero nenhuma", len(p.Runas))
	}
	if len(p.faces) != 0 {
		t.Errorf("faces = %v, quero vazia (todas liberadas)", p.faces)
	}
}

// Desafio 1: começa com 15 moedas, prontas para gastar na loja.
func TestDesafioDeMoedasComecaComSaldo(t *testing.T) {
	p := NovaPartidaDesafio(desafioPorNome(t, "BOLSO CHEIO"), gerador())

	if p.Moedas != 15 {
		t.Errorf("moedas = %d, quero 15", p.Moedas)
	}
	// E o saldo serve de verdade: compra na loja sai com ele.
	if !p.PodeComprar(5) {
		t.Error("nao da para comprar com 15 moedas")
	}
}

// Desafio 2: começa com as runas F e L já em jogo.
func TestDesafioDeRunasComecaComElasEmJogo(t *testing.T) {
	p := NovaPartidaDesafio(desafioPorNome(t, "HERANCA"), gerador())

	if len(p.Runas) != 2 {
		t.Fatalf("%d runas, quero 2", len(p.Runas))
	}
	if a, b := p.Runas[0].Letra, p.Runas[1].Letra; a != "F" || b != "L" {
		t.Errorf("runas %s e %s, quero F e L", a, b)
	}

	// E elas pontuam: F soma fichas por dado 6, L soma mult por dado 6.
	p.Dados = Mao{6, 6, 6, 1, 2}
	pont := AvaliaCom(p.Dados, p.Runas)
	if quero := (5 + 21 + 3*FichasPorDadoDaRuna) * (4 + 3*MultDaRuna); pont.Total() != quero {
		t.Errorf("trinca de 6 com as runas do desafio = %d, quero %d", pont.Total(), quero)
	}
}

// Desafio 3: nenhum dado cai 1, 2 ou 3 — e vale a partida inteira, não só a
// primeira mão.
func TestDesafioDeFacesValeNaPartidaInteira(t *testing.T) {
	d := desafioPorNome(t, "DADOS VICIADOS")
	rnd := gerador()
	p := NovaPartidaDesafio(d, rnd)

	confere := func(onde string, m Mao) {
		t.Helper()
		for i, v := range m {
			if v < 4 || v > 6 {
				t.Errorf("%s: dado %d caiu %d, so pode 4, 5 ou 6", onde, i+1, v)
			}
		}
	}
	confere("primeira mao", p.Dados)

	// Muitas rolagens, para o sorteio nao passar por acidente.
	for n := 0; n < 200; n++ {
		p = p.ProximaMao(rnd)
		confere("proxima mao", p.Dados)
	}

	// Reroll: marca todos e rerola.
	for i := 1; i <= DadosPorMao; i++ {
		p = p.Marca(i)
	}
	p.Rerolls = RerollsPorRodada
	for n := 0; n < 50; n++ {
		p.Rerolls = RerollsPorRodada
		for i := 1; i <= DadosPorMao; i++ {
			p = p.Marca(i)
		}
		p = p.Rerola(rnd)
		confere("reroll", p.Dados)
	}

	// E nos niveis seguintes, que e onde "durante toda a partida" pega.
	for n := 0; n < 50; n++ {
		p = p.ProximoNivel(rnd)
		confere("proximo nivel", p.Dados)
	}
}

// As três faces aparecem todas: a restrição estreita o dado, não o trava num
// valor só.
func TestODadoViciadoAindaVariaEntreAsFaces(t *testing.T) {
	rnd := gerador()
	p := NovaPartidaDesafio(desafioPorNome(t, "DADOS VICIADOS"), rnd)

	vistas := map[int]bool{}
	for n := 0; n < 200; n++ {
		p = p.ProximaMao(rnd)
		for _, v := range p.Dados {
			vistas[v] = true
		}
	}
	for _, f := range []int{4, 5, 6} {
		if !vistas[f] {
			t.Errorf("a face %d nunca saiu em 1000 dados", f)
		}
	}
}

// Um desafio é uma lista de ajustes, então condições se combinam sem código
// novo. É o que torna barato acrescentar desafio depois.
func TestOsAjustesSeCombinam(t *testing.T) {
	d := Desafio{
		Nome:    "TUDO JUNTO",
		Ajustes: []Ajuste{ComMoedas(7), ComRunas("A", "B"), SoComFaces(6)},
	}
	p := NovaPartidaDesafio(d, gerador())

	if p.Moedas != 7 {
		t.Errorf("moedas = %d, quero 7", p.Moedas)
	}
	if len(p.Runas) != 2 {
		t.Errorf("%d runas, quero 2", len(p.Runas))
	}
	if p.Dados != (Mao{6, 6, 6, 6, 6}) {
		t.Errorf("dados = %v, quero cinco seis", p.Dados)
	}
}

// Desafio mal escrito não derruba o jogo nem fura o limite de runas.
func TestDesafioMalEscritoNaoQuebraNada(t *testing.T) {
	d := Desafio{Ajustes: []Ajuste{ComRunas("Z", "nao existe", "A")}}
	p := NovaPartidaDesafio(d, gerador())
	if len(p.Runas) != 1 || p.Runas[0].Letra != "A" {
		t.Errorf("runas = %v, quero so a A (as invalidas sao ignoradas)", p.Runas)
	}

	muitas := make([]string, 0, 12)
	for _, r := range Catalogo {
		muitas = append(muitas, r.Letra)
	}
	cheio := NovaPartidaDesafio(Desafio{Ajustes: []Ajuste{ComRunas(muitas...)}}, gerador())
	if len(cheio.Runas) != MaxRunas {
		t.Errorf("%d runas, quero no maximo %d", len(cheio.Runas), MaxRunas)
	}
}

// O catálogo tem os três desafios, cada um com nome e resumo: a tela lista
// direto dele, e uma entrada sem texto apareceria em branco.
func TestOCatalogoEstaCompleto(t *testing.T) {
	if len(Desafios) != 3 {
		t.Errorf("%d desafios no catalogo, quero 3", len(Desafios))
	}
	vistos := map[string]bool{}
	for i, d := range Desafios {
		if d.Nome == "" {
			t.Errorf("desafio %d sem nome", i)
		}
		if d.Resumo == "" {
			t.Errorf("desafio %s sem resumo", d.Nome)
		}
		if len(d.Ajustes) == 0 {
			t.Errorf("desafio %s nao muda condicao nenhuma", d.Nome)
		}
		if vistos[d.Nome] {
			t.Errorf("desafio %s aparece duas vezes", d.Nome)
		}
		vistos[d.Nome] = true
	}
}

// Aplicar um desafio não mexe na partida que entrou: Partida é valor, e um
// ajuste que escrevesse no original vazaria para quem guardou o estado antes.
func TestAplicarNaoMexeNaPartidaQueEntrou(t *testing.T) {
	antes := base()
	antes.Runas = []Runa{Catalogo[0]}

	depois := desafioPorNome(t, "HERANCA").Aplica(antes)

	if len(antes.Runas) != 1 {
		t.Errorf("a partida de entrada ficou com %d runas, devia continuar com 1", len(antes.Runas))
	}
	if len(depois.Runas) != 3 {
		t.Errorf("a partida do desafio tem %d runas, quero 3 (a de antes mais duas)", len(depois.Runas))
	}
}
