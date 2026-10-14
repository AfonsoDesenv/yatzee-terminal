package game

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

// aplica entrega uma mensagem e devolve a tela resultante, como o runtime faria.
func aplica(s Screen, msg tea.Msg) (Screen, tea.Cmd) {
	next, cmd := s.Update(msg)
	return next.(Screen), cmd
}

// contaAteOFim roda todos os tiques de pontuação da jogada.
func contaAteOFim(s Screen) Screen {
	for i := 0; i < 20 && s.Animando(); i++ {
		var cmd tea.Cmd
		s, cmd = aplica(s, tiquePontuacao{})
		if cmd == nil {
			break
		}
	}
	return s
}

func TestEnterComecaAContagem(t *testing.T) {
	s := New()
	if s.Animando() {
		t.Fatal("a rodada já começa animando")
	}

	s, cmd := aplica(s, tea.KeyPressMsg{Code: tea.KeyEnter})
	if !s.Animando() {
		t.Error("enter não começou a contagem")
	}
	if cmd == nil {
		t.Error("enter não agendou o primeiro tique")
	}
}

// Durante a contagem o teclado não pode mexer no estado: a sequência está
// percorrendo exatamente esses números.
func TestTeclasSaoIgnoradasDuranteAContagem(t *testing.T) {
	s, _ := aplica(New(), tea.KeyPressMsg{Code: tea.KeyEnter})
	antes := s.View()

	for _, k := range []tea.KeyPressMsg{
		{Code: '1', Text: "1"},
		{Code: '3', Text: "3"},
		{Code: tea.KeyTab},
	} {
		s, _ = aplica(s, k)
	}

	if s.Phase() != PhaseRound {
		t.Error("tab trocou de fase no meio da contagem")
	}
	if s.View() != antes {
		t.Error("uma tecla mudou a tela no meio da contagem")
	}
}

// esc continua funcionando mesmo animando? Não: a volta ao menu também espera
// a contagem acabar, senão a tela sairia no meio da sequência.
func TestEscTambemEsperaAContagem(t *testing.T) {
	s, _ := aplica(New(), tea.KeyPressMsg{Code: tea.KeyEnter})
	_, cmd := aplica(s, tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd != nil {
		t.Errorf("esc produziu %v durante a contagem", cmd())
	}
}

func TestCadaTiqueAvancaUmPasso(t *testing.T) {
	s, _ := aplica(New(), tea.KeyPressMsg{Code: tea.KeyEnter})

	// O primeiro passo soma 5 fichas; antes de qualquer tique, nada somou.
	if got := s.round.fichas; got != 0 {
		t.Fatalf("fichas antes do primeiro tique = %d, quero 0", got)
	}
	s, _ = aplica(s, tiquePontuacao{})
	if got := s.round.fichas; got != 5 {
		t.Errorf("fichas depois de um tique = %d, quero 5", got)
	}
	s, _ = aplica(s, tiquePontuacao{})
	if got := s.round.fichas; got != 10 {
		t.Errorf("fichas depois de dois tiques = %d, quero 10", got)
	}
}

func TestBaterAMetaAcendeOFogoNoPainel(t *testing.T) {
	s, _ := aplica(New(), tea.KeyPressMsg{Code: tea.KeyEnter})
	s = contaAteOFim(s)

	if !s.Animando() {
		t.Fatal("a animação acabou sem comemorar: a jogada não bateu a meta")
	}
	if s.round.anim.escopo != fogoPainel {
		t.Errorf("escopo = %v, quero o fogo começar no painel", s.round.anim.escopo)
	}
	if s.round.pontos < s.round.alvo {
		t.Errorf("pontos %d não alcançaram o alvo %d", s.round.pontos, s.round.alvo)
	}
}

// A comemoração escala: do painel para a tela inteira.
func TestOFogoEspalhaParaATela(t *testing.T) {
	s, _ := aplica(New(), tea.KeyPressMsg{Code: tea.KeyEnter})
	s = contaAteOFim(s)

	s, cmd := aplica(s, espalhaFogo{})
	if s.round.anim.escopo != fogoTela {
		t.Fatalf("escopo = %v depois de espalhar, quero a tela inteira", s.round.anim.escopo)
	}
	if cmd == nil {
		t.Error("espalhar não agendou o fim do fogo")
	}
}

// Sem bater a meta não há comemoração, e a rodada volta a aceitar teclas.
func TestNaoBaterAMetaNaoAcendeNada(t *testing.T) {
	s := New()
	s.round.alvo = 1_000_000
	s, _ = aplica(s, tea.KeyPressMsg{Code: tea.KeyEnter})
	s = contaAteOFim(s)

	if s.Animando() {
		t.Error("continuou animando sem bater a meta")
	}
}

// A chama precisa aparecer no desenho, não só existir no estado.
func TestOFogoApareceNaTela(t *testing.T) {
	s, _ := aplica(New(), tea.KeyPressMsg{Code: tea.KeyEnter})
	s = contaAteOFim(s)
	for i := 0; i < 15; i++ {
		s, _ = aplica(s, tiqueFogo{})
	}

	v := s.View()
	if !strings.ContainsAny(v, "@%#*") {
		t.Errorf("nenhuma chama no desenho:\n%s", v)
	}
	if !strings.Contains(v, "ALVO") {
		t.Error("o painel sumiu embaixo do fogo")
	}
}

// E a comemoração precisa terminar, senão a tela fica presa comemorando.
func TestAComemoracaoTermina(t *testing.T) {
	s, _ := aplica(New(), tea.KeyPressMsg{Code: tea.KeyEnter})
	s = contaAteOFim(s)
	s, _ = aplica(s, espalhaFogo{})
	s, _ = aplica(s, apagaFogo{})

	for i := 0; i < 2000 && s.Animando(); i++ {
		s, _ = aplica(s, tiqueFogo{})
	}
	if s.Animando() {
		t.Fatal("a comemoração não terminou depois de 2000 quadros")
	}

	// E a rodada volta a responder.
	s2, _ := aplica(s, tea.KeyPressMsg{Code: tea.KeyTab})
	if s2.Phase() != PhaseShop {
		t.Error("a rodada não voltou a aceitar teclas depois da comemoração")
	}
}

func TestAjudaMudaDuranteAAnimacao(t *testing.T) {
	s := New()
	parada := strings.Join(s.round.ajuda(), " ")
	s, _ = aplica(s, tea.KeyPressMsg{Code: tea.KeyEnter})
	contando := strings.Join(s.round.ajuda(), " ")

	if parada == contando {
		t.Error("a barra de ajuda não avisa que a jogada está sendo contada")
	}
	if !strings.Contains(contando, "contando") {
		t.Errorf("ajuda durante a contagem = %q", contando)
	}
}

var _ = ui.DefaultWidth
