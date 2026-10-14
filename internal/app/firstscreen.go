package app

import (
	"context"
	"fmt"

	"github.com/afonsodesenv/yatzee-terminal/internal/player"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui/menu"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui/name"
)

// FirstScreen escolhe por onde a sessão começa: quem já tem nome não é
// perguntado de novo.
func FirstScreen(sess *Session) ui.Screen {
	if sess.Identity.Name == "" {
		return name.New(sess.Identity, sess.Players)
	}
	return menu.New(sess.Identity.Name)
}

// localIdentityFor monta a identidade do jogo local a partir do usuário do SO.
func localIdentityFor(ctx context.Context, osUser string, store player.Store) (player.Identity, error) {
	id := player.Identity{
		ID:   "local:" + osUser,
		Kind: player.KindLocal,
	}
	p, ok, err := store.Get(ctx, id.ID)
	if err != nil {
		return player.Identity{}, fmt.Errorf("buscar identidade local: %w", err)
	}
	if ok {
		id.Name = p.Name
	}
	return id, nil
}
