package regras

import "testing"

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

// G..L somam ao multiplicador uma única vez, mesmo com vários dados do valor.
func TestRunasDeMultSomamUmaVezSo(t *testing.T) {
	casos := []struct {
		letra string
		mao   Mao
		mult  int
	}{
		{"G", Mao{1, 2, 3, 4, 5}, MultDaRuna},
		{"G", Mao{1, 1, 1, 1, 1}, MultDaRuna}, // cinco dados, ainda uma vez
		{"G", Mao{2, 3, 4, 5, 6}, 0},
		{"L", Mao{6, 1, 2, 3, 4}, MultDaRuna},
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

	// L soma 2 ao mult por ter um 6.
	comL := AvaliaCom(m, []Runa{runa(t, "L")})
	if quero := (5 + 21) * (5 + 2); comL.Total() != quero {
		t.Errorf("com a runa L = %d, quero %d", comL.Total(), quero)
	}

	// Juntas: as fichas somam e o mult soma, nunca multiplicam entre si.
	comAmbas := AvaliaCom(m, []Runa{runa(t, "C"), runa(t, "L")})
	if quero := (5 + 21 + 30) * (5 + 2); comAmbas.Total() != quero {
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
