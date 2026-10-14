package app

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

// telaFake é uma tela de teste: devolve sempre os mesmos comandos, para o
// teste controlar qual intenção a raiz vai receber e se o Init foi chamado.
type telaFake struct {
	nome    string
	cmd     tea.Cmd
	initCmd tea.Cmd
}

func (t telaFake) Init() tea.Cmd                       { return t.initCmd }
func (t telaFake) Update(tea.Msg) (ui.Screen, tea.Cmd) { return t, t.cmd }
func (t telaFake) View() string                        { return "tela " + t.nome }
func (t telaFake) Title() string                       { return t.nome }

// telaMedidora guarda o último tamanho que recebeu, para provar que a raiz
// repassa o tamanho normalizado para a tela do topo.
type telaMedidora struct{ w, h int }

func (t telaMedidora) Init() tea.Cmd { return nil }
func (t telaMedidora) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		t.w, t.h = ws.Width, ws.Height
	}
	return t, nil
}
func (t telaMedidora) View() string  { return "medidora" }
func (t telaMedidora) Title() string { return "medidora" }

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

func TestStartGameEmpilha(t *testing.T) {
	m := NewRoot(telaFake{nome: "menu", cmd: ui.StartGame})
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
	m := NewRoot(telaFake{nome: "menu", cmd: ui.StartGame})
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

// O menu é a base da pilha: esc ali não pode encerrar o jogo por reflexo.
func TestBackNaBaseNaoDesempilha(t *testing.T) {
	m := NewRoot(telaFake{nome: "menu"})
	m = passo(m, ui.BackMsg{})
	if m.Depth() != 1 {
		t.Errorf("profundidade = %d, quero 1", m.Depth())
	}
}

// O Bubble Tea chama Init só uma vez, no model raiz. Sem a raiz devolver o
// Init da tela empilhada, uma tela que carrega algo fica presa para sempre no
// estado inicial. Vale para toda tela futura que precise carregar no Init.
func TestPushArmaOInitDaTelaEmpilhada(t *testing.T) {
	type initMsg struct{}
	armada := telaFake{
		nome:    "armada",
		initCmd: func() tea.Msg { return initMsg{} },
	}

	raiz, cmd := NewRoot(telaFake{nome: "menu"}).pushWith(armada)
	if cmd == nil {
		t.Fatal("pushWith não devolveu o Init da tela: uma tela que carrega algo ficaria presa no estado inicial")
	}
	if _, ok := cmd().(initMsg); !ok {
		t.Errorf("pushWith devolveu %T, quero o comando da própria tela empilhada", cmd())
	}
	if raiz.Depth() != 2 {
		t.Errorf("profundidade = %d, quero 2", raiz.Depth())
	}
}

func TestWindowSizeZeroReemiteTamanhoNormalizado(t *testing.T) {
	m := NewRoot(telaFake{nome: "menu"})

	next, cmd := m.Update(tea.WindowSizeMsg{Width: 0, Height: 0})
	if cmd == nil {
		t.Fatal("tamanho zerado não reemitiu nada: o renderer ficaria vazio e a tela branca")
	}
	ws, ok := cmd().(tea.WindowSizeMsg)
	if !ok {
		t.Fatalf("comando devolveu %T, quero tea.WindowSizeMsg", cmd())
	}
	if ws.Width != ui.DefaultWidth || ws.Height != ui.DefaultHeight {
		t.Errorf("reemitiu (%d, %d), quero (%d, %d)", ws.Width, ws.Height, ui.DefaultWidth, ui.DefaultHeight)
	}

	// A mensagem reemitida é ponto fixo: alimentá-la de volta não reemite de
	// novo, senão o programa giraria em laço.
	_, cmd2 := next.(Root).Update(ws)
	if cmd2 != nil {
		if _, ok := cmd2().(tea.WindowSizeMsg); ok {
			t.Error("o tamanho normalizado foi reemitido outra vez: laço")
		}
	}
}

func TestWindowSizeChegaATelaDoTopoNormalizado(t *testing.T) {
	m := passo(NewRoot(telaMedidora{}), tea.WindowSizeMsg{Width: 0, Height: 0})

	medida, ok := m.Top().(telaMedidora)
	if !ok {
		t.Fatalf("topo é %T, quero telaMedidora", m.Top())
	}
	if medida.w != ui.DefaultWidth || medida.h != ui.DefaultHeight {
		t.Errorf("a tela recebeu (%d, %d), quero (%d, %d)", medida.w, medida.h, ui.DefaultWidth, ui.DefaultHeight)
	}

	w, h := m.Size()
	if w != ui.DefaultWidth || h != ui.DefaultHeight {
		t.Errorf("Size() = (%d, %d), quero (%d, %d)", w, h, ui.DefaultWidth, ui.DefaultHeight)
	}
}

func TestViewUsaAltScreenEOTopoDaPilha(t *testing.T) {
	m := NewRoot(telaFake{nome: "menu"})
	v := m.View()
	if !v.AltScreen {
		t.Error("View().AltScreen = false, quero true")
	}
	if v.Content == "" {
		t.Error("View().Content vazio")
	}
}

// Empilhar não pode compartilhar o array de trás com uma versão anterior do
// model: sem a cópia, empilhar depois de um Back sobrescreve o histórico da
// versão antiga. O teste guarda a versão antiga viva justamente para observar
// isso — comparar só uma string copiada antes não provaria nada.
func TestPilhaNaoCompartilhaArrayDeTras(t *testing.T) {
	antes := NewRoot(telaFake{nome: "base"}).push(telaFake{nome: "primeira"})
	if got := antes.Top().Title(); got != "primeira" {
		t.Fatalf("topo = %q, quero primeira", got)
	}

	// Voltar e empilhar outra coisa: sem a cópia, este push escreve no mesmo
	// array que `antes` ainda referencia, e o topo dela muda sozinho.
	depois := antes.pop().push(telaFake{nome: "segunda"})

	if got := antes.Top().Title(); got != "primeira" {
		t.Errorf("a pilha foi compartilhada: o topo da versão antiga virou %q", got)
	}
	if got := depois.Top().Title(); got != "segunda" {
		t.Errorf("topo da versão nova = %q, quero segunda", got)
	}
}
