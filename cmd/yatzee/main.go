// Command yatzee é o jogo de dados no terminal. Sem argumento joga local; com
// "serve" sobe o servidor SSH.
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
	// SIGTERM e ctrl+c cancelam o contexto, e é isso que torna alcançável o
	// desligamento gracioso do servidor: sem este fio, a goroutine que ouve
	// ctx.Done() em Serve nunca dispara, e um systemctl stop ou docker stop
	// espera o prazo inteiro antes de matar o processo à força.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var err error
	switch {
	case len(os.Args) > 1 && os.Args[1] == "serve":
		err = app.Serve(ctx)
	case len(os.Args) > 1:
		fmt.Fprintln(os.Stderr, "uso: yatzee [serve]")
		os.Exit(2)
	default:
		err = app.PlayLocal(ctx)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "erro: %v\n", err)
		os.Exit(1)
	}
}
