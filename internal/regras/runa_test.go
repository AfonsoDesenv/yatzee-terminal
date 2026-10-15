package regras

import (
	"strings"
	"testing"
)

func runa(t *testing.T, letra string) Runa {
	t.Helper()
	r, ok := RunaPorLetra(letra)
	if !ok {
		t.Fatalf("runa %q não existe no catálogo", letra)
	}
	return r
}

func TestCatalogoTemAsDozeRunas(t *testing.T) {
	if len(Catalogo) != 12 {
		t.Fatalf("catálogo com %d runas, quero 12", len(Catalogo))
	}
	for _, letra := range []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L"} {
		if _, ok := RunaPorLetra(letra); !ok {
			t.Errorf("falta a runa %s", letra)
		}
	}
}

// A..F somam fichas uma vez por dado com o valor correspondente.
func TestRunasDeFichasSomamPorDado(t *testing.T) {
	casos := []struct {
		letra  string
		mao    Mao
		fichas int
	}{
		{"A", Mao{1, 1, 1, 2, 3}, 30}, // três dados com 1
		{"A", Mao{2, 3, 4, 5, 6}, 0},  // nenhum
		{"B", Mao{2, 2, 5, 5, 5}, 20},
		{"F", Mao{6, 6, 6, 6, 6}, 50}, // yatzee de seis
		{"C", Mao{3, 1, 2, 4, 5}, 10},
	}
	for _, c := range casos {
		f, mult := runa(t, c.letra).Aplica(c.mao)
		if f != c.fichas {
			t.Errorf("runa %s em %v somou %d fichas, quero %d", c.letra, c.mao, f, c.fichas)
		}
		if mult != 0 {
			t.Errorf("runa %s mexeu no multiplicador (%d), não devia", c.letra, mult)
		}
	}
}

// G..L somam ao multiplicador uma vez por dado com o valor, igual às de
// fichas: três dados com o valor somam três vezes.
func TestRunasDeMultSomamPorDado(t *testing.T) {
	casos := []struct {
		letra string
		mao   Mao
		mult  int
	}{
		{"G", Mao{1, 2, 3, 4, 5}, MultDaRuna},
		{"G", Mao{1, 1, 4, 5, 6}, 2 * MultDaRuna}, // dois dados, duas vezes
		{"G", Mao{1, 1, 1, 1, 1}, 5 * MultDaRuna}, // yatzee de um: cinco vezes
		{"G", Mao{2, 3, 4, 5, 6}, 0},
		{"L", Mao{6, 1, 2, 3, 4}, MultDaRuna},
		{"L", Mao{6, 6, 6, 3, 3}, 3 * MultDaRuna}, // três 6 somam +6 ao mult
		{"J", Mao{1, 2, 3, 5, 6}, 0},
	}
	for _, c := range casos {
		f, mult := runa(t, c.letra).Aplica(c.mao)
		if mult != c.mult {
			t.Errorf("runa %s em %v somou %d ao mult, quero %d", c.letra, c.mao, mult, c.mult)
		}
		if f != 0 {
			t.Errorf("runa %s mexeu em fichas (%d), não devia", c.letra, f)
		}
	}
}

func TestDisparaSoQuandoTemEfeito(t *testing.T) {
	if runa(t, "A").Dispara(Mao{2, 3, 4, 5, 6}) {
		t.Error("runa A disparou numa mão sem nenhum 1")
	}
	if !runa(t, "A").Dispara(Mao{1, 3, 4, 5, 6}) {
		t.Error("runa A não disparou com um 1 na mão")
	}
	if runa(t, "G").Dispara(Mao{2, 3, 4, 5, 6}) {
		t.Error("runa G disparou sem nenhum 1")
	}
}

func TestDescricaoDizOQueFaz(t *testing.T) {
	for _, r := range Catalogo {
		if r.Descricao() == "" {
			t.Errorf("runa %s sem descrição", r.Letra)
		}
	}
}

// A pontuação com runas soma fichas e multiplicador por cima da combinação.
func TestAvaliaComRunas(t *testing.T) {
	m := Mao{3, 3, 3, 6, 6} // full house: (5+21) x 5 = 130

	semRunas := AvaliaCom(m, nil)
	if semRunas.Total() != 130 {
		t.Fatalf("sem runas = %d, quero 130", semRunas.Total())
	}

	// C soma 10 por dado 3: três dados, +30 fichas.
	comC := AvaliaCom(m, []Runa{runa(t, "C")})
	if quero := (5 + 21 + 30) * 5; comC.Total() != quero {
		t.Errorf("com a runa C = %d, quero %d", comC.Total(), quero)
	}

	// L soma 2 ao mult por dado 6: dois dados, +4 de mult.
	comL := AvaliaCom(m, []Runa{runa(t, "L")})
	if quero := (5 + 21) * (5 + 2*2); comL.Total() != quero {
		t.Errorf("com a runa L = %d, quero %d", comL.Total(), quero)
	}

	// Juntas: as fichas somam e o mult soma, nunca multiplicam entre si.
	comAmbas := AvaliaCom(m, []Runa{runa(t, "C"), runa(t, "L")})
	if quero := (5 + 21 + 3*10) * (5 + 2*2); comAmbas.Total() != quero {
		t.Errorf("com C e L = %d, quero %d", comAmbas.Total(), quero)
	}
}

// Runa que não dispara não muda nada.
func TestRunaQueNaoDisparaNaoMudaAPontuacao(t *testing.T) {
	m := Mao{3, 3, 3, 6, 6}
	sem := AvaliaCom(m, nil).Total()
	com := AvaliaCom(m, []Runa{runa(t, "A")}).Total() // nenhum dado 1
	if sem != com {
		t.Errorf("runa inerte mudou a pontuação: %d viva %d", sem, com)
	}
}

// A regra, no exemplo que a definiu: "cada 6 é +2 no multiplicador" significa
// que três dados com 6 somam +6, não +2. As duas famílias contam por dado.
func TestCadaDadoContaUmaVez(t *testing.T) {
	tresSeis := Mao{6, 6, 6, 1, 2}

	if _, mult := runa(t, "L").Aplica(tresSeis); mult != 6 {
		t.Errorf("tres dados com 6 somaram %d ao mult, quero 6 (2 por dado)", mult)
	}
	if fichas, _ := runa(t, "F").Aplica(tresSeis); fichas != 30 {
		t.Errorf("tres dados com 6 somaram %d fichas, quero 30 (10 por dado)", fichas)
	}

	// E a conta fechada: trinca de 6 é base 5 x 4, mais 21 dos dados.
	if quero := (5 + 21) * (4 + 3*MultDaRuna); AvaliaCom(tresSeis, []Runa{runa(t, "L")}).Total() != quero {
		t.Errorf("trinca de 6 com a runa L = %d, quero %d",
			AvaliaCom(tresSeis, []Runa{runa(t, "L")}).Total(), quero)
	}

	// Um dado a mais do valor é um efeito a mais, sempre.
	for n := 1; n <= 5; n++ {
		m := Mao{}
		for i := 0; i < n; i++ {
			m[i] = 6
		}
		for i := n; i < len(m); i++ {
			m[i] = 1
		}
		_, mult := runa(t, "L").Aplica(m)
		if quero := n * MultDaRuna; mult != quero {
			t.Errorf("%d dados com 6 somaram %d ao mult, quero %d", n, mult, quero)
		}
		fichas, _ := runa(t, "F").Aplica(m)
		if quero := n * FichasPorDadoDaRuna; fichas != quero {
			t.Errorf("%d dados com 6 somaram %d fichas, quero %d", n, fichas, quero)
		}
	}
}

// A descrição tem de dizer "cada", não "tem um": a loja é onde a pessoa
// decide a compra, e a diferença entre os dois muda o valor da runa.
func TestADescricaoDizCadaDado(t *testing.T) {
	for _, r := range Catalogo {
		if got := r.Descricao(); !strings.HasPrefix(got, "cada dado ") {
			t.Errorf("runa %s descreve-se como %q, quero começar com \"cada dado\"", r.Letra, got)
		}
	}
}
