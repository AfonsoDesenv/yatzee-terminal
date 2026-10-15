package app

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/regras"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui/game"
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
	// A tela desenha dentro da moldura, então recebe o tamanho já descontado
	// da borda: duas colunas e duas linhas.
	queroW, queroH := ui.DefaultWidth-2, ui.DefaultHeight-2
	if medida.w != queroW || medida.h != queroH {
		t.Errorf("a tela recebeu (%d, %d), quero (%d, %d)", medida.w, medida.h, queroW, queroH)
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

// A moldura existe para o jogador ver onde a tela acaba, e tem de ter
// exatamente o tamanho do terminal: uma borda maior vaza e quebra o desenho.
func TestViewDesenhaAMolduraDoTamanhoDaTela(t *testing.T) {
	m := passo(NewRoot(telaFake{nome: "menu"}), tea.WindowSizeMsg{Width: 60, Height: 20})

	linhas := strings.Split(m.View().Content, "\n")
	if len(linhas) != 20 {
		t.Fatalf("a moldura tem %d linhas, quero 20", len(linhas))
	}
	for i, l := range linhas {
		if n := len([]rune(l)); n != 60 {
			t.Errorf("linha %d tem %d colunas, quero 60", i, n)
		}
	}
	if !strings.HasPrefix(linhas[0], ".---") {
		t.Errorf("topo da moldura = %q", linhas[0])
	}
	if !strings.HasPrefix(linhas[len(linhas)-1], "'---") {
		t.Errorf("base da moldura = %q", linhas[len(linhas)-1])
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

// A tela empilhada no meio da partida nunca recebe o WindowSizeMsg do arranque
// nem o da janela parada. Sem a raiz entregar o tamanho na hora do push, ela
// desenha com o padrão e só se encaixa quando a pessoa redimensiona o
// terminal — que foi exatamente o que apareceu na tela do jogo.
func TestTelaEmpilhadaJaNasceComOTamanhoDaTela(t *testing.T) {
	m := NewRoot(telaFake{nome: "menu"})
	m = passo(m, tea.WindowSizeMsg{Width: 120, Height: 40})

	var medida telaMedidora
	m, _ = m.pushWith(medida)

	topo, ok := m.Top().(telaMedidora)
	if !ok {
		t.Fatalf("topo = %T, quero telaMedidora", m.Top())
	}
	w, h := m.Size()
	if topo.w != w-2 || topo.h != h-2 {
		t.Errorf("a tela nova nasceu com %dx%d, quero %dx%d (a tela menos a moldura)",
			topo.w, topo.h, w-2, h-2)
	}
}

// O menu pede a lista de desafios, e ela entra na pilha por cima do menu.
func TestDesafioAbreALista(t *testing.T) {
	m := NewRoot(telaFake{nome: "menu"})
	m = passo(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	m = passo(m, ui.OpenChallengesMsg{})

	if m.Depth() != 2 {
		t.Fatalf("pilha com %d telas, quero 2 (menu e desafios)", m.Depth())
	}
	if got := m.Top().Title(); got != "desafios" {
		t.Errorf("tela do topo = %q, quero \"desafios\"", got)
	}
}

// Escolher um desafio troca a lista pela partida, em vez de empilhar por cima:
// voltar de dentro do jogo leva ao menu, igual a uma partida normal.
func TestEscolherODesafioTrocaAListaPelaPartida(t *testing.T) {
	m := NewRoot(telaFake{nome: "menu"})
	m = passo(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	m = passo(m, ui.OpenChallengesMsg{})
	m = passo(m, ui.StartChallengeMsg{Desafio: 0})

	if m.Depth() != 2 {
		t.Fatalf("pilha com %d telas, quero 2 (menu e jogo)", m.Depth())
	}
	if got := m.Top().Title(); got != "jogo" {
		t.Fatalf("tela do topo = %q, quero \"jogo\"", got)
	}

	// E voltar leva ao menu, nao a lista.
	m = passo(m, ui.BackMsg{})
	if got := m.Top().Title(); got != "menu" {
		t.Errorf("voltar levou a %q, quero \"menu\"", got)
	}
}

// A partida do desafio comeca com as condicoes dele, nao com as padrao.
func TestAPartidaDoDesafioComecaComAsCondicoes(t *testing.T) {
	for i, d := range regras.Desafios {
		m := NewRoot(telaFake{nome: "menu"})
		m = passo(m, tea.WindowSizeMsg{Width: 120, Height: 40})
		m = passo(m, ui.StartChallengeMsg{Desafio: i})

		jogo, ok := m.Top().(game.Screen)
		if !ok {
			t.Fatalf("desafio %q: topo = %T, quero game.Screen", d.Nome, m.Top())
		}
		esperada := d.Aplica(regras.Partida{})
		p := jogo.Partida()

		if p.Moedas != esperada.Moedas {
			t.Errorf("%s: moedas = %d, quero %d", d.Nome, p.Moedas, esperada.Moedas)
		}
		if len(p.Runas) != len(esperada.Runas) {
			t.Errorf("%s: %d runas, quero %d", d.Nome, len(p.Runas), len(esperada.Runas))
		}
	}
}

// Indice fora do catalogo nao faz nada: a mensagem pode chegar de qualquer
// tela, e derrubar o jogo por um indice ruim seria pior que ignorar.
func TestDesafioForaDoCatalogoNaoFazNada(t *testing.T) {
	m := NewRoot(telaFake{nome: "menu"})
	m = passo(m, tea.WindowSizeMsg{Width: 120, Height: 40})

	for _, i := range []int{-1, len(regras.Desafios), 999} {
		antes := m.Depth()
		m = passo(m, ui.StartChallengeMsg{Desafio: i})
		if m.Depth() != antes {
			t.Errorf("o indice %d mexeu na pilha: %d para %d", i, antes, m.Depth())
		}
	}
}
