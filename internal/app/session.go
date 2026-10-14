package app

import (
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"math/rand/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/player"
	"github.com/afonsodesenv/yatzee-terminal/internal/score"
)

// Session é um jogador conectado. O jogo local e o servidor SSH montam a mesma
// coisa, e nada aqui é compartilhado entre sessões — é o que permite N
// jogadores no mesmo processo.
type Session struct {
	Config   Config
	Identity player.Identity
	Players  player.Store
	Scores   score.Store
	// Rand é por sessão de propósito: o rand global é compartilhado entre
	// goroutines, então jogadores simultâneos interfeririam na sorte um do
	// outro e nenhuma partida seria reproduzível.
	Rand *rand.Rand
}

// NewSession monta uma sessão com gerador próprio, semeado por crypto/rand.
func NewSession(cfg Config, id player.Identity, players player.Store, scores score.Store) (*Session, error) {
	seed, err := randomSeed()
	if err != nil {
		return nil, err
	}
	return &Session{
		Config:   cfg,
		Identity: id,
		Players:  players,
		Scores:   scores,
		Rand:     rand.New(rand.NewPCG(seed[0], seed[1])),
	}, nil
}

func randomSeed() ([2]uint64, error) {
	var buf [16]byte
	if _, err := crand.Read(buf[:]); err != nil {
		return [2]uint64{}, fmt.Errorf("semear gerador da sessão: %w", err)
	}
	return [2]uint64{
		binary.LittleEndian.Uint64(buf[0:8]),
		binary.LittleEndian.Uint64(buf[8:16]),
	}, nil
}
