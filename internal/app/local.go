package app

import (
	"context"
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

	if _, err := tea.NewProgram(NewRoot(sess, FirstScreen(sess))).Run(); err != nil {
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
