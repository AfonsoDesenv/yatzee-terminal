package fogo

import (
	"math/rand/v2"
	"strings"
	"testing"
)

func novo(t *testing.T, w, h int) *Fogo {
	t.Helper()
	return Novo(w, h, rand.New(rand.NewPCG(7, 13)))
}

func TestTamanhoDoDesenho(t *testing.T) {
	f := novo(t, 20, 8)
	l := f.Linhas()
	if len(l) != 8 {
		t.Fatalf("desenhou %d linhas, quero 8", len(l))
	}
	for i, linha := range l {
		if len([]rune(linha)) != 20 {
			t.Errorf("linha %d tem %d colunas, quero 20", i, len([]rune(linha)))
		}
	}
}

func TestAcendeNaBase(t *testing.T) {
	f := novo(t, 10, 5)
	l := f.Linhas()

	base := l[len(l)-1]
	if strings.Trim(base, string(Rampa[len(Rampa)-1])) != "" {
		t.Errorf("a base não está toda em brasa: %q", base)
	}
	if strings.TrimSpace(l[0]) != "" {
		t.Errorf("o topo já começa quente: %q", l[0])
	}
}

// A chama tem que subir: depois de alguns passos, o topo deixa de estar vazio.
func TestOCalorSobe(t *testing.T) {
	f := novo(t, 24, 8)
	for i := 0; i < 12; i++ {
		f.Passo()
	}
	if strings.TrimSpace(f.Linhas()[0]) == "" {
		t.Error("depois de 12 passos o topo continua frio: a chama não está subindo")
	}
}

// Enquanto acesa a fogueira não morre sozinha, senão a comemoração apagaria no
// meio.
func TestAcesaNaoMorre(t *testing.T) {
	f := novo(t, 24, 8)
	for i := 0; i < 200; i++ {
		f.Passo()
	}
	if !f.Vivo() {
		t.Error("a fogueira acesa morreu sozinha")
	}
}

// Depois de Apagar ela tem que acabar, senão a animação nunca termina e a tela
// fica presa no estado de comemoração.
func TestApagarMataAFogueira(t *testing.T) {
	f := novo(t, 24, 8)
	for i := 0; i < 10; i++ {
		f.Passo()
	}
	f.Apagar()

	for i := 0; i < 500 && f.Vivo(); i++ {
		f.Passo()
	}
	if f.Vivo() {
		t.Error("depois de Apagar e 500 passos a fogueira continua viva")
	}
	if strings.TrimSpace(strings.Join(f.Linhas(), "")) != "" {
		t.Error("sobrou calor desenhado depois da fogueira morrer")
	}
}

// Nenhuma célula pode esquentar além da brasa: um índice fora da rampa seria
// pânico na hora de desenhar.
func TestCalorNuncaPassaDaRampa(t *testing.T) {
	f := novo(t, 30, 10)
	for i := 0; i < 100; i++ {
		f.Passo()
		for _, c := range f.calor {
			if c < 0 || c >= len(Rampa) {
				t.Fatalf("calor %d fora da rampa (0..%d)", c, len(Rampa)-1)
			}
		}
	}
}

// Mesma semente, mesma chama: é o que torna a animação testável.
func TestMesmaSementeMesmaChama(t *testing.T) {
	a := Novo(20, 6, rand.New(rand.NewPCG(1, 2)))
	b := Novo(20, 6, rand.New(rand.NewPCG(1, 2)))
	for i := 0; i < 20; i++ {
		a.Passo()
		b.Passo()
	}
	if strings.Join(a.Linhas(), "\n") != strings.Join(b.Linhas(), "\n") {
		t.Error("duas fogueiras com a mesma semente divergiram")
	}
}

func TestSobreMantemOTextoLegivel(t *testing.T) {
	f := novo(t, 30, 9)
	for i := 0; i < 20; i++ {
		f.Passo()
	}
	painel := []string{"| ALVO   2.000 |", "| VOCE   2.340 |"}
	saida := strings.Join(Sobre(f.Linhas(), painel), "\n")

	for _, termo := range []string{"ALVO", "2.000", "VOCE", "2.340"} {
		if !strings.Contains(saida, termo) {
			t.Errorf("o fogo engoliu %q:\n%s", termo, saida)
		}
	}
}

// Espaço no texto sobreposto deixa a chama passar: sem isso o painel vira um
// retângulo opaco cravado no meio do fogo.
func TestSobreDeixaAChamaPassarNoEspaco(t *testing.T) {
	fundo := []string{"@@@@@@@@@@", "@@@@@@@@@@", "@@@@@@@@@@"}
	saida := Sobre(fundo, []string{"A A"})
	meio := saida[1]
	if !strings.Contains(meio, "@") {
		t.Errorf("o espaço do texto apagou a chama: %q", meio)
	}
}
