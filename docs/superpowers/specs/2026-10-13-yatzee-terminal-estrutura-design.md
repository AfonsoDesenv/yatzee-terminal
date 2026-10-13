# Yatzee Terminal — Estrutura Base

**Data:** 2026-10-13
**Status:** Aprovado para implementação

## Objetivo

Jogo de dados estilo Yahtzee com progressão e loja inspiradas em Balatro, jogável
no terminal. Esta entrega monta **apenas a estrutura de telas e navegação**. As
mecânicas do jogo são a próxima conversa e não entram aqui.

O jogo precisa ser jogável remotamente no futuro (servidor próprio atrás de
Cloudflare), sem que o jogador instale nada. Esse requisito molda a arquitetura
desde agora.

## Escopo

**Nesta entrega:**

- Tela de menu (New Game / Scores / Quit)
- Tela de scores lendo os melhores jogadores do armazenamento local. Sem
  gameplay ainda não existem partidas, então ela renderiza um estado vazio
  ("nenhuma partida registrada"); a renderização da lista preenchida é coberta
  por teste com dados de fixture
- Tela de jogo com duas fases placeholder (rodadas e loja), alternáveis
- Navegação completa entre todas as telas, com volta ao menu
- Modo servidor SSH funcional (`yatzee serve`)
- Testes de navegação

**Fora desta entrega:**

- Regras do jogo (dados, pontuação, rodadas), economia e itens da loja
- Pergunta de nome do jogador no `New Game` — a tela de jogo abre direto, porque
  a identidade só é decidida quando houver score para gravar
- Autenticação, contas e scores globais compartilhados
- Web terminal (ttyd), Dockerfile e configuração do Cloudflare Tunnel

## Stack

| Peça | Escolha | Motivo |
|---|---|---|
| Linguagem | Go 1.26 (`apt install golang-go`) | ~10 MB por sessão contra ~60 MB do Python; binário único sem runtime se um dia distribuir local |
| TUI | Bubble Tea | Arquitetura Elm (model/update/view), testável sem terminal |
| Estilo | Lip Gloss | Renderer por sessão, essencial no modo multi-cliente |
| Componentes | Bubbles (`list`, `key`, `help`) | Evita reescrever lista e keybindings |
| Servidor | Wish | Serve um app Bubble Tea como servidor SSH nativo |

O jogador não instala nada em nenhuma das opções de linguagem, porque o runtime
roda no servidor. O critério de decisão foi custo por sessão no servidor.

## Arquitetura

A única regra: **a dependência aponta pra dentro**. A UI conhece as regras; as
regras nunca conhecem a UI.

```
cmd/yatzee ──> internal/app ──> internal/ui ──> internal/game   (regras puras)
                     │
                     └────────────────> internal/store          (persistência)
```

- **`internal/game`** — regras em Go puro. Zero I/O, zero Bubble Tea. Aqui só
  existem os tipos mínimos de costura nesta entrega.
- **`internal/ui`** — telas, componentes e tema.
- **`internal/store`** — interface `ScoreStore` e implementação em arquivo JSON.
- **`internal/app`** — a sessão e o router.
- **`cmd/yatzee`** — `yatzee` joga local, `yatzee serve` sobe o servidor SSH.

### Navegação: pilha de telas

Um model raiz guarda uma pilha de telas. Cada tela é um model independente que
nunca navega por conta própria — ela emite uma mensagem e o router decide.
Telas novas entram sem tocar nas existentes.

```go
// internal/ui/screen.go
type Screen interface {
    Init() tea.Cmd
    Update(tea.Msg) (Screen, tea.Cmd)
    View() string
    Title() string
}

// Navegação por mensagem, nunca por chamada direta entre telas.
type PushMsg struct{ Screen Screen }
type PopMsg  struct{}
type QuitMsg struct{}
```

`Update` devolve `Screen` em vez de `tea.Model` para eliminar type assertion em
cada tela.

### Telas

```
Menu ──> Scores ──> (esc volta)
  └────> Game ──┬── RoundView  (rodadas)   [tab alterna]
                └── ShopView   (lojinha)
                    └─ esc volta pro Menu
```

Round e Shop **não** são telas independentes na pilha: são fases de uma partida.
A tela Game é dona das duas e escolhe qual exibir pela fase atual, mantendo o
estado da partida num lugar só. Nesta entrega ambas são placeholders e o `tab`
alterna entre elas.

### Teclas

Definidas em `internal/ui/keys.go` e exibidas numa barra de help em cada tela.

| Tecla | Efeito | Onde |
|---|---|---|
| `↑`/`↓`, `k`/`j` | Move a seleção | Menu, Scores |
| `enter` | Confirma a seleção | Menu |
| `esc` | Volta uma tela (pop da pilha) | Todas menos o Menu |
| `tab` | Alterna rodadas e loja | Tela de jogo |
| `q` | Sai do jogo | Menu |
| `ctrl+c` | Sai de qualquer lugar | Global |

`esc` no Menu não faz nada: é a base da pilha. Sair exige `q` ou `ctrl+c`, para
que ninguém encerre a partida por reflexo de voltar.

### Sessão

Uma sessão representa um jogador conectado. Local e SSH constroem a mesma coisa.

```go
// internal/app/session.go
type Session struct {
    Config   Config
    Store    store.ScoreStore
    Rand     *rand.Rand         // por sessão, nunca o rand global
    Renderer *lipgloss.Renderer // perfil de cor do cliente
    Width    int
    Height   int
}
```

### Armazenamento

```go
// internal/store/store.go
type Entry struct {
    Player string    `json:"player"`
    Score  int       `json:"score"`
    At     time.Time `json:"at"`
}

type ScoreStore interface {
    Top(ctx context.Context, n int) ([]Entry, error)
    Save(ctx context.Context, e Entry) error
}
```

Implementação inicial em arquivo JSON, com escrita atômica (arquivo temporário
mais rename) e mutex, porque várias sessões podem gravar ao mesmo tempo. O
`context.Context` já está na assinatura para que a troca por Postgres ou HTTP
não altere os chamadores.

## Restrições impostas pelo acesso remoto

Baratas agora, caras depois. Todas valem já nesta entrega.

| Restrição | Motivo | Consequência de ignorar |
|---|---|---|
| RNG por sessão, injetado | `math/rand` global é compartilhado entre goroutines | Jogadores simultâneos interferem na sorte um do outro; nenhuma partida é reproduzível |
| Zero estado mutável de pacote | N sessões no mesmo processo | Um jogador vê ou corrompe a partida do outro |
| Tamanho e cor vindos da sessão, não de `os.Stdout` | Cada cliente SSH tem o próprio PTY | Layout quebrado para todos menos o primeiro cliente |
| Config por env var | Container atrás do Tunnel | Rebuild de imagem para mudar uma porta |

Nota sobre Cloudflare, para a etapa futura: o proxy gratuito só encaminha
HTTP/HTTPS/WebSocket, então SSH puro não passa pelo proxy. Os dois caminhos sem
atrito para o jogador são **SSH direto** (Cloudflare apenas como DNS) e
**browser via web terminal** (passa pelo proxy). A arquitetura acima suporta os
dois sem reescrita, porque o jogo não fala com o terminal diretamente.

## Configuração

| Variável | Default | Uso |
|---|---|---|
| `YATZEE_SSH_HOST` | `0.0.0.0` | Interface do servidor SSH |
| `YATZEE_SSH_PORT` | `2222` | Porta do servidor SSH |
| `YATZEE_DATA_DIR` | `~/.local/share/yatzee` | Onde ficam o arquivo de scores e a chave do host |
| `YATZEE_SSH_HOST_KEY` | `$YATZEE_DATA_DIR/ssh_host_ed25519` | Chave do host, gerada no primeiro boot se ausente |

Caminho absoluto em todos os casos, nunca relativo ao diretório de trabalho: um
serviço iniciado pelo systemd ou por container tem CWD imprevisível.

## Testes

- **`internal/game`**: teste unitário direto, por ser puro. TDD passa a valer
  quando as mecânicas entrarem.
- **Router**: testável sem terminal — injetar `tea.KeyMsg` no model raiz e
  afirmar qual tela está no topo da pilha. É a lógica que existe nesta entrega.
- **`internal/store`**: round-trip de salvar e ler, e ordenação do `Top`.

## Estrutura de arquivos

```
yatzee_terminal/
├── go.mod                      # github.com/afonsodesenv/yatzee-terminal
├── Makefile                    # run, serve, test, build
├── cmd/yatzee/main.go          # CLI: play local | serve
├── internal/
│   ├── app/
│   │   ├── config.go           # Config via env
│   │   ├── session.go          # Session
│   │   ├── root.go             # model raiz: pilha de telas
│   │   └── root_test.go        # testes de navegação
│   ├── ui/
│   │   ├── screen.go           # interface Screen + mensagens de navegação
│   │   ├── theme.go            # estilos Lip Gloss num lugar só
│   │   ├── keys.go             # keybindings + help
│   │   ├── menu/menu.go
│   │   ├── scores/scores.go
│   │   └── game/
│   │       ├── game.go         # container das fases
│   │       ├── round.go        # placeholder
│   │       └── shop.go         # placeholder
│   ├── game/game.go            # tipos de costura (regras vêm depois)
│   └── store/
│       ├── store.go            # ScoreStore + Entry
│       ├── jsonfile.go         # implementação local
│       └── jsonfile_test.go
├── docs/superpowers/specs/
└── README.md
```

## Critérios de aceitação

1. `go run ./cmd/yatzee` no WSL abre o menu.
2. Do menu dá para ir à tela de scores e voltar.
3. `New Game` abre a tela de jogo; `tab` alterna rodadas e loja; `esc` volta ao menu.
4. `Quit` no menu encerra limpando o terminal.
5. `go run ./cmd/yatzee serve` e `ssh -p 2222 localhost` abrem o mesmo jogo pela rede.
6. `go test ./...` passa.
7. `go vet ./...` limpo.

## Decisões postergadas

- Mecânicas: dados, categorias de pontuação, estrutura de rodadas.
- Economia e itens da loja.
- Identidade do jogador: hoje é um nome digitado; no modo SSH a chave pública do
  cliente é a identidade natural, mas isso depende das regras de score global.
- Scores globais e banco de dados.

## Nota operacional

O repositório fica no sistema de arquivos do Windows e o WSL o acessa via
`/mnt/c`, onde a compilação é mais lenta. O cache de build do Go fica no home do
WSL (default), o que mantém os rebuilds rápidos.
