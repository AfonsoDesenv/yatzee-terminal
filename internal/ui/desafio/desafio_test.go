package desafio

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/regras"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

var nomeadas = map[string]rune{
	"enter": tea.KeyEnter,
	"esc":   tea.KeyEscape,
	"space": tea.KeySpace,
	"up":    tea.KeyUp,
	"down":  tea.KeyDown,
}

func tecla(s Screen, k string) (Screen, tea.Cmd) {
	msg := tea.KeyPressMsg{Code: rune(k[0]), Text: k}
	if code, ok := nomeadas[k]; ok {
		msg = tea.KeyPressMsg{Code: code}
	}
	next, cmd := s.Update(msg)
	return next.(Screen), cmd
}

func naTela() Screen {
	s := New()
	next, _ := s.Update(tea.WindowSizeMsg{Width: 118, Height: 32})
	return next.(Screen)
}

func TestCursorAndaEParaNasPontas(t *testing.T) {
	s := naTela()
	if s.Cursor() != 0 {
		t.Fatalf("cursor inicial = %d, quero 0", s.Cursor())
	}

	s, _ = tecla(s, "up")
	if s.Cursor() != 0 {
		t.Errorf("subiu do topo para %d", s.Cursor())
	}

	fim := len(regras.Desafios) - 1
	for i := 0; i < len(regras.Desafios)+2; i++ {
		s, _ = tecla(s, "down")
	}
	if s.Cursor() != fim {
		t.Errorf("cursor = %d depois de descer demais, quero parar em %d", s.Cursor(), fim)
	}
}

// Espaço começa o desafio sob o cursor, e o índice é o do catálogo: é por ele
// que a raiz acha o desafio.
func TestEspacoComecaODesafioSobOCursor(t *testing.T) {
	for i := range regras.Desafios {
		s := naTela()
		for j := 0; j < i; j++ {
			s, _ = tecla(s, "down")
		}

		_, cmd := tecla(s, "space")
		if cmd == nil {
			t.Fatalf("desafio %d: espaco nao devolveu comando", i)
		}
		msg, ok := cmd().(ui.StartChallengeMsg)
		if !ok {
			t.Fatalf("desafio %d emitiu %T, quero ui.StartChallengeMsg", i, cmd())
		}
		if msg.Desafio != i {
			t.Errorf("com o cursor em %d a tela pediu o desafio %d", i, msg.Desafio)
		}
	}
}

func TestEscVolta(t *testing.T) {
	_, cmd := tecla(naTela(), "esc")
	if cmd == nil {
		t.Fatal("esc nao devolveu comando")
	}
	if _, ok := cmd().(ui.BackMsg); !ok {
		t.Errorf("esc emitiu %T, quero ui.BackMsg", cmd())
	}
}

// A tela lê o catálogo na hora: todo desafio aparece, sem esta tela guardar
// cópia nenhuma. É o que faz um desafio novo não pedir mudança aqui.
func TestATelaListaTodoOCatalogo(t *testing.T) {
	v := naTela().View()
	for _, d := range regras.Desafios {
		if !strings.Contains(v, d.Nome) {
			t.Errorf("o desafio %q nao aparece na lista", d.Nome)
		}
	}
}

// O resumo acompanha o cursor, e inteiro: cortado pela metade nao explicaria
// a condicao, que e o que a pessoa precisa para escolher.
func TestOResumoSegueOCursorESaiInteiro(t *testing.T) {
	s := naTela()
	for i, d := range regras.Desafios {
		s.cursor = i
		v := s.View()

		if !strings.Contains(v, d.Resumo) {
			t.Errorf("com o cursor no desafio %d a tela nao mostra %q", i, d.Resumo)
		}
		for j, outro := range regras.Desafios {
			if j == i || outro.Resumo == d.Resumo {
				continue
			}
			if strings.Contains(v, outro.Resumo) {
				t.Errorf("com o cursor no desafio %d a tela mostra o resumo do %d", i, j)
			}
		}
	}
}

// O desafio em foco sai pintado, e só ele.
func TestSoODesafioEmFocoFicaPintado(t *testing.T) {
	s := naTela()
	cor := s.theme.Selected.Render("x")
	cor = cor[:strings.Index(cor, "x")]

	for i := range regras.Desafios {
		s.cursor = i

		// A placa em foco sai pintada e as outras limpas: verificar a posição,
		// e não só a presença da cor, é o que pega acender a placa errada.
		for j, d := range regras.Desafios {
			var pintada bool
			for _, l := range strings.Split(s.View(), "\n") {
				if strings.Contains(l, d.Nome) && strings.Contains(l, cor) {
					pintada = true
				}
			}
			if i == j && !pintada {
				t.Errorf("cursor em %d: a placa de %q nao esta pintada", i, d.Nome)
			}
			if i != j && pintada {
				t.Errorf("cursor em %d: a placa de %q acendeu tambem", i, d.Nome)
			}
		}
	}
}

// A largura acompanha o resumo mais longo: sem isso um desafio novo com
// explicacao comprida teria o texto cortado.
func TestALarguraCabeOResumoMaisLongo(t *testing.T) {
	maior := 0
	for _, d := range regras.Desafios {
		if l := ui.Largura(d.Resumo); l > maior {
			maior = l
		}
	}
	if got := larguraInterna(); got < maior {
		t.Errorf("largura interna %d nao cabe o resumo de %d colunas", got, maior)
	}
}

// Nada pode vazar da largura da tela.
func TestATelaNaoVazaDaLargura(t *testing.T) {
	s := naTela()
	for i, l := range strings.Split(s.View(), "\n") {
		if got := ui.Largura(l); got > 118 {
			t.Errorf("linha %d mede %d colunas, a tela tem 118", i, got)
		}
	}
}
