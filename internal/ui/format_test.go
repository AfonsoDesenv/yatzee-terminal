package ui

import "testing"

func TestNumeroPorExtenso(t *testing.T) {
	casos := []struct {
		n     int64
		quero string
	}{
		{0, "0"},
		{7, "7"},
		{999, "999"},
		{1000, "1.000"},
		{12345, "12.345"},
		{999999, "999.999"},
		{1234567, "1.234.567"},
		{LimiteExtenso, "9.999.999"}, // o último que cabe por extenso
	}
	for _, c := range casos {
		if got := Numero(c.n); got != c.quero {
			t.Errorf("Numero(%d) = %q, quero %q", c.n, got, c.quero)
		}
	}
}

func TestNumeroEmNotacaoCientifica(t *testing.T) {
	casos := []struct {
		n     int64
		quero string
	}{
		{LimiteExtenso + 1, "1.00e7"}, // primeiro que não cabe mais
		{12_345_678, "1.23e7"},
		{98_765_432, "9.88e7"},
		{1_000_000_000, "1.00e9"},
		{5_400_000_000_000, "5.40e12"},
	}
	for _, c := range casos {
		if got := Numero(c.n); got != c.quero {
			t.Errorf("Numero(%d) = %q, quero %q", c.n, got, c.quero)
		}
	}
}

// A mantissa arredondada pode estourar o 10 (9,999... vira "10.00"), e aí o
// expoente tem que subir em vez de sair um "10.00e11" fora do formato.
func TestNumeroQuandoAMantissaEstoura(t *testing.T) {
	if got := Numero(999_999_999_999); got != "1.00e12" {
		t.Errorf("Numero(999999999999) = %q, quero 1.00e12", got)
	}
}

// O painel tem largura fixa: nenhum valor plausível pode transbordar.
func TestNumeroCabeEmOitoColunas(t *testing.T) {
	valores := []int64{0, 999, LimiteExtenso, LimiteExtenso + 1, 1 << 40, 1 << 62}
	for _, v := range valores {
		if got := Numero(v); len(got) > 9 {
			t.Errorf("Numero(%d) = %q, com %d colunas — não cabe no painel", v, got, len(got))
		}
	}
}

func TestNumeroNegativo(t *testing.T) {
	if got := Numero(-1234); got != "-1.234" {
		t.Errorf("Numero(-1234) = %q, quero -1.234", got)
	}
}
