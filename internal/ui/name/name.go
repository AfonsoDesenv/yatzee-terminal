// Package name tem a tela que pergunta o nome na primeira vez.
package name

import (
	"context"
	"errors"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/player"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

// Screen pergunta o nome e só libera quando ele é válido e livre.
type Screen struct {
	id    player.Identity
	store player.Store
	input textinput.Model
	err   string
	theme ui.Theme
	keys  ui.Keys
}

// New monta a tela para uma identidade ainda sem nome.
func New(id player.Identity, store player.Store) Screen {
	ti := textinput.New()
	ti.Prompt = "> "
	ti.Placeholder = "seu-nome"
	ti.CharLimit = player.NameMaxLen
	ti.SetWidth(player.NameMaxLen + 4)
	ti.Focus()

	return Screen{
		id:    id,
		store: store,
		input: ti,
		theme: ui.NewTheme(),
		keys:  ui.DefaultKeys(),
	}
}

// Value é o texto atual do campo.
func (s Screen) Value() string { return s.input.Value() }

// Error é a mensagem de erro visível, vazia quando não há.
func (s Screen) Error() string { return s.err }

// Title identifica a tela.
func (s Screen) Title() string { return "nome" }

// Init pisca o cursor.
func (s Screen) Init() tea.Cmd { return textinput.Blink }

// Update trata o enter e delega o resto ao campo de texto.
func (s Screen) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		if s.keys.Matches(key.String(), s.keys.Confirm) {
			return s.confirm()
		}
	}

	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)

	// Baixa a caixa depois do campo processar, não interceptando a tecla:
	// assim texto colado também entra minúsculo, e a pessoa vê o nome real
	// enquanto digita em vez de descobrir a conversão no envio.
	if v := s.input.Value(); v != strings.ToLower(v) {
		s.input.SetValue(strings.ToLower(v))
	}
	return s, cmd
}

// confirm valida, reivindica e anuncia. Erro aqui não é falha do sistema: a
// tela mantém o texto e espera outro nome.
func (s Screen) confirm() (ui.Screen, tea.Cmd) {
	ctx := context.Background()

	name, err := player.CleanName(s.input.Value())
	if err != nil {
		s.err = player.NameRule()
		return s, nil
	}

	// Convidado não reserva nome, mas também não pode se passar por alguém
	// registrado no placar.
	if !s.id.Persistent() {
		taken, err := s.store.NameTaken(ctx, name)
		if err != nil {
			s.err = "não consegui verificar o nome agora"
			return s, nil
		}
		if taken {
			s.err = "esse nome já está em uso"
			return s, nil
		}
		s.err = ""
		return s, ui.NameSet(name)
	}

	if _, err := s.store.Claim(ctx, s.id.ID, name); err != nil {
		if errors.Is(err, player.ErrNameTaken) {
			s.err = "esse nome já está em uso"
		} else {
			s.err = "não consegui gravar o nome agora"
		}
		return s, nil
	}
	s.err = ""
	return s, ui.NameSet(name)
}

// View desenha o campo, a regra e o erro.
func (s Screen) View() string {
	b := &strings.Builder{}
	b.WriteString(s.theme.Title.Render("bem-vindo ao yatzee"))
	b.WriteString("\n\ncomo você quer ser chamado?\n\n")
	b.WriteString(s.input.View())
	b.WriteString("\n\n")
	b.WriteString(s.theme.Dim.Render(player.NameRule()))
	if s.err != "" {
		b.WriteString("\n")
		b.WriteString(s.theme.Error.Render(s.err))
	}
	b.WriteString("\n\n")
	b.WriteString(s.keys.Help([]string{"enter confirma", "ctrl+c sai"}))
	return s.theme.Panel.Render(b.String())
}
