package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"net"
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/ssh"
	"charm.land/wish/v2"
	bm "charm.land/wish/v2/bubbletea"
	gossh "golang.org/x/crypto/ssh"

	"github.com/afonsodesenv/yatzee-terminal/internal/player"
	"github.com/afonsodesenv/yatzee-terminal/internal/score"
)

// IdentityFromKey transforma a chave pública do cliente em identidade. O
// fingerprint é estável por anos e atravessa proxy, ao contrário do IP.
func IdentityFromKey(key gossh.PublicKey) player.Identity {
	return player.Identity{
		ID:   "key:" + gossh.FingerprintSHA256(key),
		Kind: player.KindPubKey,
	}
}

// GuestIdentity monta uma identidade de sessão para quem conecta sem chave.
func GuestIdentity(rnd *rand.Rand) player.Identity {
	return player.Identity{
		ID:   fmt.Sprintf("guest:%016x", rnd.Uint64()),
		Kind: player.KindGuest,
	}
}

// Serve sobe o servidor SSH. Cada cliente ganha uma sessão isolada: nada de
// estado compartilhado entre jogadores. Cancelar o contexto encerra o servidor
// e conta como parada deliberada: retorna nil, não erro.
func Serve(ctx context.Context) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	// Antes de subir: o WithHostKeyPath faz chmod no diretório pai, então ele
	// precisa existir e ser nosso.
	if err := os.MkdirAll(cfg.DataDir, 0o700); err != nil {
		return fmt.Errorf("criar %s: %w", cfg.DataDir, err)
	}

	players := player.NewJSONStore(cfg.PlayersPath())
	scores := score.NewJSONStore(cfg.ScoresPath())

	srv, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(cfg.SSHHost, cfg.SSHPort)),
		wish.WithHostKeyPath(cfg.SSHHostKey),
		// Qualquer chave serve: aqui ela é identidade, não autorização.
		wish.WithPublicKeyAuth(func(ssh.Context, ssh.PublicKey) bool { return true }),
		// Sem isto, quem não tem chave nem conecta.
		wish.WithKeyboardInteractiveAuth(func(ssh.Context, gossh.KeyboardInteractiveChallenge) bool {
			return true
		}),
		wish.WithMiddleware(
			bm.Middleware(func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
				model, err := sessionModel(s, cfg, players, scores)
				if err != nil {
					log.Printf("abrir sessão: %v", err)
					wish.Errorln(s, "não consegui abrir o jogo agora, tente de novo")
					return nil, nil
				}
				// MakeOptions não vai aqui: o middleware já a aplica, e
				// chamar de novo duplicaria WithInput/WithOutput na mesma
				// sessão.
				return model, nil
			}),
		),
	)
	if err != nil {
		return fmt.Errorf("montar servidor ssh: %w", err)
	}

	// Close não bloqueia esperando clientes: o desligamento nunca fica preso.
	// O canal evita vazar a goroutine se o servidor cair antes do cancelamento.
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			_ = srv.Close()
		case <-done:
		}
	}()

	log.Printf("yatzee ouvindo ssh em %s:%s", cfg.SSHHost, cfg.SSHPort)
	// Fechar o servidor de propósito não é erro. Comparar com a sentinela, não
	// com o texto da mensagem.
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		return fmt.Errorf("servir ssh: %w", err)
	}
	return nil
}

// sessionModel monta o model raiz para um cliente.
func sessionModel(s ssh.Session, cfg Config, players player.Store, scores score.Store) (tea.Model, error) {
	ctx := s.Context()

	var id player.Identity
	if key := s.PublicKey(); key != nil {
		id = IdentityFromKey(key)
	}

	sess, err := NewSession(cfg, id, players, scores)
	if err != nil {
		return nil, err
	}
	if id.ID == "" {
		// Sem chave: convidado, com ID sorteado pelo gerador da sessão.
		sess.Identity = GuestIdentity(sess.Rand)
	} else {
		p, ok, err := players.Get(ctx, id.ID)
		switch {
		case err != nil:
			// Não recusa a conexão por falha de leitura — é um jogo, não um
			// cofre — mas registra. Um players.json corrompido faria todo
			// jogador que volta ver a tela de nome outra vez, e sem este log
			// o operador não teria sinal nenhum de que isso está acontecendo.
			log.Printf("ler identidade %s: %v", id.ID, err)
		case ok:
			sess.Identity.Name = p.Name
			// Falha em marcar presença não justifica recusar o jogador.
			_ = players.Touch(ctx, id.ID)
		}
	}

	return NewRoot(sess, FirstScreen(sess)), nil
}
