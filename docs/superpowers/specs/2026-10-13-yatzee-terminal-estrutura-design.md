# Yatzee Terminal — Estrutura Base

**Data:** 2026-10-13
**Status:** Aprovado para implementação

## Objetivo

Jogo de dados estilo Yahtzee com progressão e loja inspiradas em Balatro, jogável
no terminal. Esta entrega monta **a estrutura de telas, navegação e identidade**.
As mecânicas do jogo são a próxima conversa e não entram aqui.

O jogo precisa ser jogável remotamente no futuro (servidor próprio atrás de
Cloudflare), sem que o jogador instale nada. Esse requisito molda a arquitetura
desde agora.

## Escopo

**Nesta entrega:**

- Identidade do jogador por chave pública SSH: primeira conexão pergunta o nome,
  as seguintes reconhecem a pessoa
- Nome único e definitivo: se já estiver em uso, a tela acusa o erro e pede outro
- Tela de menu (New Game / Scores / Quit)
- Tela de scores lendo os melhores jogadores do armazenamento local. Sem
  gameplay ainda não existem partidas, então ela renderiza um estado vazio
  ("nenhuma partida registrada"); a renderização da lista preenchida é coberta
  por teste com dados de fixture
- Tela de jogo com duas fases placeholder (rodadas e loja), alternáveis
- Navegação completa entre todas as telas, com volta ao menu
- Modo servidor SSH funcional (`yatzee serve`)
- Testes de navegação, de identidade e de persistência

**Fora desta entrega:**

- Regras do jogo (dados, pontuação, rodadas), economia e itens da loja
- Scores globais compartilhados e banco de dados
- **Trocar o nome depois de escolhido** — nesta versão o nome é definitivo
- Web terminal (ttyd), Dockerfile e configuração do Cloudflare Tunnel
- Rate limit e log de abuso

## Stack

| Peça | Escolha | Motivo |
|---|---|---|
| Linguagem | Go 1.26 (`apt install golang-go`) | ~10 MB por sessão contra ~60 MB do Python; binário único sem runtime se um dia distribuir local |
| TUI | Bubble Tea | Arquitetura Elm (model/update/view), testável sem terminal |
| Estilo | Lip Gloss | Renderer por sessão, essencial no modo multi-cliente |
| Componentes | Bubbles (`list`, `textinput`, `key`, `help`) | Evita reescrever lista, input e keybindings |
| Servidor | Wish | Serve um app Bubble Tea como servidor SSH nativo, com a chave do cliente em mãos |

O jogador não instala nada em nenhuma das opções de linguagem, porque o runtime
roda no servidor. O critério de decisão foi custo por sessão no servidor.

## Arquitetura

A única regra: **a dependência aponta pra dentro**. A UI conhece as regras; as
regras nunca conhecem a UI.

```
cmd/yatzee ──> internal/app ──> internal/ui ──> internal/game    (regras puras)
                     │
                     ├────────────────> internal/score           (persistência)
                     └────────────────> internal/player          (identidade)
```

- **`internal/game`** — regras em Go puro. Zero I/O, zero Bubble Tea. Aqui só
  existem os tipos mínimos de costura nesta entrega.
- **`internal/ui`** — telas, componentes e tema.
- **`internal/score`** — `score.Entry`, `score.Store` e implementação em JSON.
- **`internal/player`** — `player.Identity`, `player.Player`, `player.Store` e
  implementação em JSON. Não importa nada do projeto, então não há ciclo.
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
type PushMsg    struct{ Screen Screen }  // empilha
type PopMsg     struct{}                 // volta uma
type ReplaceMsg struct{ Screen Screen }  // troca o topo, sem deixar volta
type QuitMsg    struct{}
```

`Update` devolve `Screen` em vez de `tea.Model` para eliminar type assertion em
cada tela. `ReplaceMsg` existe para a tela de nome: depois de se apresentar,
ninguém deve poder voltar pra ela com `esc`.

### Telas

```
[primeira vez] Name ──(replace)──> Menu
                                    ├──> Scores ──> (esc volta)
                                    └──> Game ──┬── RoundView  (rodadas)
                                                └── ShopView   (lojinha)
                                                    [tab alterna, esc volta]
```

A tela de nome é a base da pilha apenas quando a identidade é desconhecida; nas
sessões seguintes o Menu já abre direto.

Round e Shop **não** são telas independentes na pilha: são fases de uma partida.
A tela Game é dona das duas e escolhe qual exibir pela fase atual, mantendo o
estado da partida num lugar só. Nesta entrega ambas são placeholders e o `tab`
alterna entre elas.

### Teclas

Definidas em `internal/ui/keys.go` e exibidas numa barra de help em cada tela.

| Tecla | Efeito | Onde |
|---|---|---|
| `↑`/`↓`, `k`/`j` | Move a seleção | Menu, Scores |
| `enter` | Confirma a seleção / o nome | Menu, Name |
| `esc` | Volta uma tela (pop da pilha) | Todas menos Menu e Name |
| `tab` | Alterna rodadas e loja | Tela de jogo |
| `q` | Sai do jogo | Menu |
| `ctrl+c` | Sai de qualquer lugar | Global |

`esc` no Menu não faz nada: é a base da pilha. Sair exige `q` ou `ctrl+c`, para
que ninguém encerre a partida por reflexo de voltar.

## Identidade

O jogador é identificado pelo **fingerprint SHA256 da sua chave pública SSH**,
não pelo IP. O IP identifica conexão, não pessoa: operadoras compartilham um
IPv4 entre milhares de clientes via CGNAT, o lease residencial muda sozinho em
dias, e trocar wifi por 4G muda o IP no meio da sessão. Pior, o IP real **não
sobrevive ao Cloudflare Tunnel** — o `cloudflared` termina a conexão e o jogo vê
`127.0.0.1` para todos. O fingerprint da chave é estável por anos, único por
pessoa mesmo atrás de um NAT compartilhado, e é parte do protocolo SSH, então
atravessa o Tunnel intacto.

```go
// internal/player/player.go
type Kind string

const (
    KindPubKey Kind = "pubkey" // chave SSH: estável entre sessões
    KindLocal  Kind = "local"  // jogo local, usuário do SO
    KindGuest  Kind = "guest"  // SSH sem chave: nome por sessão, não persiste
)

type Identity struct {
    ID   string // "key:SHA256:aBc..." | "local:afons" | "guest:<aleatório>"
    Kind Kind
    Name string // vazio na primeira vez
}

// Persistent informa se vale gravar esta identidade.
func (i Identity) Persistent() bool { return i.Kind != KindGuest }

type Player struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    FirstSeen time.Time `json:"first_seen"`
    LastSeen  time.Time `json:"last_seen"`
}

var ErrNameTaken = errors.New("nome já está em uso")

type Store interface {
    Get(ctx context.Context, id string) (Player, bool, error)
    // Claim grava id→nome se o nome estiver livre, senão devolve ErrNameTaken.
    // Uma chamada só em vez de Existe+Grava porque é menos código, não porque
    // resolve concorrência: colisão simultânea está postergada.
    Claim(ctx context.Context, id, name string) (Player, error)
    Touch(ctx context.Context, id string) error // atualiza LastSeen
}
```

### Nome

Único, definitivo e primeiro que chegar leva. A validação mora no pacote
`player`, não na tela, para valer igual em qualquer entrada — e uma regex
expressa quase tudo, aplicada depois de remover os espaços das pontas:

```go
var nameRE = regexp.MustCompile(`^[a-z0-9_-]+( [a-z0-9_-]+)?$`)
```

| Regra | Valor | Motivo |
|---|---|---|
| Caracteres | Apenas `a-z`, `0-9`, `-`, `_` | Whitelist, não blacklist: o que não está na lista não entra |
| Maiúscula | Convertida na digitação pela tela, e recusada pelo validador | A tela baixa a caixa tecla por tecla, então quem digita "Afonso" lê "afonso" na hora — não há conversão escondida no envio de um nome que é definitivo. O validador continua recusando porque ele é a autoridade, e a tela não é o único caminho de entrada possível |
| Espaço | No máximo um, e nunca nas pontas nem dobrado | O grupo opcional da regex expressa exatamente isso |
| Comprimento | 3 a 16 caracteres | O placar precisa alinhar |
| Conteúdo mínimo | Pelo menos uma letra ou dígito | Rejeita `___` e `---` |
| Unicidade | Comparação direta | Sem maiúscula e sem acento, o nome válido já é a própria chave: nada a normalizar |

O que o whitelist entrega de graça, sem nenhuma regra extra: sequência de escape
ANSI não sobrevive a `[a-z0-9_-]`, então a regra anti-ANSI fica redundante; e
barrar não-ASCII fecha a porta pra impersonação por homóglifo, o truque de
registrar "аfonso" com um "а" cirílico que desenha igual ao latino. Como letras
maiúsculas e acento estão fora, "João", "Ação" e "Afonso" são todos rejeitados —
liberar acento depois é trocar a classe da regex, e aí o comprimento volta a
precisar de `lipgloss.Width` em vez de `len`.

Erro de nome, seja de formato ou de nome em uso, não é falha do sistema: a tela
mantém o texto digitado, mostra a mensagem e espera outro nome. Ela também
exibe a regra abaixo do campo, para a pessoa não descobrir o formato por
tentativa e erro.

**Consequência aceita nesta versão:** nome tomado é nome perdido para sempre, e
quem errou a digitação na primeira conexão fica com o erro. Trocar nome entra
numa versão futura.

**Convidado** (SSH sem chave) escolhe nome sob as mesmas regras e o nome também
precisa estar livre — senão qualquer um se passaria por um jogador registrado no
placar. Mas o nome dele não é reservado, porque a identidade não persiste.

Como cada modo monta a identidade:

| Modo | ID | Nome |
|---|---|---|
| SSH com chave | `key:` + `gossh.FingerprintSHA256(key)` | Perguntado na primeira vez, persistido |
| SSH sem chave | `guest:` + aleatório por sessão | Perguntado a cada sessão, não persistido |
| Local (`yatzee`) | `local:` + usuário do SO | Perguntado na primeira vez, persistido |

No servidor isso exige aceitar qualquer chave como válida — a chave é identidade,
não autorização — e também permitir quem não tem chave:

```go
wish.WithPublicKeyAuth(func(ssh.Context, ssh.PublicKey) bool { return true })
wish.WithKeyboardInteractiveAuth(func(ssh.Context, gossh.KeyboardInteractiveChallenge) bool { return true })
```

Sem o segundo, cliente sem chave sequer conecta.

**IP:** não é gravado nesta entrega. Quando entrar, vai para log de operação
(rate limit, abuso) e nunca como identidade — hasheado ou truncado, porque IP
associado a nome é dado pessoal sob a LGPD num servidor público.

### Sessão

Uma sessão representa um jogador conectado. Local e SSH constroem a mesma coisa.

```go
// internal/app/session.go
type Session struct {
    Config   Config
    Identity player.Identity
    Players  player.Store
    Scores   score.Store
    Rand     *rand.Rand         // por sessão, nunca o rand global
    Renderer *lipgloss.Renderer // perfil de cor do cliente
    Width    int
    Height   int
}
```

### Armazenamento de scores

```go
// internal/score/score.go
type Entry struct {
    PlayerID string    `json:"player_id"`
    Name     string    `json:"name"`
    Points   int       `json:"points"`
    At       time.Time `json:"at"`
}

type Store interface {
    Top(ctx context.Context, n int) ([]Entry, error)
    Save(ctx context.Context, e Entry) error
}
```

Implementações em arquivo JSON, com escrita atômica (arquivo temporário mais
rename) e mutex, porque várias sessões podem gravar ao mesmo tempo. O
`context.Context` já está nas assinaturas para que a troca por Postgres ou HTTP
não altere os chamadores.

## Restrições impostas pelo acesso remoto

Baratas agora, caras depois. Todas valem já nesta entrega.

| Restrição | Motivo | Consequência de ignorar |
|---|---|---|
| RNG por sessão, injetado | `math/rand` global é compartilhado entre goroutines | Jogadores simultâneos interferem na sorte um do outro; nenhuma partida é reproduzível |
| Zero estado mutável de pacote | N sessões no mesmo processo | Um jogador vê ou corrompe a partida do outro |
| Tamanho e cor vindos da sessão, não de `os.Stdout` | Cada cliente SSH tem o próprio PTY | Layout quebrado para todos menos o primeiro cliente |
| Identidade pela chave, não pelo IP | O Tunnel apaga o IP; CGNAT o compartilha | Jogadores se confundem entre si e perdem o nome ao trocar de rede |
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
| `YATZEE_DATA_DIR` | `~/.local/share/yatzee` | Scores, jogadores e chave do host |
| `YATZEE_SSH_HOST_KEY` | `$YATZEE_DATA_DIR/ssh_host_ed25519` | Chave do host, gerada no primeiro boot se ausente |

Caminho absoluto em todos os casos, nunca relativo ao diretório de trabalho: um
serviço iniciado pelo systemd ou por container tem CWD imprevisível.

## Testes

- **`internal/game`**: teste unitário direto, por ser puro. TDD passa a valer
  quando as mecânicas entrarem.
- **Router**: testável sem terminal — injetar `tea.KeyMsg` no model raiz e
  afirmar qual tela está no topo da pilha.
- **Identidade**: tela inicial é Name quando o ID é desconhecido e Menu quando
  conhecido; convidado não é persistido; `Replace` não deixa volta pra tela de
  nome.
- **Nome**: tabela de casos contra a regex — aceita `afo_nso-1` e `afonso dev`,
  rejeita `Afonso` (maiúscula), `ab` (curto), nome de 17, `a  b` (dois espaços),
  `a b c` (dois espaços separados), ` ab ` (pontas), `joão` (acento), `___` (sem
  alfanumérico) e nome com ANSI embutido. `Claim` de nome já gravado devolve
  `ErrNameTaken`; a tela preserva o texto digitado ao mostrar o erro.
- **Tela de nome**: digitar `A` deixa `a` no campo; o validador recusar maiúscula
  continua testado à parte, porque a tela não é a autoridade da regra.
- **Persistência**: round-trip de salvar e ler, ordenação do `Top`, e escrita
  concorrente de duas sessões sem corromper o arquivo.

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
│   │   ├── local.go            # monta sessão local
│   │   ├── server.go           # servidor Wish, sessão por cliente
│   │   ├── root.go             # model raiz: pilha de telas
│   │   └── root_test.go        # navegação e identidade
│   ├── ui/
│   │   ├── screen.go           # interface Screen + mensagens de navegação
│   │   ├── theme.go            # estilos Lip Gloss num lugar só
│   │   ├── keys.go             # keybindings + help
│   │   ├── name/name.go        # pergunta o nome na primeira vez
│   │   ├── menu/menu.go
│   │   ├── scores/scores.go
│   │   └── game/
│   │       ├── game.go         # container das fases
│   │       ├── round.go        # placeholder
│   │       └── shop.go         # placeholder
│   ├── game/game.go            # tipos de costura (regras vêm depois)
│   ├── player/
│   │   ├── player.go           # Identity, Player, Store
│   │   ├── jsonfile.go
│   │   └── jsonfile_test.go
│   └── score/
│       ├── score.go            # Entry, Store
│       ├── jsonfile.go
│       └── jsonfile_test.go
├── docs/superpowers/specs/
└── README.md
```

## Critérios de aceitação

1. `go run ./cmd/yatzee` no WSL pergunta o nome na primeira execução e abre o menu.
2. Rodar de novo vai direto pro menu, com o nome lembrado.
3. Do menu dá para ir à tela de scores e voltar.
4. `New Game` abre a tela de jogo; `tab` alterna rodadas e loja; `esc` volta ao menu.
5. `Quit` no menu encerra limpando o terminal.
6. `go run ./cmd/yatzee serve` e `ssh -p 2222 localhost` abrem o mesmo jogo pela
   rede, perguntando o nome na primeira conexão e reconhecendo na segunda.
7. Tentar um nome já usado mostra o erro na própria tela e aceita outro nome em
   seguida, sem derrubar a sessão.
8. `go test -race ./...` passa.
9. `go vet ./...` limpo.

## Decisões postergadas

- Mecânicas: dados, categorias de pontuação, estrutura de rodadas.
- Economia e itens da loja.
- Scores globais e banco de dados.
- **Dois jogadores reivindicando o mesmo nome no mesmo instante.** A garantia de
  verdade é índice único no banco, então isso casa com a troca do JSON por banco.
- Trocar o nome depois de escolhido, e o que fazer com nome de jogador inativo.
- Acento no nome (exige voltar a medir largura em colunas).
- Se partida de convidado pode pontuar no placar, já que a identidade dele não
  persiste.

## Nota operacional

O repositório fica no sistema de arquivos do Windows e o WSL o acessa via
`/mnt/c`, onde a compilação é mais lenta. O cache de build do Go fica no home do
WSL (default), o que mantém os rebuilds rápidos.
