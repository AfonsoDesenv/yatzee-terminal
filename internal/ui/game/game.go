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

	// Depois que a conta cruza a meta, o painel queima por um tempo antes de
	// a vitoria abrir. Sem esta pausa o fogo "durante o calculo" nao teria
	// duracao nenhuma: o multiplicador so entra no ultimo passo, entao a
	// travessia acontece junto com o fim da contagem.
	FogoNoPainel = 1400 * time.Millisecond
	DuracaoFogo  = 2500 * time.Millisecond
)

// Phase é onde a partida está. As fases de animação são estados próprios: é o
// que permite à tela ignorar o teclado enquanto a contagem acontece.
type Phase int

const (
	// PhaseRodada espera o jogador marcar, rerolar ou enviar.
	PhaseRodada Phase = iota
	// PhaseContando percorre a pontuação, um passo por vez.
	PhaseContando
	// PhaseVitoria mostra o que a vitória pagou, sobre a tela em chamas.
	PhaseVitoria
	// PhaseDerrota aparece quando as mãos acabam sem bater a meta.
	PhaseDerrota
	// PhaseLoja é o intervalo entre níveis.
	PhaseLoja
)

type (
	tiquePontuacao struct{}
	tiqueFogo      struct{}
	abreVitoria    struct{}
	apagaFogo      struct{}
	// fechaContagem encerra a contagem um tique depois do último passo. Sem
	// essa espera o último passo não teria duração nenhuma: a tela trocaria
	// de fase no mesmo instante em que ele entrou, e como as runas contam no
	// fim, a runa nunca chegaria a acender.
	fechaContagem struct{}
)

// Screen é a partida inteira: uma máquina de estados sobre uma regras.Partida.
type Screen struct {
	phase   Phase
	partida regras.Partida
	rnd     *rand.Rand

	anim animacao
	// pontosAntes é o acumulado antes da jogada em contagem. A tela mostra
	// pontosAntes mais o que a contagem já rendeu, para o número subir junto
	// com os passos em vez de pular de uma vez no fim.
	pontosAntes int64

	chama *fogo.Fogo

	// vitrine são as runas à venda no nível atual.
	vitrine []regras.Runa
	// sel é onde o foco da loja está: focoContinuar ou uma das vitrines. Só a
	// loja usa, mas mora aqui porque Screen é o estado inteiro da partida.
	sel int

	largura, altura int
	theme           ui.Theme
	keys            ui.Keys
}

// New começa uma partida nova no primeiro nível.
func New() Screen { return nova(regras.NovaPartida) }

// NovoDesafio começa uma partida com as condições de um desafio.
func NovoDesafio(d regras.Desafio) Screen {
	return nova(func(rnd *rand.Rand) regras.Partida {
		return regras.NovaPartidaDesafio(d, rnd)
	})
}

// nova monta a tela em volta de uma partida qualquer. O gerador nasce aqui e
// é passado a quem começa a partida, para a partida normal e o desafio
// dividirem o mesmo caminho.
func nova(comeca func(*rand.Rand) regras.Partida) Screen {
	var semente [16]byte
	_, _ = crand.Read(semente[:])
	rnd := rand.New(rand.NewPCG(
		binary.LittleEndian.Uint64(semente[0:8]),
		binary.LittleEndian.Uint64(semente[8:16]),
	))

	return Screen{
		phase:   PhaseRodada,
		partida: comeca(rnd),
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

// Pegando diz se há chama acesa agora.
func (s Screen) Pegando() bool { return s.chama != nil }

// acende diz qual dado e qual runa estão acesos agora. Fora da contagem nada
// acende, e semItem nunca casa com posição nenhuma.
func (s Screen) acende() (dado, runa int) {
	if s.phase != PhaseContando {
		return semItem, semItem
	}
	return s.anim.causa()
}

// chamaPintada é a chama com a cor do tema. A tradução de calor em cor mora
// aqui, e não no pacote fogo: ele simula, e continua sem saber o que é cor.
func (s Screen) chamaPintada() []string {
	return ui.PintaRampa(s.chama.Linhas(), fogo.Rampa, s.theme.Fogo)
}

// Title identifica a tela.
func (s Screen) Title() string { return "jogo" }

// Init não precisa de nada.
func (s Screen) Init() tea.Cmd { return nil }

// Update roteia teclas e tiques. Enquanto conta, o teclado é ignorado: mexer
// no estado no meio da sequência que ela percorre deixaria a tela
// inconsistente.
func (s Screen) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.largura, s.altura = msg.Width, msg.Height
		return s, nil

	case tiquePontuacao:
		return s.avanca()

	case fechaContagem:
		return s.fecha()

	case abreVitoria:
		if s.phase != PhaseContando {
			return s, nil
		}
		// A recompensa entra agora, nao ao sair da loja: e o saldo que a loja
		// mostra e usa.
		s.partida = s.partida.Premia()
		s.phase = PhaseVitoria
		// O fogo do painel da lugar a um que toma a tela inteira, por tras do
		// modal.
		s.chama = fogo.Novo(s.largura, s.altura-1, s.rnd)
		return s, tea.Batch(
			tique(QuadroFogo, tiqueFogo{}),
			tique(DuracaoFogo, apagaFogo{}),
		)

	case tiqueFogo:
		if s.chama == nil {
			return s, nil
		}
		s.chama.Passo()
		if !s.chama.Vivo() {
			s.chama = nil
			return s, nil
		}
		return s, tique(QuadroFogo, tiqueFogo{})

	case apagaFogo:
		if s.chama != nil {
			s.chama.Apagar()
		}
		return s, nil

	case tea.KeyPressMsg:
		if s.phase == PhaseContando {
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
			s.chama = nil
			s.vitrine = sorteiaVitrine(s.rnd)
			// Abre na primeira runa, não no botão de continuar: a loja existe
			// para olhar o que está à venda, e a caixa de texto já nasce
			// dizendo o que a runa em foco faz.
			s.sel = focoPrimeiraRuna
		}
		return s, nil

	case PhaseDerrota:
		if s.keys.Matches(k, s.keys.Confirm) {
			return s, ui.Back
		}
		return s, nil

	case PhaseLoja:
		// As quatro setas andam pelo mesmo anel. O botão fica à esquerda das
		// runas na tela, então esquerda/direita é o movimento natural; cima e
		// baixo seguem junto para nenhuma seta ficar morta na mão de quem
		// estiver com os dedos no outro lado do teclado.
		switch {
		case s.keys.Matches(k, s.keys.Left), s.keys.Matches(k, s.keys.Up):
			s = s.move(-1)
		case s.keys.Matches(k, s.keys.Right), s.keys.Matches(k, s.keys.Down):
			s = s.move(1)
		case s.keys.Matches(k, s.keys.Confirm):
			if s.sel == focoContinuar {
				s.partida = s.partida.ProximoNivel(s.rnd)
				s.phase = PhaseRodada
				return s, nil
			}
			s = s.compra(s.sel - focoPrimeiraRuna)
		}
		return s, nil
	}

	switch {
	case s.keys.Matches(k, s.keys.Back):
		return s, ui.Back

	case s.keys.Matches(k, s.keys.Confirm):
		if !s.partida.PodeEnviar() {
			return s, nil
		}
		s.pontosAntes = s.partida.Pontos
		var pont regras.Pontuacao
		s.partida, pont = s.partida.Envia()
		s.anim = novaAnimacao(pont, s.partida.Dados, s.partida.Runas)
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

// parcial é quanto o jogador tem enquanto a contagem corre: o acumulado de
// antes mais o que a jogada já rendeu até aqui.
func (s Screen) parcial() int64 {
	if s.phase != PhaseContando {
		return s.partida.Pontos
	}
	return s.pontosAntes + int64(s.anim.fichas*s.anim.mult)
}

// cruzouAMeta diz se o que já foi contado basta para bater o alvo. É o gatilho
// do fogo no painel: ele pega no instante em que a conta passa da meta, e não
// no fim da contagem.
func (s Screen) cruzouAMeta() bool { return s.parcial() >= s.partida.Alvo }

// avanca aplica um passo da contagem e decide o que vem depois dela.
func (s Screen) avanca() (ui.Screen, tea.Cmd) {
	// Tique que chega com a contagem já no fim não faz nada: quem encerra é
	// fechaContagem, e repetir a resolução aqui marcaria mão duas vezes.
	if s.phase != PhaseContando || s.anim.acabou() {
		return s, nil
	}
	s.anim = s.anim.avanca()

	var cmds []tea.Cmd
	// O painel pega fogo assim que a conta cruza a meta, no meio da contagem.
	if s.chama == nil && s.cruzouAMeta() {
		s.chama = fogo.Novo(larguraFogoPainel, alturaFogoPainel, s.rnd)
		cmds = append(cmds, tique(QuadroFogo, tiqueFogo{}))
	}

	if !s.anim.acabou() {
		cmds = append(cmds, tique(PassoPontuacao, tiquePontuacao{}))
		return s, tea.Batch(cmds...)
	}

	// Acabaram os passos. A fase continua em PhaseContando até fechar, para o
	// último passo ficar aceso pelo tempo dele como todos os outros.
	if s.partida.Venceu() {
		// O painel em chamas já segura a tela: a vitória abre depois, e nesse
		// meio-tempo o último passo continua aceso.
		if s.chama == nil {
			s.chama = fogo.Novo(larguraFogoPainel, alturaFogoPainel, s.rnd)
			cmds = append(cmds, tique(QuadroFogo, tiqueFogo{}))
		}
		return s, tea.Batch(append(cmds, tique(FogoNoPainel, abreVitoria{}))...)
	}
	return s, tea.Batch(append(cmds, tique(PassoPontuacao, fechaContagem{}))...)
}

// fecha sai da contagem: derrota se acabaram as mãos, senão a mão seguinte.
// Partida ganha não passa por aqui — quem conduz é abreVitoria, e fechar
// adiantado pularia a vitória inteira.
func (s Screen) fecha() (ui.Screen, tea.Cmd) {
	if s.phase != PhaseContando || s.partida.Venceu() {
		return s, nil
	}
	s.chama = nil

	if s.partida.Perdeu() {
		s.phase = PhaseDerrota
		return s, nil
	}
	s.partida = s.partida.ProximaMao(s.rnd)
	s.phase = PhaseRodada
	return s, nil
}

const (
	larguraFogoPainel = 34
	alturaFogoPainel  = 11
)

// View desenha a fase atual com a barra de ajuda no rodapé.
func (s Screen) View() string {
	altura := s.altura - 1
	var corpo string
	var dicas []string

	switch s.phase {
	case PhaseVitoria:
		corpo = s.telaVitoria(s.largura, altura)
		dicas = []string{"espaco continua"}
	case PhaseDerrota:
		corpo = s.modalDerrota(s.largura, altura)
		dicas = []string{"espaco volta ao menu"}
	case PhaseLoja:
		corpo = s.loja(s.largura, altura)
		dicas = []string{"setas escolhem", "espaco confirma"}
	case PhaseContando:
		corpo = s.mesa(s.largura, altura)
		if s.cruzouAMeta() {
			dicas = []string{"meta batida!"}
		} else {
			dicas = []string{"contando a jogada..."}
		}
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
	return append(d, "espaco envia", "esc menu")
}

// PrecoRuna é quanto custa uma runa na loja. Valor único por enquanto: a
// diferença de poder entre as básicas é pequena, e preço variável só faria
// sentido com runas de forças diferentes.
const PrecoRuna = 5

// Os focos da loja, na ordem em que as setas passam por eles. O botão de
// continuar é o primeiro porque é o que fica mais à esquerda na tela: andar
// com a seta tem de bater com o que o olho vê.
const (
	focoContinuar    = 0
	focoPrimeiraRuna = 1
)

// focos é quantas paradas o anel tem.
func (s Screen) focos() int { return focoPrimeiraRuna + len(s.vitrine) }

// move anda com o foco, dando a volta nas pontas: com quatro paradas, parar
// na borda faria a pessoa voltar o caminho todo para chegar ao outro lado.
func (s Screen) move(passo int) Screen {
	n := s.focos()
	s.sel = ((s.sel+passo)%n + n) % n
	return s
}

// emFoco devolve a runa sob o foco. O bool é falso no botão de continuar, que
// não é runa nenhuma.
func (s Screen) emFoco() (regras.Runa, bool) {
	i := s.sel - focoPrimeiraRuna
	if i < 0 || i >= len(s.vitrine) {
		return regras.Runa{}, false
	}
	return s.vitrine[i], true
}

// sorteiaVitrine escolhe três runas distintas do catálogo.
func sorteiaVitrine(rnd *rand.Rand) []regras.Runa {
	baralho := append([]regras.Runa{}, regras.Catalogo...)
	rnd.Shuffle(len(baralho), func(i, j int) {
		baralho[i], baralho[j] = baralho[j], baralho[i]
	})
	return baralho[:3]
}

// compra leva a runa da vitrine, se houver saldo e espaço, e tira o item da
// prateleira para não ser comprado duas vezes.
func (s Screen) compra(i int) Screen {
	if i < 0 || i >= len(s.vitrine) {
		return s
	}
	r := s.vitrine[i]
	if r.Letra == "" || !s.partida.PodeComprar(PrecoRuna) {
		return s
	}
	s.partida = s.partida.Compra(r, PrecoRuna)

	nova := append([]regras.Runa{}, s.vitrine...)
	nova[i] = regras.Runa{} // vendida
	s.vitrine = nova
	return s
}
