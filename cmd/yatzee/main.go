// Command yatzee é o jogo de dados no terminal. Sem argumento joga local; com
// "serve" sobe o servidor SSH.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/afonsodesenv/yatzee-terminal/internal/app"
)

func main() {
	ctx := context.Background()

	var err error
	switch {
	case len(os.Args) > 1 && os.Args[1] == "serve":
		err = app.Serve(ctx)
	case len(os.Args) > 1:
		fmt.Fprintf(os.Stderr, "uso: yatzee [serve]\n")
		os.Exit(2)
	default:
		err = app.PlayLocal(ctx)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "erro: %v\n", err)
		os.Exit(1)
	}
}
