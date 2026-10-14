package app

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

// telaFake é uma tela de teste: devolve sempre o mesmo comando, para o teste
// controlar qual intenção a raiz vai receber.
type telaFake struct {
	nome string
	cmd  tea.Cmd
}

func (t telaFake) Init() tea.Cmd                       { return nil }
func (t telaFake) Update(tea.Msg) (ui.Screen, tea.Cmd) { return t, t.cmd }
func (t telaFake) View() string                        { return "tela " + t.nome }
func (t telaFake) Title() string                       { return t.nome }

func rootDeTeste(t *testing.T, inicial ui.Screen) Root {
	t.Helper()
	return NewRoot(sessaoDeTeste(t), inicial)
}

// Aplica a mensagem e, se vier comando, aplica o resultado dele também: é como
// o runtime do Bubble Tea se comporta, então o teste exercita o caminho real.
func passo(m Root, msg tea.Msg) Root {
	next, cmd := m.Update(msg)
	m = next.(Root)
	// Cap de iterações: uma tela que devolvesse comando a cada Update
	// travaria o teste em laço infinito em vez de falhar com mensagem.
	for i := 0; cmd != nil && i < 10; i++ {
		produced := cmd()
		if produced == nil {
			break
		}
		next, cmd = m.Update(produced)
		m = next.(Root)
	}
	return m
}

func TestStartGameEmpilhaAVezDeTrocar(t *testing.T) {
	m := rootDeTeste(t, telaFake{nome: "menu", cmd: ui.StartGame})
	if m.Depth() != 1 {
		t.Fatalf("profundidade inicial = %d, quero 1", m.Depth())
	}

	m = passo(m, tea.KeyPressMsg{})
	if m.Depth() != 2 {
		t.Fatalf("depois de StartGame a profundidade = %d, quero 2", m.Depth())
	}
	if got := m.Top().Title(); got != "jogo" {
		t.Errorf("topo = %q, quero jogo", got)
	}
}

func TestBackVoltaUmaTela(t *testing.T) {
	m := rootDeTeste(t, telaFake{nome: "menu", cmd: ui.ShowScores})
	m = passo(m, tea.KeyPressMsg{})
	if m.Depth() != 2 {
		t.Fatalf("profundidade = %d, quero 2", m.Depth())
	}

	m = passo(m, ui.BackMsg{})
	if m.Depth() != 1 {
		t.Errorf("depois de Back a profundidade = %d, quero 1", m.Depth())
	}
	if got := m.Top().Title(); got != "menu" {
		t.Errorf("topo = %q, quero menu", got)
	}
}

// O menu é a base da pilha: esc ali não pode encerrar a sessão por reflexo.
func TestBackNaBaseNaoDesempilha(t *testing.T) {
	m := rootDeTeste(t, telaFake{nome: "menu"})
	m = passo(m, ui.BackMsg{})
	if m.Depth() != 1 {
		t.Errorf("profundidade = %d, quero 1", m.Depth())
	}
}

// Depois de se apresentar, esc não pode levar de volta pra tela de nome.
func TestNameSetTrocaATelaSemDeixarVolta(t *testing.T) {
	m := rootDeTeste(t, telaFake{nome: "nome"})
	m = passo(m, ui.NameSetMsg{Name: "afonso"})

	if m.Depth() != 1 {
		t.Errorf("profundidade = %d, quero 1 (troca, não empilha)", m.Depth())
	}
	if got := m.Top().Title(); got != "menu" {
		t.Errorf("topo = %q, quero menu", got)
	}
	if m.Session().Identity.Name != "afonso" {
		t.Errorf("identidade da sessão = %q, quero afonso", m.Session().Identity.Name)
	}
}

func TestWindowSizeZeroCaiNoDefault(t *testing.T) {
	m := rootDeTeste(t, telaFake{nome: "menu"})

	// Primeiro um tamanho real, para o estado deixar de ser o default. Sem
	// este passo o teste passaria mesmo se o tratamento de WindowSizeMsg
	// fosse removido, porque NewRoot já começa no default — e aí uma
	// regressão que apagasse o fallback 0x0 passaria batida.
	m = passo(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	if w, h := m.Size(); w != 120 || h != 40 {
		t.Fatalf("Size() = (%d, %d), quero (120, 40)", w, h)
	}

	m = passo(m, tea.WindowSizeMsg{Width: 0, Height: 0})
	w, h := m.Size()
	if w != ui.DefaultWidth || h != ui.DefaultHeight {
		t.Errorf("Size() = (%d, %d), quero (%d, %d)", w, h, ui.DefaultWidth, ui.DefaultHeight)
	}
}

func TestViewUsaAltScreenEOTopoDaPilha(t *testing.T) {
	m := rootDeTeste(t, telaFake{nome: "menu"})
	v := m.View()
	if !v.AltScreen {
		t.Error("View().AltScreen = false, quero true")
	}
	if v.Content == "" {
		t.Error("View().Content vazio")
	}
}

// Empilhar não pode compartilhar o array de trás com a pilha anterior, senão
// uma tela empilhada depois de um Back sobrescreve o histórico.
func TestPilhaNaoCompartilhaArrayDeTras(t *testing.T) {
	m := rootDeTeste(t, telaFake{nome: "menu", cmd: ui.ShowScores})
	m = passo(m, tea.KeyPressMsg{})
	antesDoBack := m.Top().Title()
	m = passo(m, ui.BackMsg{})
	m = passo(m, ui.StartGameMsg{})

	if antesDoBack != "scores" {
		t.Fatalf("topo antes do back = %q, quero scores", antesDoBack)
	}
	if got := m.Top().Title(); got != "jogo" {
		t.Errorf("topo = %q, quero jogo", got)
	}
	if m.Depth() != 2 {
		t.Errorf("profundidade = %d, quero 2", m.Depth())
	}
}

// O Bubble Tea chama Init só no model raiz. Uma tela empilhada precisa ser
// armada pela raiz, senão o placar fica em "carregando..." para sempre.
func TestEmpilharPlacarArmaOInitDaTela(t *testing.T) {
	m := rootDeTeste(t, telaFake{nome: "menu"})

	next, cmd := m.Update(ui.ShowScoresMsg{})
	m = next.(Root)
	if cmd == nil {
		t.Fatal("empilhar o placar não devolveu o Init da tela")
	}
	if !strings.Contains(m.View().Content, "carregando") {
		t.Fatalf("antes da carga o placar deveria estar carregando:\n%s", m.View().Content)
	}

	m = passo(m, cmd())
	if strings.Contains(m.View().Content, "carregando") {
		t.Errorf("depois de aplicar o comando o placar continua carregando:\n%s", m.View().Content)
	}
}
