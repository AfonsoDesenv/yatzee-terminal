package ui

import (
	"strings"
	"testing"
)

func TestDadoTemOTamanhoQueAMesaAssume(t *testing.T) {
	for v := 1; v <= 6; v++ {
		d := Dado(v)
		if len(d) != AlturaDado {
			t.Errorf("dado %d tem %d linhas, quero %d", v, len(d), AlturaDado)
		}
		for i, l := range d {
			if Largura(l) != LarguraDado {
				t.Errorf("dado %d, linha %d: largura %d, quero %d (%q)", v, i, Largura(l), LarguraDado, l)
			}
		}
	}
}

// A contagem de pinos é a propriedade que importa: um erro na grade produz um
// dado que parece certo mas mostra o valor errado.
func TestDadoMostraOValorCerto(t *testing.T) {
	for v := 1; v <= 6; v++ {
		got := strings.Count(strings.Join(Dado(v), ""), "o")
		if got != v {
			t.Errorf("dado %d desenhou %d pinos", v, got)
		}
	}
}

// Ímpar tem pino no centro, par não. É o que distingue 5 de 4 e 3 de 2.
func TestDadoImparTemCentro(t *testing.T) {
	for v := 1; v <= 6; v++ {
		centro := []rune(Dado(v)[2])[4] == 'o'
		quero := v%2 == 1
		if centro != quero {
			t.Errorf("dado %d: centro=%v, quero %v", v, centro, quero)
		}
	}
}

func TestMarcaDadoOcupaALarguraDoDado(t *testing.T) {
	for _, marcado := range []bool{true, false} {
		if got := Largura(MarcaDado(marcado)); got != LarguraDado {
			t.Errorf("MarcaDado(%v) tem largura %d, quero %d", marcado, got, LarguraDado)
		}
	}
	if MarcaDado(true) == MarcaDado(false) {
		t.Error("dado marcado e não marcado desenham igual: a marcação não aparece")
	}
}

func TestDadoArteFechaOCubo(t *testing.T) {
	arte := DadoArte(5, 3, 2)
	if len(arte) < 10 {
		t.Fatalf("arte com %d linhas, esperava o cubo inteiro", len(arte))
	}

	// O topo é a face mais clara e a lateral a mais escura: se a escala de
	// brilho se inverter ou achatar, o cubo deixa de ler como volume.
	junto := strings.Join(arte, "\n")
	for _, ch := range []string{"%", "+", ":"} {
		if !strings.Contains(junto, ch) {
			t.Errorf("a arte não usa %q: as três faces não estão se distinguindo", ch)
		}
	}
}

func TestDadoArteMostraOsPinosDaFrente(t *testing.T) {
	// Um pino é um buraco escuro no meio da face clara. Com valor 1 há um só,
	// com 6 há seis: a quantidade de espaços internos tem que crescer.
	um := strings.Count(strings.Join(DadoArte(1, 1, 1), ""), " ")
	seis := strings.Count(strings.Join(DadoArte(6, 1, 1), ""), " ")
	if seis <= um {
		t.Errorf("face 6 tem %d espaços e face 1 tem %d: os pinos não estão sendo vazados", seis, um)
	}
}
