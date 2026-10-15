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

// nomeadas são as teclas que não são um caractere digitável.
var nomeadas = map[string]rune{
	"enter": tea.KeyEnter,
	"esc":   tea.KeyEscape,
	"space": tea.KeySpace,
	"left":  tea.KeyLeft,
	"right": tea.KeyRight,
	"up":    tea.KeyUp,
	"down":  tea.KeyDown,
}

func tecla(s Screen, k string) (Screen, tea.Cmd) {
	if code, ok := nomeadas[k]; ok {
		return aplica(s, tea.KeyPressMsg{Code: code})
	}
	return aplica(s, tea.KeyPressMsg{Code: rune(k[0]), Text: k})
}

// contaTudo roda todos os tiques da contagem, inclusive o que a fecha. A
// contagem só sai de PhaseContando no fecho: o último passo fica aceso pelo
// tempo dele, como todos os outros.
func contaTudo(s Screen) Screen {
	for i := 0; i < 30 && s.Phase() == PhaseContando && !s.anim.acabou(); i++ {
		s, _ = aplica(s, tiquePontuacao{})
	}
	if s.Phase() == PhaseContando {
		s, _ = aplica(s, fechaContagem{})
	}
	return s
}

// venceAgora força uma vitória: dá pontos suficientes e envia.
func venceAgora(s Screen) Screen {
	s.partida.Alvo = 1 // qualquer jogada basta
	s, _ = tecla(s, "enter")
	s = contaTudo(s)
	s, _ = aplica(s, abreVitoria{})
	return s
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

// A combinação entra PRIMEIRO, fixando fichas e multiplicador, e só depois os
// dados somam. A ordem não é estética: com o multiplicador definido já no
// primeiro passo, o total cresce de verdade a cada tique e pode cruzar a meta
// no meio da contagem. Na ordem inversa a conta fica rasteira até o fim.
func TestAContagemComecaPelaCombinacao(t *testing.T) {
	s := New()
	s.partida.Dados = regras.Mao{3, 3, 3, 6, 6} // full house, soma 21
	s, _ = tecla(s, "enter")

	s, _ = aplica(s, tiquePontuacao{})
	if got := s.anim.mult; got != 5 {
		t.Errorf("mult depois do primeiro passo = %d, quero 5 (a combinação vem primeiro)", got)
	}
	if got := s.anim.fichas; got != 5 {
		t.Errorf("fichas depois do primeiro passo = %d, quero 5", got)
	}

	// Os cinco dados somam por cima, um por tique.
	for i := 0; i < regras.DadosPorMao; i++ {
		s, _ = aplica(s, tiquePontuacao{})
	}
	if got := s.anim.fichas; got != 26 {
		t.Errorf("fichas com os dados somados = %d, quero 26", got)
	}
	if got := s.anim.mult; got != 5 {
		t.Errorf("mult = %d, os dados não deviam mexer nele", got)
	}
}

// As runas disparam depois dos dados, e só as que têm efeito viram passo.
func TestAsRunasEntramDepoisDosDados(t *testing.T) {
	c, _ := regras.RunaPorLetra("C") // cada dado 3: +10 fichas
	l, _ := regras.RunaPorLetra("L") // cada dado 6: +2 mult
	a, _ := regras.RunaPorLetra("A") // cada dado 1: não dispara aqui

	s := New()
	s.partida.Dados = regras.Mao{3, 3, 3, 6, 6}
	s.partida.Runas = []regras.Runa{c, l, a}
	s, _ = tecla(s, "enter")

	// A runa inerte não vira passo. As que disparam contam um passo por dado:
	// combinação + 5 dados + C nos três 3 + L nos dois 6.
	if got := len(s.anim.passos); got != 1+5+3+2 {
		t.Errorf("a contagem tem %d passos, quero %d (a runa sem efeito não entra)", got, 1+5+3+2)
	}
	for _, p := range s.anim.passos {
		if strings.Contains(p.rotulo, "runa A") {
			t.Error("a runa A entrou na contagem sem disparar")
		}
	}

	s = contaTudo(s)
	// Full house (5 base, 5 mult) + 21 dos dados; C soma 10 por dado 3 (tres)
	// e L soma 2 por dado 6 (dois).
	if quero := (5 + 21 + 3*10) * (5 + 2*2); s.Partida().Pontos != int64(quero) {
		t.Errorf("pontos = %d, quero %d", s.Partida().Pontos, quero)
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

func TestBaterAMetaAbreAVitoriaComATelaEmChamas(t *testing.T) {
	s := venceAgora(New())

	if s.Phase() != PhaseVitoria {
		t.Fatalf("fase = %v depois de bater a meta, quero PhaseVitoria", s.Phase())
	}
	if !s.Pegando() {
		t.Error("a vitória não acendeu a tela")
	}
	v := s.View()
	if !strings.Contains(v, "META BATIDA") {
		t.Error("o modal de vitória não apareceu")
	}
	if !strings.ContainsAny(v, "@%#*") {
		t.Error("não há chama na tela da vitória")
	}
}

// As moedas entram quando a vitória é anunciada, não ao sair da loja: é o
// saldo que a loja mostra e vai usar para vender.
func TestAVitoriaCreditaAsMoedas(t *testing.T) {
	s := New()
	s.partida.Alvo = 1
	antes := s.Partida().Moedas

	s, _ = tecla(s, "enter")
	s = contaTudo(s)
	premiada := s.Partida()
	s, _ = aplica(s, abreVitoria{})

	esperado := antes + premiada.Recompensa()
	if got := s.Partida().Moedas; got != esperado {
		t.Errorf("moedas = %d depois da vitória, quero %d", got, esperado)
	}

	// E a loja mostra esse saldo, não o de antes.
	s, _ = tecla(s, "enter")
	if s.Phase() != PhaseLoja {
		t.Fatalf("fase = %v, quero PhaseLoja", s.Phase())
	}
	if !strings.Contains(s.View(), ui.Numero(int64(esperado))) {
		t.Errorf("a loja não mostra o saldo de %d moedas", esperado)
	}
}

// O modal é opaco: o fogo queima em volta, não por dentro, senão os números da
// recompensa ficariam ilegíveis justo na hora de lê-los.
func TestOModalDeVitoriaNaoPegaFogo(t *testing.T) {
	s := venceAgora(New())
	for i := 0; i < 30; i++ {
		s, _ = aplica(s, tiqueFogo{})
	}

	var achou bool
	for _, l := range strings.Split(s.View(), "\n") {
		if !strings.Contains(l, "ganhou") && !strings.Contains(l, "total") {
			continue
		}
		achou = true

		// O fogo em volta do modal é o esperado; o que não pode é chama
		// dentro dele. Recorta o miolo entre as bordas do painel.
		ini := strings.Index(l, "|")
		fim := strings.LastIndex(l, "|")
		if ini < 0 || fim <= ini {
			t.Errorf("linha do modal sem bordas: %q", l)
			continue
		}
		if miolo := l[ini+1 : fim]; strings.ContainsAny(miolo, "@%#") {
			t.Errorf("chama dentro do modal: %q", miolo)
		}
	}
	if !achou {
		t.Fatal("o modal de vitória não apareceu na tela")
	}
}

// O painel da meta pega fogo durante a contagem, no instante em que a conta
// cruza o alvo — não no fim dela.
func TestOPainelPegaFogoAoCruzarAMetaDurandoAContagem(t *testing.T) {
	s := New()
	s.partida.Dados = regras.Mao{6, 6, 6, 6, 6}
	// Alvo baixo de propósito: a conta cruza já no primeiro dado, muito antes
	// do último passo. Com um alvo alto a travessia coincidiria com o fim da
	// contagem, e o teste não distinguiria "fogo durante" de "fogo no fim".
	s.partida.Alvo = 1
	s, _ = tecla(s, "enter")

	if s.Pegando() {
		t.Fatal("pegou fogo antes de contar nada")
	}

	s, _ = aplica(s, tiquePontuacao{})
	if !s.Pegando() {
		t.Error("a conta cruzou a meta no primeiro passo e o painel não pegou fogo")
	}
	if s.anim.acabou() {
		t.Fatal("a contagem acabou num passo só: o teste não prova nada")
	}
	if s.Phase() != PhaseContando {
		t.Errorf("fase = %v, o fogo devia pegar ainda durante a contagem", s.Phase())
	}

	// E a vitoria so abre depois, para o fogo no painel ter duracao visivel.
	s, _ = aplica(s, abreVitoria{})
	if s.Phase() != PhaseVitoria {
		t.Errorf("fase = %v depois de abreVitoria, quero PhaseVitoria", s.Phase())
	}
}

// E o número do jogador sobe junto com os passos, em vez de pular de uma vez.
func TestOAcumuladoSobeDuranteAContagem(t *testing.T) {
	s := New()
	s.partida.Dados = regras.Mao{3, 3, 3, 6, 6}
	antes := s.parcial()
	s, _ = tecla(s, "enter")

	if got := s.parcial(); got != antes {
		t.Errorf("parcial = %d logo após enviar, quero %d", got, antes)
	}
	s, _ = aplica(s, tiquePontuacao{})
	if got := s.parcial(); got <= antes {
		t.Errorf("parcial = %d depois de um passo, não subiu de %d", got, antes)
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

	s, _ = tecla(s, "space")
	if s.Phase() != PhaseLoja {
		t.Fatalf("fase = %v depois de continuar, quero PhaseLoja", s.Phase())
	}

	// A loja abre com o foco numa runa; continuar é um botão à parte.
	s, _ = tecla(s, "left")
	s, _ = tecla(s, "space")
	if s.Phase() != PhaseRodada {
		t.Fatalf("fase = %v depois da loja, quero PhaseRodada", s.Phase())
	}
	p := s.Partida()
	_ = recompensa
	if p.Nivel != nivel+1 {
		t.Errorf("nível = %d, quero %d", p.Nivel, nivel+1)
	}
	if p.Alvo != alvo*2 {
		t.Errorf("alvo = %d, quero o dobro de %d", p.Alvo, alvo)
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
