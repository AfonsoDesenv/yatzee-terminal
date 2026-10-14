package ui

import "strings"

// grade devolve as posições acesas de um valor, numa grade 3x3. É a disposição
// canônica do dado: ímpares têm o centro, pares não.
func grade(v int) [3][3]bool {
	switch v {
	case 1:
		return [3][3]bool{{false, false, false}, {false, true, false}, {false, false, false}}
	case 2:
		return [3][3]bool{{true, false, false}, {false, false, false}, {false, false, true}}
	case 3:
		return [3][3]bool{{true, false, false}, {false, true, false}, {false, false, true}}
	case 4:
		return [3][3]bool{{true, false, true}, {false, false, false}, {true, false, true}}
	case 5:
		return [3][3]bool{{true, false, true}, {false, true, false}, {true, false, true}}
	default:
		return [3][3]bool{{true, false, true}, {true, false, true}, {true, false, true}}
	}
}

// LarguraDado e AlturaDado são o tamanho de um dado da mesa, em células.
const (
	LarguraDado = 9
	AlturaDado  = 5
)

// Dado desenha um dado da mesa. Cinco deles cabem lado a lado em 53 colunas,
// que é o que permite a mesa inteira num terminal de 80.
func Dado(v int) []string {
	out := []string{".-------."}
	g := grade(v)
	for r := 0; r < 3; r++ {
		linha := []rune("       ")
		for c := 0; c < 3; c++ {
			if g[r][c] {
				linha[c*3] = 'o'
			}
		}
		out = append(out, "|"+string(linha)+"|")
	}
	return append(out, "'-------'")
}

// MarcaDado é a linha que vai sob um dado escolhido para rerolar. A marcação
// fica fora do dado de propósito: assim o desenho não muda de forma, e sobra
// espaço para indicar outra coisa sobre o próprio dado mais tarde.
func MarcaDado(marcado bool) string {
	if marcado {
		return " ^^^^^^^ "
	}
	return strings.Repeat(" ", LarguraDado)
}

// ---------- dado grande, em arte de densidade, para o menu ----------

// rampa vai do mais escuro ao mais claro. Degraus bem separados de propósito:
// uma rampa longa faz uma face lisa virar mosaico de letras, porque vizinhos
// na escala de densidade parecem muito diferentes entre si.
const rampa = " .:-=+*#%@"

// arteDado descreve o cubo em projeção de gabinete: a face da frente é um
// retângulo, topo e lateral saem na diagonal.
type arteDado struct {
	w, h, d float64
}

// brilho devolve a luminosidade de um ponto contínuo, ou -1 fora do cubo. O
// topo é o mais claro, a lateral o mais escuro, e os pinos são buracos.
func (a arteDado) brilho(x, y float64, frente, topo, lado int) float64 {
	switch {
	case x >= 0 && x < a.w && y >= a.d && y < a.d+a.h:
		u, v := x/a.w, (y-a.d)/a.h
		if pinoEm(u, v, grade(frente), 0.085, a.h*2/a.w) {
			return 0.05
		}
		return 0.60 - 0.04*v

	case y >= 0 && y < a.d:
		ini := (a.d - y) / a.d * a.d
		if x >= ini && x < ini+a.w {
			u := (x - ini) / a.w
			v := 1 - (a.d-y)/a.d
			if pinoEm(u, v, grade(topo), 0.075, a.d*2/a.w) {
				return 0.10
			}
			return 0.95 - 0.03*(1-v)
		}

	case x >= a.w && x < a.w+a.d:
		t := (x - a.w) / a.d
		alto := a.d * (1 - t)
		if y >= alto && y < alto+a.h {
			if pinoEm(t, (y-alto)/a.h, grade(lado), 0.075, a.h*2/a.d) {
				return 0.03
			}
			return 0.30 - 0.04*t
		}
	}
	return -1
}

// pinoEm diz se (u,v), em coordenadas 0..1 da face, cai num pino. O aspecto
// corrige a célula 1:2 do terminal, senão o pino sai oval.
func pinoEm(u, v float64, g [3][3]bool, raio, aspecto float64) bool {
	for r := 0; r < 3; r++ {
		for c := 0; c < 3; c++ {
			if !g[r][c] {
				continue
			}
			du := u - float64(2*c+1)/6
			dv := (v - float64(2*r+1)/6) * aspecto
			if du*du+dv*dv < raio*raio {
				return true
			}
		}
	}
	return false
}

// DadoArte desenha o dado grande do menu. Cada célula é amostrada em subpixel
// e vira o caractere da rampa correspondente ao brilho médio — é a amostragem
// que suaviza a borda e dá textura, em vez de blocos chapados.
func DadoArte(frente, topo, lado int) []string {
	a := arteDado{w: 24, h: 12, d: 6}
	const amostras = 3

	largura := int(a.w+a.d) + 1
	altura := int(a.h+a.d) + 1

	out := make([]string, 0, altura)
	for cy := 0; cy < altura; cy++ {
		var b strings.Builder
		for cx := 0; cx < largura; cx++ {
			soma := 0.0
			for sy := 0; sy < amostras; sy++ {
				for sx := 0; sx < amostras; sx++ {
					px := float64(cx) + (float64(sx)+0.5)/amostras
					py := float64(cy) + (float64(sy)+0.5)/amostras
					if v := a.brilho(px, py, frente, topo, lado); v >= 0 {
						soma += v
					}
				}
			}
			i := int(soma / float64(amostras*amostras) * float64(len(rampa)-1))
			if i < 0 {
				i = 0
			}
			if i >= len(rampa) {
				i = len(rampa) - 1
			}
			b.WriteByte(rampa[i])
		}
		out = append(out, strings.TrimRight(b.String(), " "))
	}
	return out
}
