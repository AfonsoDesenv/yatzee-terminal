# Yatzee Terminal

Jogo de dados no terminal, com progressão e loja inspiradas em Balatro.

## Estado atual

A estrutura está de pé; as mecânicas do jogo não. O que funciona hoje:

- Menu com novo jogo e sair
- Tela de partida com duas fases, rodada e loja, alternadas por `tab`
- Navegação de volta ao menu

O que **ainda não existe**: os dados, a pontuação e os itens da loja. As duas
fases são telas com o texto dizendo o que vem ali. O jogo não grava nada em
disco e não acessa a rede.

## Rodar

Precisa de Go 1.27 ou mais novo.

```bash
go run ./cmd/yatzee
```

Para gerar um binário:

```bash
go build -o bin/yatzee ./cmd/yatzee
./bin/yatzee
```

O `Makefile` faz o mesmo, mas usa `$(HOME)/.local/go/bin/go`, que é onde o Go
está instalado no ambiente de desenvolvimento deste projeto. Se o seu Go
estiver em outro lugar, use os comandos acima direto.

```bash
make run        # joga
make test       # go test ./...
make test-race  # idem com detector de corrida (precisa de compilador C)
make build      # binario em bin/yatzee
```

## Teclas

| Tecla | Efeito |
|---|---|
| `↑`/`↓`, `k`/`j` | Move a seleção no menu |
| `enter` | Confirma |
| `tab` | Alterna rodada e loja |
| `esc` | Volta ao menu |
| `q` | Sai (no menu) |
| `ctrl+c` | Sai de qualquer lugar |

## Documentação

- Desenho atual: [docs/superpowers/specs/2026-10-14-jogo-local-design.md](docs/superpowers/specs/2026-10-14-jogo-local-design.md)
- O desenho anterior, de um servidor SSH com identidade por chave pública e
  placar persistido, está em
  [docs/superpowers/specs/2026-10-13-yatzee-terminal-estrutura-design.md](docs/superpowers/specs/2026-10-13-yatzee-terminal-estrutura-design.md).
  Essa ideia foi abandonada; o código correspondente é recuperável no commit
  `5bc34ea`.
