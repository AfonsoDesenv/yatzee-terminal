package game

import (
	"strings"

	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

// roundView é a fase de rodadas. Placeholder: as mecânicas entram depois, e é
// este arquivo que vai crescer.
type roundView struct{ theme ui.Theme }

func (v roundView) render() string {
	b := &strings.Builder{}
	b.WriteString(v.theme.Title.Render("rodada"))
	b.WriteString("\n\n")
	b.WriteString(v.theme.Dim.Render("as mecânicas de dados entram aqui"))
	b.WriteString("\n")
	return b.String()
}
