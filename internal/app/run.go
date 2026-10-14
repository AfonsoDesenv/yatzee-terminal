package app

import (
	"context"
	"errors"
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/ui/menu"
)

// Run roda o jogo no terminal, começando no menu.
func Run(ctx context.Context) error {
	// WithContext liga o cancelamento ao programa, para um SIGTERM encerrar a
	// partida e devolver o terminal em vez de deixar o alt-screen sujo.
	prog := tea.NewProgram(NewRoot(menu.New()), tea.WithContext(ctx))
	if _, err := prog.Run(); err != nil {
		// Um desligamento que nós pedimos pelo contexto é saída deliberada, não
		// falha: quem chamou lê exit 1 como crash. Um kill por outra causa
		// continua sendo erro.
		if errors.Is(err, tea.ErrProgramKilled) && ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("rodar o jogo: %w", err)
	}
	return nil
}
