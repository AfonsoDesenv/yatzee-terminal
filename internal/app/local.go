package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/user"

	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/player"
	"github.com/afonsodesenv/yatzee-terminal/internal/score"
)

// PlayLocal roda o jogo no terminal da própria máquina.
func PlayLocal(ctx context.Context) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(cfg.DataDir, 0o700); err != nil {
		return fmt.Errorf("criar %s: %w", cfg.DataDir, err)
	}

	players := player.NewJSONStore(cfg.PlayersPath())
	scores := score.NewJSONStore(cfg.ScoresPath())

	id, err := localIdentityFor(ctx, osUsername(), players)
	if err != nil {
		return err
	}
	sess, err := NewSession(cfg, id, players, scores)
	if err != nil {
		return err
	}
	if id.Name != "" {
		// Falha em marcar presença não justifica recusar a partida.
		_ = players.Touch(ctx, id.ID)
	}

	// WithContext liga o cancelamento ao programa. Sem isto o ctx chega
	// apenas aos stores, e um SIGTERM não encerraria a partida local.
	prog := tea.NewProgram(NewRoot(sess, FirstScreen(sess)), tea.WithContext(ctx))
	if _, err := prog.Run(); err != nil {
		// Um desligamento que nós pedimos pelo contexto (SIGTERM) é saída
		// deliberada, não falha: o supervisor lê exit 1 como crash. Um kill por
		// outra causa continua sendo erro.
		if errors.Is(err, tea.ErrProgramKilled) && ctx.Err() != nil {
			return nil
		}
		return fmt.Errorf("rodar o jogo: %w", err)
	}
	return nil
}

// osUsername identifica o jogador local. Cai num nome fixo se o SO não
// informar: pior caso, todo mundo na máquina compartilha a identidade local.
func osUsername() string {
	if u, err := user.Current(); err == nil && u.Username != "" {
		return u.Username
	}
	if v := os.Getenv("USER"); v != "" {
		return v
	}
	return "jogador"
}
