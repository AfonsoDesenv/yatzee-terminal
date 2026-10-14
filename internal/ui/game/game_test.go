package game

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/regras"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

func aplica(s Screen, msg tea.Msg) (Screen, tea.Cmd) {
	next, cmd := s.Update(msg)
	return next.(Screen), cmd
}

func tecla(s Screen, k string) (Screen, tea.Cmd) {
	if k == "enter" {
		return aplica(s, tea.KeyPressMsg{Code: tea.KeyEnter})
	}
	if k == "esc" {
		return aplica(s, tea.KeyPressMsg{Code: tea.KeyEscape})
	}
	return aplica(s, tea.KeyPressMsg{Code: rune(k[0]), Text: k})
}

// contaTudo roda todos os tiques da contagem até ela acabar.
func contaTudo(s Screen) Screen {
	for i := 0; i < 30 && s.Phase() == PhaseContando; i++ {
		s, _ = aplica(s, tiquePontuacao{})
	}
	return s
}

// venceAgora força uma vitória: dá pontos suficientes e envia.
func venceAgora(s Screen) Screen {
	s.partida.Pontos = s.partida.Alvo
	s, _ = tecla(s, "enter")
	return contaTudo(s)
}

func TestComecaNaRodadaComDadosRolados(t *testing.T) {
	s := New()
	if s.Phase() != PhaseRodada {
		t.Errorf("fase inicial = %v, quero PhaseRodada", s.Phase())
	}
	for i, d := range s.Partida().Dados {
		if d < 1 || d > 6 {
			t.Errorf("dado %d fora de 1..6: %d", i, d)
		}
	}
}

func TestDigitoMarcaDado(t *testing.T) {
	s, _ := tecla(New(), "3")
	if !s.Partida().Marcados[2] {
		t.Error("digitar 3 não marcou o terceiro dado")
	}
	s, _ = tecla(s, "3")
	if s.Partida().Marcados[2] {
		t.Error("digitar de novo não desmarcou")
	}
}

func TestRerollGastaUmRerollETrocaOsMarcados(t *testing.T) {
	s := New()
	antes := s.Partida().Dados
	s, _ = tecla(s, "1")
	s, _ = tecla(s, "r")

	if got := s.Partida().Rerolls; got != regras.RerollsPorRodada-1 {
		t.Errorf("rerolls = %d, quero %d", got, regras.RerollsPorRodada-1)
	}
	for i := 1; i < regras.DadosPorMao; i++ {
		if s.Partida().Dados[i] != antes[i] {
			t.Errorf("dado %d mudou sem estar marcado", i+1)
		}
	}
}

func TestRerollSemMarcarNaoGasta(t *testing.T) {
	s, _ := tecla(New(), "r")
	if got := s.Partida().Rerolls; got != regras.RerollsPorRodada {
		t.Errorf("rerolls = %d depois de rerolar sem marcar, quero %d", got, regras.RerollsPorRodada)
	}
}

func TestEnviarComecaAContagemEGastaUmaMao(t *testing.T) {
	s, cmd := tecla(New(), "enter")

	if s.Phase() != PhaseContando {
		t.Errorf("fase = %v depois de enviar, quero PhaseContando", s.Phase())
	}
	if cmd == nil {
		t.Error("enviar não agendou o primeiro tique")
	}
	if got := s.Partida().Maos; got != regras.MaosPorRodada-1 {
		t.Errorf("mãos = %d, quero %d", got, regras.MaosPorRodada-1)
	}
}

// Durante a contagem o teclado não pode mexer no estado que ela percorre.
func TestTeclasSaoIgnoradasDuranteAContagem(t *testing.T) {
	s, _ := tecla(New(), "enter")
	antes := s.View()

	for _, k := range []string{"1", "3", "r", "esc"} {
		s, _ = tecla(s, k)
	}
	if s.Phase() != PhaseContando {
		t.Errorf("fase = %v, uma tecla escapou durante a contagem", s.Phase())
	}
	if s.View() != antes {
		t.Error("uma tecla mudou a tela no meio da contagem")
	}
}

func TestContagemSomaOsDadosEDepoisACombinacao(t *testing.T) {
	s := New()
	s.partida.Dados = regras.Mao{3, 3, 3, 6, 6} // full house, soma 21
	s, _ = tecla(s, "enter")

	// Cinco tiques somam os dados, um por um.
	for i := 0; i < regras.DadosPorMao; i++ {
		s, _ = aplica(s, tiquePontuacao{})
	}
	if got := s.anim.fichas; got != 21 {
		t.Errorf("fichas depois dos dados = %d, quero 21", got)
	}
	if got := s.anim.mult; got != 1 {
		t.Errorf("mult antes da combinação = %d, quero 1", got)
	}

	// O último tique aplica a combinação.
	s, _ = aplica(s, tiquePontuacao{})
	if got := s.anim.fichas; got != 26 {
		t.Errorf("fichas com a base do full house = %d, quero 26", got)
	}
	if got := s.anim.mult; got != 5 {
		t.Errorf("mult = %d, quero 5", got)
	}
}

func TestNaoBaterAMetaRolaOsDadosDeNovo(t *testing.T) {
	s := New()
	s.partida.Alvo = 1_000_000 // inalcançável numa mão
	s, _ = tecla(s, "enter")
	s = contaTudo(s)

	if s.Phase() != PhaseRodada {
		t.Errorf("fase = %v, quero voltar para a rodada", s.Phase())
	}
	if s.Partida().AlgumMarcado() {
		t.Error("as marcas sobreviveram à jogada")
	}
}

func TestBaterAMetaComemoraEDepoisAbreOModal(t *testing.T) {
	s := venceAgora(New())
	if s.Phase() != PhaseComemorando {
		t.Fatalf("fase = %v depois de bater a meta, quero PhaseComemorando", s.Phase())
	}

	s, _ = aplica(s, espalhaFogo{})
	s, _ = aplica(s, apagaFogo{})
	for i := 0; i < 2000 && s.Phase() == PhaseComemorando; i++ {
		s, _ = aplica(s, tiqueFogo{})
	}

	if s.Phase() != PhaseVitoria {
		t.Fatalf("fase = %v depois do fogo apagar, quero PhaseVitoria", s.Phase())
	}
	if !strings.Contains(s.View(), "META BATIDA") {
		t.Error("o modal de vitória não apareceu")
	}
}

func TestAcabarAsMaosAbreADerrota(t *testing.T) {
	s := New()
	s.partida.Alvo = 1_000_000
	s.partida.Maos = 1
	s, _ = tecla(s, "enter")
	s = contaTudo(s)

	if s.Phase() != PhaseDerrota {
		t.Fatalf("fase = %v com as mãos esgotadas, quero PhaseDerrota", s.Phase())
	}
	if !strings.Contains(s.View(), "ACABARAM AS MAOS") {
		t.Error("o modal de derrota não apareceu")
	}
}

func TestDerrotaVoltaAoMenu(t *testing.T) {
	s := New()
	s.phase = PhaseDerrota
	_, cmd := tecla(s, "enter")
	if cmd == nil {
		t.Fatal("enter na derrota não produziu comando")
	}
	if _, ok := cmd().(ui.BackMsg); !ok {
		t.Errorf("enter na derrota emitiu %T, quero ui.BackMsg", cmd())
	}
}

func TestVitoriaLevaALojaELojaAoProximoNivel(t *testing.T) {
	s := New()
	s.phase = PhaseVitoria
	nivel := s.Partida().Nivel
	alvo := s.Partida().Alvo
	recompensa := s.Partida().Recompensa()

	s, _ = tecla(s, "enter")
	if s.Phase() != PhaseLoja {
		t.Fatalf("fase = %v depois de continuar, quero PhaseLoja", s.Phase())
	}

	s, _ = tecla(s, "enter")
	if s.Phase() != PhaseRodada {
		t.Fatalf("fase = %v depois da loja, quero PhaseRodada", s.Phase())
	}
	p := s.Partida()
	if p.Nivel != nivel+1 {
		t.Errorf("nível = %d, quero %d", p.Nivel, nivel+1)
	}
	if p.Alvo != alvo*2 {
		t.Errorf("alvo = %d, quero o dobro de %d", p.Alvo, alvo)
	}
	if p.Moedas != recompensa {
		t.Errorf("moedas = %d, quero %d", p.Moedas, recompensa)
	}
	if p.Maos != regras.MaosPorRodada || p.Rerolls != regras.RerollsPorRodada {
		t.Errorf("mãos/rerolls = %d/%d, quero cheios", p.Maos, p.Rerolls)
	}
}

// O modal tem de detalhar a conta, senão o jogador não aprende que sobra vira
// moeda.
func TestModalDeVitoriaDetalhaARecompensa(t *testing.T) {
	s := New()
	s.phase = PhaseVitoria
	s.partida.Maos = 2
	s.partida.Rerolls = 1

	v := s.View()
	for _, termo := range []string{"vitoria", "maos de sobra", "rerolls de sobra", "total"} {
		if !strings.Contains(v, termo) {
			t.Errorf("o modal não mostra %q:\n%s", termo, v)
		}
	}
}

func TestEscVoltaAoMenuDuranteARodada(t *testing.T) {
	_, cmd := tecla(New(), "esc")
	if cmd == nil {
		t.Fatal("esc não produziu comando")
	}
	if _, ok := cmd().(ui.BackMsg); !ok {
		t.Errorf("esc emitiu %T, quero ui.BackMsg", cmd())
	}
}

func TestMensagemNaoTeclaNaoMudaAFase(t *testing.T) {
	s, cmd := aplica(New(), tea.WindowSizeMsg{Width: 100, Height: 30})
	if s.Phase() != PhaseRodada {
		t.Errorf("fase = %v, quero PhaseRodada", s.Phase())
	}
	if cmd != nil {
		t.Errorf("mensagem não-tecla devolveu comando %v", cmd())
	}
}
