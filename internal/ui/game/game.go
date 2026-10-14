// Package game tem a tela de partida e as fases por que ela passa.
package game

import (
	crand "crypto/rand"
	"encoding/binary"
	"math/rand/v2"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/regras"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui/fogo"
)

// Os relógios da jogada, separados de propósito: o passo da pontuação é lento
// o bastante para a pessoa ler o que cada etapa fez, e a chama precisa ser
// rápida o bastante para parecer fogo em vez de slideshow.
const (
	PassoPontuacao = 500 * time.Millisecond
	QuadroFogo     = 50 * time.Millisecond
	FogoNoPainel   = 1200 * time.Millisecond
	FogoNaTela     = 1800 * time.Millisecond
)

// Phase é onde a partida está. As fases de animação são estados próprios: é o
// que permite à tela ignorar o teclado enquanto a contagem acontece.
type Phase int

const (
	// PhaseRodada espera o jogador marcar, rerolar ou enviar.
	PhaseRodada Phase = iota
	// PhaseContando percorre a pontuação, um passo por vez.
	PhaseContando
	// PhaseComemorando queima a tela depois de bater a meta.
	PhaseComemorando
	// PhaseVitoria mostra o que a vitória pagou.
	PhaseVitoria
	// PhaseDerrota aparece quando as mãos acabam sem bater a meta.
	PhaseDerrota
	// PhaseLoja é o intervalo entre níveis.
	PhaseLoja
)

type (
	tiquePontuacao struct{}
	tiqueFogo      struct{}
	espalhaFogo    struct{}
	apagaFogo      struct{}
)

// Screen é a partida inteira: uma máquina de estados sobre uma regras.Partida.
type Screen struct {
	phase   Phase
	partida regras.Partida
	rnd     *rand.Rand

	anim  animacao
	chama *fogo.Fogo
	largo bool // a chama já se espalhou para a tela toda

	// Resultado da última jogada, para a tela mostrar enquanto conta.
	ultima regras.Pontuacao

	largura, altura int
	theme           ui.Theme
	keys            ui.Keys
}

// New começa uma partida nova no primeiro nível.
func New() Screen {
	var semente [16]byte
	_, _ = crand.Read(semente[:])
	rnd := rand.New(rand.NewPCG(
		binary.LittleEndian.Uint64(semente[0:8]),
		binary.LittleEndian.Uint64(semente[8:16]),
	))

	return Screen{
		phase:   PhaseRodada,
		partida: regras.NovaPartida(rnd),
		rnd:     rnd,
		largura: ui.DefaultWidth,
		altura:  ui.DefaultHeight,
		theme:   ui.NewTheme(),
		keys:    ui.DefaultKeys(),
	}
}

// Phase é a fase atual.
func (s Screen) Phase() Phase { return s.phase }

// Partida dá acesso ao estado do jogo, para teste.
func (s Screen) Partida() regras.Partida { return s.partida }

// Animando diz se a tela está no meio de uma animação e por isso muda.
func (s Screen) Animando() bool {
	return s.phase == PhaseContando || s.phase == PhaseComemorando
}

// Title identifica a tela.
func (s Screen) Title() string { return "jogo" }

// Init não precisa de nada.
func (s Screen) Init() tea.Cmd { return nil }

// Update roteia teclas e tiques. A ordem importa: enquanto anima, o teclado é
// ignorado, porque mexer no estado no meio da sequência que ela percorre
// deixaria a tela inconsistente.
func (s Screen) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.largura, s.altura = msg.Width, msg.Height
		return s, nil

	case tiquePontuacao:
		return s.avanca()
	case espalhaFogo:
		return s.espalha()
	case tiqueFogo:
		return s.queima()
	case apagaFogo:
		if s.chama != nil {
			s.chama.Apagar()
		}
		return s, nil

	case tea.KeyPressMsg:
		if s.Animando() {
			return s, nil
		}
		return s.tecla(msg.String())
	}
	return s, nil
}

func (s Screen) tecla(k string) (ui.Screen, tea.Cmd) {
	switch s.phase {
	case PhaseVitoria:
		if s.keys.Matches(k, s.keys.Confirm) {
			s.phase = PhaseLoja
		}
		return s, nil

	case PhaseDerrota:
		if s.keys.Matches(k, s.keys.Confirm) {
			return s, ui.Back
		}
		return s, nil

	case PhaseLoja:
		if s.keys.Matches(k, s.keys.Confirm) {
			s.partida = s.partida.ProximoNivel(s.rnd)
			s.phase = PhaseRodada
		}
		return s, nil
	}

	// PhaseRodada
	switch {
	case s.keys.Matches(k, s.keys.Back):
		return s, ui.Back

	case s.keys.Matches(k, s.keys.Confirm):
		if !s.partida.PodeEnviar() {
			return s, nil
		}
		var pont regras.Pontuacao
		s.partida, pont = s.partida.Envia()
		s.ultima = pont
		s.anim = novaAnimacao(pont, s.partida.Dados)
		s.phase = PhaseContando
		return s, tique(PassoPontuacao, tiquePontuacao{})

	case k == "r":
		s.partida = s.partida.Rerola(s.rnd)
		return s, nil

	default:
		if len(k) == 1 && k[0] >= '1' && k[0] <= '9' {
			s.partida = s.partida.Marca(int(k[0] - '0'))
		}
	}
	return s, nil
}

func tique(d time.Duration, msg tea.Msg) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg { return msg })
}

// avanca aplica um passo da contagem e decide o que vem depois dela.
func (s Screen) avanca() (ui.Screen, tea.Cmd) {
	if s.phase != PhaseContando {
		return s, nil
	}
	s.anim = s.anim.avanca()
	if !s.anim.acabou() {
		return s, tique(PassoPontuacao, tiquePontuacao{})
	}

	switch {
	case s.partida.Venceu():
		s.phase = PhaseComemorando
		s.largo = false
		s.chama = fogo.Novo(larguraFogoPainel, alturaFogoPainel, s.rnd)
		return s, tea.Batch(
			tique(QuadroFogo, tiqueFogo{}),
			tique(FogoNoPainel, espalhaFogo{}),
		)
	case s.partida.Perdeu():
		s.phase = PhaseDerrota
		return s, nil
	default:
		// Ainda há mão: rola os dados de novo para a jogada seguinte.
		s.partida = s.partida.ProximaMao(s.rnd)
		s.phase = PhaseRodada
		return s, nil
	}
}

const (
	larguraFogoPainel = 34
	alturaFogoPainel  = 11
)

// espalha troca a fogueira do painel por uma do tamanho da tela.
func (s Screen) espalha() (ui.Screen, tea.Cmd) {
	if s.phase != PhaseComemorando || s.largo {
		return s, nil
	}
	s.largo = true
	s.chama = fogo.Novo(s.largura, s.altura-1, s.rnd)
	return s, tique(FogoNaTela, apagaFogo{})
}

// queima avança um quadro da chama e abre o modal quando ela morre.
func (s Screen) queima() (ui.Screen, tea.Cmd) {
	if s.phase != PhaseComemorando || s.chama == nil {
		return s, nil
	}
	s.chama.Passo()
	if !s.chama.Vivo() {
		s.chama = nil
		s.phase = PhaseVitoria
		return s, nil
	}
	return s, tique(QuadroFogo, tiqueFogo{})
}

// View desenha a fase atual com a barra de ajuda no rodapé.
func (s Screen) View() string {
	altura := s.altura - 1
	var corpo string
	var dicas []string

	switch s.phase {
	case PhaseVitoria:
		corpo = s.modalVitoria(s.largura, altura)
		dicas = []string{"enter continua"}
	case PhaseDerrota:
		corpo = s.modalDerrota(s.largura, altura)
		dicas = []string{"enter volta ao menu"}
	case PhaseLoja:
		corpo = s.loja(s.largura, altura)
		dicas = []string{"enter continua"}
	case PhaseComemorando:
		corpo = s.comemoracao(s.largura, altura)
		dicas = []string{"meta batida!"}
	case PhaseContando:
		corpo = s.mesa(s.largura, altura)
		dicas = []string{"contando a jogada..."}
	default:
		corpo = s.mesa(s.largura, altura)
		dicas = s.dicasRodada()
	}

	return strings.Join([]string{corpo, ui.Centraliza(s.keys.Help(dicas), s.largura)}, "\n")
}

func (s Screen) dicasRodada() []string {
	d := []string{"1-5 marca"}
	if s.partida.PodeRerolar() {
		d = append(d, "r reroll")
	}
	return append(d, "enter envia", "esc menu")
}
