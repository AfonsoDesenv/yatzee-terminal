package app

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/afonsodesenv/yatzee-terminal/internal/player"
	"github.com/afonsodesenv/yatzee-terminal/internal/score"
)

func sessaoDeTeste(t *testing.T) *Session {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{DataDir: dir}
	sess, err := NewSession(
		cfg,
		player.Identity{ID: "local:teste", Kind: player.KindLocal},
		player.NewJSONStore(filepath.Join(dir, "players.json")),
		score.NewJSONStore(filepath.Join(dir, "scores.json")),
	)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	return sess
}

func ctx() context.Context { return context.Background() }
