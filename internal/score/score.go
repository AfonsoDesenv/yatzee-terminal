// Package score guarda o placar. Não conhece UI nem transporte.
package score

import (
	"context"
	"time"
)

// Entry é uma partida registrada. Guarda PlayerID e Name juntos de propósito:
// o placar continua legível mesmo que a identidade mude de nome algum dia.
type Entry struct {
	PlayerID string    `json:"player_id"`
	Name     string    `json:"name"`
	Points   int       `json:"points"`
	At       time.Time `json:"at"`
}

// Store persiste partidas.
type Store interface {
	// Top devolve as n melhores, da maior pontuação para a menor.
	Top(ctx context.Context, n int) ([]Entry, error)
	Save(ctx context.Context, e Entry) error
}
