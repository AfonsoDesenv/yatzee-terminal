package game

import (
	"strings"

	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

// shopView é a fase de loja. Placeholder, como a rodada.
type shopView struct{ theme ui.Theme }

func (v shopView) render() string {
	b := &strings.Builder{}
	b.WriteString(v.theme.Title.Render("loja"))
	b.WriteString("\n\n")
	b.WriteString(v.theme.Dim.Render("os itens entram aqui"))
	b.WriteString("\n")
	return b.String()
}
