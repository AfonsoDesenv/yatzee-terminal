package game

import (
	"strings"
	"testing"

	"github.com/afonsodesenv/yatzee-terminal/internal/regras"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

// naLoja leva a tela até a loja, com saldo suficiente para comprar.
func naLoja(t *testing.T, moedas int) Screen {
	t.Helper()
	s := venceAgora(New())
	s.partida.Moedas = moedas
	s, _ = tecla(s, "space") // vitória -> loja
	if s.Phase() != PhaseLoja {
		t.Fatalf("fase = %v, quero PhaseLoja", s.Phase())
	}
	return s
}

func TestALojaOfereceTresRunasDistintas(t *testing.T) {
	s := naLoja(t, 50)

	if got := len(s.vitrine); got != 3 {
		t.Fatalf("a vitrine tem %d runas, quero 3", got)
	}
	vistas := map[string]bool{}
	for _, r := range s.vitrine {
		if r.Letra == "" {
			t.Error("vitrine com espaço vazio logo ao abrir")
		}
		if vistas[r.Letra] {
			t.Errorf("a runa %s apareceu duas vezes na mesma vitrine", r.Letra)
		}
		vistas[r.Letra] = true
	}
}

// A loja abre em cima de uma runa, não do botão: quem acabou de vencer veio
// aqui para olhar a prateleira, e a caixa de texto já nasce dizendo algo.
func TestALojaAbreComAPrimeiraRunaEmFoco(t *testing.T) {
	s := naLoja(t, 50)
	if s.sel != focoPrimeiraRuna {
		t.Errorf("foco em %d ao abrir, quero a primeira runa (%d)", s.sel, focoPrimeiraRuna)
	}
	r, ok := s.emFoco()
	if !ok {
		t.Fatal("o foco inicial não está numa runa")
	}
	if r.Letra != s.vitrine[0].Letra {
		t.Errorf("o foco está na runa %s, quero a primeira da vitrine (%s)", r.Letra, s.vitrine[0].Letra)
	}
}

// As setas percorrem o botão e as três runas, dando a volta nas pontas: parar
// na borda obrigaria a refazer o caminho todo para chegar ao outro lado.
func TestAsSetasAndamPeloBotaoEPelasRunas(t *testing.T) {
	s := naLoja(t, 50)

	quero := []int{2, 3, focoContinuar, 1, 2}
	for i, esperado := range quero {
		s, _ = tecla(s, "right")
		if s.sel != esperado {
			t.Fatalf("%da seta para a direita parou em %d, quero %d", i+1, s.sel, esperado)
		}
	}

	s = naLoja(t, 50)
	for i, esperado := range []int{focoContinuar, 3, 2, 1} {
		s, _ = tecla(s, "left")
		if s.sel != esperado {
			t.Fatalf("%da seta para a esquerda parou em %d, quero %d", i+1, s.sel, esperado)
		}
	}
}

// Cima e baixo andam no mesmo anel: nenhuma seta fica morta na mão de quem
// estiver com os dedos do outro lado do teclado.
func TestCimaEBaixoAndamComoEsquerdaEDireita(t *testing.T) {
	comSetas := naLoja(t, 50)
	comSetas, _ = tecla(comSetas, "down")

	if comSetas.sel != focoPrimeiraRuna+1 {
		t.Errorf("baixo levou o foco a %d, quero %d", comSetas.sel, focoPrimeiraRuna+1)
	}
	comSetas, _ = tecla(comSetas, "up")
	if comSetas.sel != focoPrimeiraRuna {
		t.Errorf("cima não voltou o foco: está em %d", comSetas.sel)
	}
}

// O retângulo abaixo da prateleira descreve a runa em foco e só ela: era o
// ponto de trocar as três descrições soltas por uma caixa que acompanha.
func TestODescritivoSegueOFoco(t *testing.T) {
	s := naLoja(t, 50)

	for i, r := range s.vitrine {
		s.sel = focoPrimeiraRuna + i
		v := s.View()

		if !strings.Contains(v, r.Descricao()) {
			t.Errorf("com o foco na runa %s a loja não mostra %q", r.Letra, r.Descricao())
		}
		for j, outra := range s.vitrine {
			if j == i || outra.Descricao() == r.Descricao() {
				continue
			}
			if strings.Contains(v, outra.Descricao()) {
				t.Errorf("com o foco na runa %s a loja ainda mostra a descrição de %s", r.Letra, outra.Letra)
			}
		}
	}
}

// No botão a caixa fala do botão, e avisa qual é a meta que vem.
func TestODescritivoDoBotaoAvisaOProximoAlvo(t *testing.T) {
	s := naLoja(t, 50)
	s.sel = focoContinuar
	v := s.View()

	if !strings.Contains(v, "proximo nivel") {
		t.Error("a caixa não diz o que o botão faz")
	}
	if alvo := ui.Numero(s.partida.ProximoAlvo()); !strings.Contains(v, alvo) {
		t.Errorf("a caixa não mostra o próximo alvo (%s)", alvo)
	}
}

func TestEspacoCompraARunaEmFoco(t *testing.T) {
	s := naLoja(t, 50)
	s, _ = tecla(s, "right") // segunda runa
	escolhida := s.vitrine[1]

	s, _ = tecla(s, "space")

	if got := len(s.Partida().Runas); got != 1 {
		t.Fatalf("o jogador tem %d runas, quero 1", got)
	}
	if got := s.Partida().Runas[0].Letra; got != escolhida.Letra {
		t.Errorf("comprou a runa %s, esperava a que estava em foco (%s)", got, escolhida.Letra)
	}
	if got := s.Partida().Moedas; got != 50-PrecoRuna {
		t.Errorf("saldo = %d, quero %d", got, 50-PrecoRuna)
	}
}

// Comprada, some da prateleira: senão daria para levar a mesma runa duas
// vezes com um toque repetido.
func TestRunaCompradaSaiDaVitrine(t *testing.T) {
	s := naLoja(t, 50)
	s, _ = tecla(s, "space")
	if s.vitrine[0].Letra != "" {
		t.Error("a runa continua à venda depois de comprada")
	}

	s, _ = tecla(s, "space")
	if got := len(s.Partida().Runas); got != 1 {
		t.Errorf("comprou de novo no mesmo slot: %d runas", got)
	}
}

// O slot vendido continua ocupando lugar no anel, com a caixa dizendo por quê.
func TestOSlotVendidoSeAnuncia(t *testing.T) {
	s := naLoja(t, 50)
	s, _ = tecla(s, "space")

	v := s.View()
	if !strings.Contains(v, "vendida") {
		t.Error("a carta vendida não se identifica")
	}
	if !strings.Contains(v, "ja foi levada") {
		t.Error("a caixa não explica o slot vendido")
	}
}

func TestSemSaldoNaoCompra(t *testing.T) {
	s := naLoja(t, PrecoRuna-1)
	s, _ = tecla(s, "space")

	if got := len(s.Partida().Runas); got != 0 {
		t.Errorf("comprou sem saldo: %d runas", got)
	}
	if got := s.Partida().Moedas; got != PrecoRuna-1 {
		t.Errorf("saldo = %d, não devia mudar", got)
	}
}

// Sem o motivo na tela, apertar espaço e não acontecer nada parece bug.
func TestACaixaDizPorQueNaoDaPraComprar(t *testing.T) {
	semSaldo := naLoja(t, PrecoRuna-1)
	if v := semSaldo.View(); !strings.Contains(v, "voce tem so") {
		t.Error("a caixa não avisa que falta moeda")
	}

	cheio := naLoja(t, 50)
	for i := 0; i < regras.MaxRunas; i++ {
		cheio.partida = cheio.partida.Compra(regras.Catalogo[i], 0)
	}
	if v := cheio.View(); !strings.Contains(v, "ja estao ocupadas") {
		t.Error("a caixa não avisa que não cabe mais runa")
	}
}

// Cabem cinco runas; a sexta não entra nem com dinheiro sobrando.
func TestNaoPassaDoLimiteDeRunas(t *testing.T) {
	s := naLoja(t, 500)
	for i := 0; i < regras.MaxRunas; i++ {
		s.partida = s.partida.Compra(regras.Catalogo[i], 0)
	}

	s, _ = tecla(s, "space")
	if got := len(s.Partida().Runas); got != regras.MaxRunas {
		t.Errorf("passou do limite: %d runas, máximo %d", got, regras.MaxRunas)
	}
}

// A runa comprada participa da jogada seguinte.
func TestARunaCompradaEntraNaContagem(t *testing.T) {
	s := naLoja(t, 50)
	c, _ := regras.RunaPorLetra("C")
	s.partida = s.partida.Compra(c, 0)

	s, _ = tecla(s, "left")  // botão de continuar
	s, _ = tecla(s, "space") // próximo nível
	s.partida.Dados = regras.Mao{3, 3, 3, 6, 6}
	s, _ = tecla(s, "space") // envia

	var achou bool
	for _, p := range s.anim.passos {
		if strings.Contains(p.rotulo, "runa C") {
			achou = true
		}
	}
	if !achou {
		t.Error("a runa comprada não virou passo da contagem")
	}
}
