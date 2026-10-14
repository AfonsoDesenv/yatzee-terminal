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

O jogo é um binário único sem dependência de runtime. Para compilá-lo você
precisa de Go 1.27 ou mais novo — depois disso, só o executável.

### Já tem o executável

```powershell
.\bin\yatzee.exe    # Windows
```

```bash
./bin/yatzee        # Linux e macOS
```

### Compilar no Windows

Precisa do Go instalado no Windows (`winget install GoLang.Go`, ou de
<https://go.dev/dl/>). Depois:

```powershell
go build -o bin\yatzee.exe .\cmd\yatzee
.\bin\yatzee.exe
```

### Compilar no Linux, macOS ou WSL

```bash
go build -o bin/yatzee ./cmd/yatzee
./bin/yatzee
```

Ou direto, sem gerar arquivo: `go run ./cmd/yatzee`.

### Nota sobre este ambiente de desenvolvimento

Neste repositório o Go está instalado **dentro do WSL**, em
`~/.local/go/bin/go`, e não no lado Windows — então `go` não funciona no
PowerShell aqui. Os alvos do `make` assumem esse caminho. Para gerar o `.exe`
a partir do WSL:

```bash
make build-windows
```

```bash
make run           # joga (dentro do WSL)
make test          # go test ./...
make test-race     # idem com detector de corrida (precisa de compilador C)
make build         # binario linux em bin/yatzee
make build-windows # binario windows em bin/yatzee.exe
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
