// Command yatzee é o jogo de dados no terminal.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/afonsodesenv/yatzee-terminal/internal/app"
)

func main() {
	// SIGTERM e ctrl+c cancelam o contexto. Serve para o jogo encerrar e
	// devolver o terminal ao normal quando alguém mata o processo, em vez de
	// deixar o alt-screen sujo.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if len(os.Args) > 1 {
		fmt.Fprintln(os.Stderr, "uso: yatzee")
		os.Exit(2)
	}

	if err := app.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "erro: %v\n", err)
		os.Exit(1)
	}
}
