package ui

import (
	"fmt"
	"strconv"
	"strings"
)

// LimiteExtenso é o maior valor mostrado por extenso. Acima dele o número vira
// notação científica, senão o painel teria de crescer sem fim — é o mesmo
// recurso que o Balatro usa quando a pontuação explode.
const LimiteExtenso = 9_999_999

// Numero formata um valor para os painéis. Até LimiteExtenso sai por extenso
// com ponto separando milhar; acima disso vira notação científica com três
// algarismos significativos, o que cabe em seis colunas (1.23e7).
func Numero(n int64) string {
	if n < 0 {
		return "-" + Numero(-n)
	}
	if n <= LimiteExtenso {
		return comMilhar(n)
	}

	exp := 0
	f := float64(n)
	for f >= 10 {
		f /= 10
		exp++
	}
	// Arredondar a mantissa pode estourar o 10 (9,999 vira 10,00): aí o
	// expoente sobe e a mantissa volta para 1.
	if s := fmt.Sprintf("%.2f", f); s == "10.00" {
		f, exp = 1, exp+1
	}
	return fmt.Sprintf("%.2fe%d", f, exp)
}

// comMilhar insere ponto a cada três casas, da direita para a esquerda.
func comMilhar(n int64) string {
	s := strconv.FormatInt(n, 10)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	primeiro := len(s) % 3
	if primeiro > 0 {
		b.WriteString(s[:primeiro])
	}
	for i := primeiro; i < len(s); i += 3 {
		if b.Len() > 0 {
			b.WriteByte('.')
		}
		b.WriteString(s[i : i+3])
	}
	return b.String()
}
