# Yatzee Terminal — Estrutura Base: Plano de Implementação

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Entregar a estrutura navegável do jogo — menu, scores, jogo com duas fases placeholder e identidade por chave SSH — jogável localmente e por SSH.

**Architecture:** Dependência aponta pra dentro: `cmd` → `app` → `ui` → (`player`, `score`). Cada tela é um model Bubble Tea independente que não conhece nenhuma outra tela: emite mensagem de intenção e o model raiz, único dono do grafo de navegação, traduz em empilhar/voltar/trocar. Cada tela recebe só a dependência que usa, nunca um objeto de sessão inteiro.

**Tech Stack:** Go 1.27.2, `charm.land/bubbletea/v2` v2.1.0, `charm.land/lipgloss/v2` v2.0.6, `charm.land/bubbles/v2` v2.2.1 (só `textinput`), `charm.land/wish/v2` v2.0.5, `charm.land/ssh` v0.4.3, `golang.org/x/crypto/ssh`.

**Spec:** `docs/superpowers/specs/2026-10-13-yatzee-terminal-estrutura-design.md`

## Global Constraints

Valem para toda task. Violar qualquer uma destas é motivo de rejeição na review.

- **Go toolchain:** `go1.27.2`, instalado em `~/.local/go/bin` **dentro do WSL**. O repositório fica no disco do Windows, então todo comando Go roda através do WSL, nunca direto:

  ```bash
  wsl -e bash -lc 'export PATH=$HOME/.local/go/bin:$PATH; cd /mnt/c/GIT/yatzee_terminal && go test ./...'
  ```

  Os comandos `go ...` escritos nas tasks são a parte interna: embrulhe cada um
  nesse formato. Não existe Go no lado Windows — chamar `go` direto falha com
  "command not found".
- **Module path:** `github.com/afonsodesenv/yatzee-terminal`.
- **Imports v2 vêm de `charm.land`**, nunca de `github.com/charmbracelet`. O caminho antigo existe e compila na v1, com API diferente — errar aqui gera erro de tipo confuso.
- **Versão pinada na primeira importação, sempre explícita.** O `go mod tidy` remove do `go.mod` toda dependência que nenhum código importa ainda — foi o que aconteceu com `bubbles`, `wish`, `ssh` e `crypto` depois da Task 5. Então a task que **primeiro** importa cada uma precisa pedir a versão verificada por extenso, nunca um `go get` sem versão (que resolveria para a mais recente e poderia trazer uma API diferente da verificada):

  ```bash
  go get charm.land/bubbles/v2/textinput@v2.2.1   # Task 7
  go get charm.land/wish/v2@v2.0.5 charm.land/wish/v2/bubbletea@v2.0.5 \
         charm.land/ssh@v0.4.3 golang.org/x/crypto@latest   # Task 12
  ```
- **Zero estado mutável de pacote.** Nenhuma variável global que mude em runtime. Várias sessões vivem no mesmo processo.
- **Nenhum uso de `math/rand` global.** Só `*rand.Rand` por sessão (`math/rand/v2`).
- **Nenhuma tela importa outra tela.** Nenhum pacote `ui/*` importa `internal/app`.
- **Nada lê `os.Stdout` para descobrir tamanho ou cor.** Tamanho chega por `tea.WindowSizeMsg`.
- **Texto de interface em português**, minúsculo no corpo das telas, sem emoji.
- **Arquivos de dados com permissão `0600`, diretórios `0700`.**
- **Regra de nome, verbatim do spec:** regex `^[a-z0-9_-]+( [a-z0-9_-]+)?$`, 3 a 16 caracteres após trim, pelo menos um alfanumérico. Unicidade por comparação direta.
- **Defaults de config, verbatim:** `YATZEE_SSH_HOST=0.0.0.0`, `YATZEE_SSH_PORT=2222`, `YATZEE_DATA_DIR=~/.local/share/yatzee`, `YATZEE_SSH_HOST_KEY=$YATZEE_DATA_DIR/ssh_host_ed25519`.
- **Commits:** um por task, mensagem em português no imperativo, prefixo `feat:`/`test:`/`chore:`. Terminar com `Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>`.

### Armadilhas já confirmadas rodando o stack

Não "descubra" estas de novo: elas custaram um probe inteiro.

| Armadilha | O que fazer |
|---|---|
| `View()` da raiz devolve `tea.View` | `v := tea.NewView(conteudo); v.AltScreen = true; return v` |
| Telas devolvem `string` | Só a raiz lida com `tea.View` |
| Mensagem de tecla | `tea.KeyPressMsg`, com `msg.String()` para comparar |
| `lipgloss.NewRenderer` | Não existe na v2. Use `lipgloss.NewStyle()` direto |
| `bubbletea.MakeOptions` | **Nunca** chamar no handler: o middleware já aplica. Devolva `nil` como opções |
| `wish.WithHostKeyPath` | Faz `chmod` no diretório pai. Garanta `os.MkdirAll(cfg.DataDir, 0o700)` antes de subir o servidor |
| Cliente com PTY `0x0` | Passar por `ui.Size()` antes de desenhar, senão a tela sai branca |
| Testar TUI por pipe/PTY falso | Não funciona: o programa adia o frame esperando resposta do terminal. Teste injetando `tea.Msg` em `Update` e lendo `View()` |
| Constantes de tecla | Verificadas rodando: `tea.KeyEnter`→`"enter"`, `tea.KeyEscape`→`"esc"`, `tea.KeyTab`→`"tab"`, `tea.KeyUp`→`"up"`, `tea.KeyDown`→`"down"`. Letra se monta com `tea.KeyPressMsg{Code: 'j', Text: "j"}` e `String()` devolve `"j"` |

---

### Task 1: Fundação do módulo e configuração

**Files:**
- Create: `go.mod` (via comando)
- Create: `Makefile`
- Create: `internal/app/config.go`
- Test: `internal/app/config_test.go`

**Interfaces:**
- Consumes: nada (primeira task)
- Produces: `app.Config{SSHHost, SSHPort, DataDir, SSHHostKey string}`, `app.LoadConfig() (Config, error)`, `app.PlayersPath() string`, `app.ScoresPath() string`

- [ ] **Step 1: Criar o módulo e travar as dependências**

```bash
export PATH=$HOME/.local/go/bin:$PATH
cd /mnt/c/GIT/yatzee_terminal
go mod init github.com/afonsodesenv/yatzee-terminal
go get charm.land/bubbletea/v2@v2.1.0
go get charm.land/lipgloss/v2@v2.0.6
go get charm.land/bubbles/v2@v2.2.1
go get charm.land/wish/v2@v2.0.5
go get charm.land/ssh@v0.4.3
go get golang.org/x/crypto/ssh
```

- [ ] **Step 2: Escrever o teste de config**

Crie `internal/app/config_test.go`:

```go
package app

import (
	"path/filepath"
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {
	t.Setenv("YATZEE_DATA_DIR", "")
	t.Setenv("YATZEE_SSH_HOST", "")
	t.Setenv("YATZEE_SSH_PORT", "")
	t.Setenv("YATZEE_SSH_HOST_KEY", "")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.SSHHost != "0.0.0.0" {
		t.Errorf("SSHHost = %q, quero 0.0.0.0", cfg.SSHHost)
	}
	if cfg.SSHPort != "2222" {
		t.Errorf("SSHPort = %q, quero 2222", cfg.SSHPort)
	}
	if !filepath.IsAbs(cfg.DataDir) {
		t.Errorf("DataDir = %q, quero caminho absoluto", cfg.DataDir)
	}
	if want := filepath.Join(cfg.DataDir, "ssh_host_ed25519"); cfg.SSHHostKey != want {
		t.Errorf("SSHHostKey = %q, quero %q", cfg.SSHHostKey, want)
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("YATZEE_DATA_DIR", dir)
	t.Setenv("YATZEE_SSH_HOST", "127.0.0.1")
	t.Setenv("YATZEE_SSH_PORT", "3333")
	t.Setenv("YATZEE_SSH_HOST_KEY", "")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.DataDir != dir {
		t.Errorf("DataDir = %q, quero %q", cfg.DataDir, dir)
	}
	if cfg.SSHHost != "127.0.0.1" || cfg.SSHPort != "3333" {
		t.Errorf("host/porta = %q/%q", cfg.SSHHost, cfg.SSHPort)
	}
	if cfg.PlayersPath() != filepath.Join(dir, "players.json") {
		t.Errorf("PlayersPath = %q", cfg.PlayersPath())
	}
	if cfg.ScoresPath() != filepath.Join(dir, "scores.json") {
		t.Errorf("ScoresPath = %q", cfg.ScoresPath())
	}
}

// Um serviço sob systemd ou container tem CWD imprevisível: caminho relativo
// na env tem que virar absoluto.
func TestLoadConfigTornaDataDirAbsoluto(t *testing.T) {
	t.Setenv("YATZEE_DATA_DIR", "dados-relativos")
	t.Setenv("YATZEE_SSH_HOST_KEY", "")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if !filepath.IsAbs(cfg.DataDir) {
		t.Errorf("DataDir = %q, quero absoluto", cfg.DataDir)
	}
}

// A mesma garantia vale para o override da chave do host: é justamente o
// cenário que a regra existe para evitar (unit do systemd com caminho relativo).
func TestLoadConfigTornaHostKeyAbsoluta(t *testing.T) {
	t.Setenv("YATZEE_DATA_DIR", t.TempDir())
	t.Setenv("YATZEE_SSH_HOST_KEY", "chave-relativa")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if !filepath.IsAbs(cfg.SSHHostKey) {
		t.Errorf("SSHHostKey = %q, quero absoluto", cfg.SSHHostKey)
	}
	if filepath.Base(cfg.SSHHostKey) != "chave-relativa" {
		t.Errorf("SSHHostKey = %q, quero preservar o nome do arquivo", cfg.SSHHostKey)
	}
}
```

- [ ] **Step 3: Rodar o teste e confirmar que falha**

Run: `export PATH=$HOME/.local/go/bin:$PATH && go test ./internal/app/ -run TestLoadConfig -v`
Expected: FAIL com `undefined: LoadConfig`

- [ ] **Step 4: Implementar a config**

Crie `internal/app/config.go`:

```go
// Package app monta a sessão de jogo e roteia as telas.
package app

import (
	"fmt"
	"os"
	"path/filepath"
)

// Config reúne tudo que vem do ambiente. Nada aqui é lido de novo depois do
// boot: uma sessão recebe a config já resolvida.
type Config struct {
	SSHHost    string
	SSHPort    string
	DataDir    string
	SSHHostKey string
}

// LoadConfig lê a configuração do ambiente, aplicando os defaults do spec.
func LoadConfig() (Config, error) {
	dataDir := os.Getenv("YATZEE_DATA_DIR")
	if dataDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return Config{}, fmt.Errorf("descobrir home do usuário: %w", err)
		}
		dataDir = filepath.Join(home, ".local", "share", "yatzee")
	}
	// Caminho relativo é inútil num serviço: o CWD de um systemd ou container
	// não é o que o operador imagina.
	dataDir, err := filepath.Abs(dataDir)
	if err != nil {
		return Config{}, fmt.Errorf("resolver caminho absoluto de %q: %w", dataDir, err)
	}

	hostKey := os.Getenv("YATZEE_SSH_HOST_KEY")
	if hostKey == "" {
		hostKey = filepath.Join(dataDir, "ssh_host_ed25519")
	} else {
		// Override do operador também precisa ser absoluto, pelo mesmo motivo
		// do DataDir: um serviço não escolhe o próprio CWD.
		abs, absErr := filepath.Abs(hostKey)
		if absErr != nil {
			return Config{}, fmt.Errorf("resolver caminho absoluto de %q: %w", hostKey, absErr)
		}
		hostKey = abs
	}

	return Config{
		SSHHost:    envOr("YATZEE_SSH_HOST", "0.0.0.0"),
		SSHPort:    envOr("YATZEE_SSH_PORT", "2222"),
		DataDir:    dataDir,
		SSHHostKey: hostKey,
	}, nil
}

// PlayersPath é o arquivo de identidades.
func (c Config) PlayersPath() string { return filepath.Join(c.DataDir, "players.json") }

// ScoresPath é o arquivo de placar.
func (c Config) ScoresPath() string { return filepath.Join(c.DataDir, "scores.json") }

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
```

- [ ] **Step 5: Rodar o teste e confirmar que passa**

Run: `export PATH=$HOME/.local/go/bin:$PATH && go test ./internal/app/ -v -count=1`
Expected: PASS nos três testes

- [ ] **Step 6: Criar o Makefile**

```makefile
GO := $(HOME)/.local/go/bin/go

.PHONY: run serve test test-race vet build

run:
	$(GO) run ./cmd/yatzee

serve:
	$(GO) run ./cmd/yatzee serve

test:
	$(GO) test ./...

test-race:
	@command -v gcc >/dev/null || { echo "precisa de compilador C: sudo apt install build-essential"; exit 1; }
	CGO_ENABLED=1 $(GO) test -race ./...

vet:
	$(GO) vet ./...

build:
	$(GO) build -o bin/yatzee ./cmd/yatzee
```

- [ ] **Step 7: Commit**

```bash
git add go.mod go.sum Makefile internal/app/
git commit -m "feat: fundacao do modulo e configuracao por ambiente

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

### Task 2: Identidade e regra de nome

**Files:**
- Create: `internal/player/player.go`
- Test: `internal/player/player_test.go`

**Interfaces:**
- Consumes: nada
- Produces: `player.Kind` (`KindPubKey`, `KindLocal`, `KindGuest`), `player.Identity{ID string, Kind Kind, Name string}`, `(Identity).Persistent() bool`, `player.Player{ID, Name string, FirstSeen, LastSeen time.Time}`, `player.Store` (interface), `player.CleanName(string) (string, error)`, `player.NameRule() string`, `player.ErrNameInvalid`, `player.ErrNameTaken`, `player.NameMinLen`, `player.NameMaxLen`

- [ ] **Step 1: Escrever o teste da regra de nome**

Crie `internal/player/player_test.go`:

```go
package player

import (
	"errors"
	"strings"
	"testing"
)

func TestCleanNameAceita(t *testing.T) {
	casos := []struct {
		entrada string
		quero   string
	}{
		{"afonso", "afonso"},
		{"afo_nso-1", "afo_nso-1"},
		{"afonso dev", "afonso dev"},
		{"abc", "abc"},
		{"1234567890123456", "1234567890123456"},
		{"  afonso  ", "afonso"}, // trim antes de validar
	}
	for _, c := range casos {
		got, err := CleanName(c.entrada)
		if err != nil {
			t.Errorf("CleanName(%q) = erro %v, quero aceitar", c.entrada, err)
			continue
		}
		if got != c.quero {
			t.Errorf("CleanName(%q) = %q, quero %q", c.entrada, got, c.quero)
		}
	}
}

func TestCleanNameRejeita(t *testing.T) {
	casos := []struct {
		entrada string
		motivo  string
	}{
		{"Afonso", "maiúscula"},
		{"AFONSO", "tudo maiúsculo"},
		{"ab", "curto demais"},
		{"12345678901234567", "longo demais"},
		{"a  b", "dois espaços seguidos"},
		{"a b c", "dois espaços separados"},
		{"joão", "acento"},
		{"afonso!", "pontuação"},
		{"___", "sem alfanumérico"},
		{"---", "sem alfanumérico"},
		{"", "vazio"},
		{"   ", "só espaço"},
		{" ab ", "curto depois do trim"},
		{"af\x1b[31monso", "ANSI embutido"},
		{"afonso\n", "controle"},
	}
	for _, c := range casos {
		if _, err := CleanName(c.entrada); !errors.Is(err, ErrNameInvalid) {
			t.Errorf("CleanName(%q) = %v, quero ErrNameInvalid (%s)", c.entrada, err, c.motivo)
		}
	}
}

func TestNameRuleMencionaOsLimites(t *testing.T) {
	rule := NameRule()
	for _, termo := range []string{"3", "16"} {
		if !strings.Contains(rule, termo) {
			t.Errorf("NameRule() = %q, quero mencionar %q", rule, termo)
		}
	}
}

func TestIdentityPersistent(t *testing.T) {
	casos := []struct {
		kind  Kind
		quero bool
	}{
		{KindPubKey, true},
		{KindLocal, true},
		{KindGuest, false},
	}
	for _, c := range casos {
		if got := (Identity{Kind: c.kind}).Persistent(); got != c.quero {
			t.Errorf("Identity{%s}.Persistent() = %v, quero %v", c.kind, got, c.quero)
		}
	}
}
```

- [ ] **Step 2: Rodar o teste e confirmar que falha**

Run: `export PATH=$HOME/.local/go/bin:$PATH && go test ./internal/player/ -v -count=1`
Expected: FAIL com `undefined: CleanName`

- [ ] **Step 3: Implementar identidade e nome**

Crie `internal/player/player.go`:

```go
// Package player guarda a identidade do jogador e a regra do nome. Não
// conhece UI nem transporte: a mesma regra vale no jogo local e no servidor.
package player

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"
)

// Limites do nome. O teto existe porque o placar precisa alinhar.
const (
	NameMinLen = 3
	NameMaxLen = 16
)

var (
	// ErrNameInvalid é formato recusado.
	ErrNameInvalid = errors.New("nome inválido")
	// ErrNameTaken é nome já reivindicado por outra identidade.
	ErrNameTaken = errors.New("nome já está em uso")
)

// Um espaço interno no máximo, nunca nas pontas nem dobrado: o grupo opcional
// expressa exatamente isso.
var nameRE = regexp.MustCompile(`^[a-z0-9_-]+( [a-z0-9_-]+)?$`)

// CleanName devolve o nome canônico ou ErrNameInvalid. É o único caminho para
// um nome entrar no sistema, então nenhum chamador precisa lembrar de validar.
func CleanName(s string) (string, error) {
	// Caractere de controle é recusado ANTES do trim. Ao contrário: o
	// TrimSpace removeria o "\n" de "afonso\n" e um nome com controle
	// embutido passaria a válido.
	if strings.ContainsFunc(s, unicode.IsControl) {
		return "", fmt.Errorf("%w: caractere de controle", ErrNameInvalid)
	}
	s = strings.TrimSpace(s)
	if len(s) < NameMinLen || len(s) > NameMaxLen {
		return "", fmt.Errorf("%w: precisa de %d a %d caracteres", ErrNameInvalid, NameMinLen, NameMaxLen)
	}
	if !nameRE.MatchString(s) {
		return "", fmt.Errorf("%w: %s", ErrNameInvalid, NameRule())
	}
	if !strings.ContainsFunc(s, isAlnum) {
		return "", fmt.Errorf("%w: precisa de pelo menos uma letra ou número", ErrNameInvalid)
	}
	return s, nil
}

// NameRule descreve a regra para mostrar na tela, para ninguém descobrir o
// formato por tentativa e erro.
func NameRule() string {
	return fmt.Sprintf("%d a %d: minúsculas, números, - e _, no máximo um espaço", NameMinLen, NameMaxLen)
}

func isAlnum(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
}

// Kind diz de onde a identidade veio, o que determina se ela persiste.
type Kind string

const (
	KindPubKey Kind = "pubkey" // chave SSH: estável entre sessões
	KindLocal  Kind = "local"  // jogo local, usuário do SO
	KindGuest  Kind = "guest"  // SSH sem chave: nome por sessão
)

// Identity é quem está jogando nesta sessão.
type Identity struct {
	ID   string
	Kind Kind
	Name string // vazio quando a pessoa ainda não se apresentou
}

// Persistent informa se vale gravar esta identidade.
func (i Identity) Persistent() bool { return i.Kind != KindGuest }

// Player é a identidade gravada.
type Player struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
}

// Store persiste identidades. Interface pequena de propósito: trocar JSON por
// banco depois é implementar três métodos.
type Store interface {
	Get(ctx context.Context, id string) (Player, bool, error)
	// Claim grava id→nome se o nome estiver livre, senão devolve ErrNameTaken.
	// Uma chamada só em vez de Existe+Grava porque é menos código; colisão
	// simultânea entre duas sessões está postergada no spec.
	Claim(ctx context.Context, id, name string) (Player, error)
	// NameTaken diz se o nome já pertence a outra identidade. Serve ao
	// convidado, que precisa de nome livre mas não reserva nada.
	NameTaken(ctx context.Context, name string) (bool, error)
	Touch(ctx context.Context, id string) error
}
```

- [ ] **Step 4: Rodar o teste e confirmar que passa**

Run: `export PATH=$HOME/.local/go/bin:$PATH && go test ./internal/player/ -v -count=1`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/player/
git commit -m "feat: identidade do jogador e regra rigida de nome

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

### Task 3: Persistência de identidades em JSON

**Files:**
- Create: `internal/player/jsonfile.go`
- Test: `internal/player/jsonfile_test.go`

**Interfaces:**
- Consumes: `player.Player`, `player.Store`, `player.ErrNameTaken` (Task 2)
- Produces: `player.NewJSONStore(path string) *JSONStore` implementando `player.Store`

- [ ] **Step 1: Escrever o teste da persistência**

Crie `internal/player/jsonfile_test.go`:

```go
package player

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func novoStore(t *testing.T) *JSONStore {
	t.Helper()
	return NewJSONStore(filepath.Join(t.TempDir(), "players.json"))
}

func TestClaimEGet(t *testing.T) {
	ctx := context.Background()
	s := novoStore(t)

	p, err := s.Claim(ctx, "key:abc", "afonso")
	if err != nil {
		t.Fatalf("Claim: %v", err)
	}
	if p.Name != "afonso" || p.ID != "key:abc" {
		t.Fatalf("Claim devolveu %+v", p)
	}
	if p.FirstSeen.IsZero() || p.LastSeen.IsZero() {
		t.Errorf("Claim não preencheu os timestamps: %+v", p)
	}

	got, ok, err := s.Get(ctx, "key:abc")
	if err != nil || !ok {
		t.Fatalf("Get = %v, %v, %v", got, ok, err)
	}
	if got.Name != "afonso" {
		t.Errorf("Get().Name = %q, quero afonso", got.Name)
	}
}

func TestGetIdentidadeDesconhecida(t *testing.T) {
	_, ok, err := novoStore(t).Get(context.Background(), "key:nao-existe")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if ok {
		t.Error("Get de id desconhecido devolveu ok=true")
	}
}

func TestClaimNomeEmUsoPorOutraIdentidade(t *testing.T) {
	ctx := context.Background()
	s := novoStore(t)

	if _, err := s.Claim(ctx, "key:abc", "afonso"); err != nil {
		t.Fatalf("primeiro Claim: %v", err)
	}
	if _, err := s.Claim(ctx, "key:xyz", "afonso"); !errors.Is(err, ErrNameTaken) {
		t.Errorf("segundo Claim = %v, quero ErrNameTaken", err)
	}
}

func TestClaimDoMesmoIdComMesmoNomeNaoFalha(t *testing.T) {
	ctx := context.Background()
	s := novoStore(t)

	if _, err := s.Claim(ctx, "key:abc", "afonso"); err != nil {
		t.Fatalf("primeiro Claim: %v", err)
	}
	if _, err := s.Claim(ctx, "key:abc", "afonso"); err != nil {
		t.Errorf("reivindicar o próprio nome falhou: %v", err)
	}
}

func TestNameTaken(t *testing.T) {
	ctx := context.Background()
	s := novoStore(t)
	if _, err := s.Claim(ctx, "key:abc", "afonso"); err != nil {
		t.Fatalf("Claim: %v", err)
	}

	taken, err := s.NameTaken(ctx, "afonso")
	if err != nil || !taken {
		t.Errorf("NameTaken(afonso) = %v, %v; quero true, nil", taken, err)
	}
	taken, err = s.NameTaken(ctx, "outro")
	if err != nil || taken {
		t.Errorf("NameTaken(outro) = %v, %v; quero false, nil", taken, err)
	}
}

func TestTouchAtualizaLastSeen(t *testing.T) {
	ctx := context.Background()
	s := novoStore(t)
	p, err := s.Claim(ctx, "key:abc", "afonso")
	if err != nil {
		t.Fatalf("Claim: %v", err)
	}

	if err := s.Touch(ctx, "key:abc"); err != nil {
		t.Fatalf("Touch: %v", err)
	}
	got, _, err := s.Get(ctx, "key:abc")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.LastSeen.Before(p.LastSeen) {
		t.Errorf("LastSeen andou pra trás: %v < %v", got.LastSeen, p.LastSeen)
	}
	if !got.FirstSeen.Equal(p.FirstSeen) {
		t.Errorf("Touch mexeu no FirstSeen: %v != %v", got.FirstSeen, p.FirstSeen)
	}
}

func TestArquivoNasceCom0600(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	path := filepath.Join(dir, "players.json")
	s := NewJSONStore(path)
	if _, err := s.Claim(ctx, "key:abc", "afonso"); err != nil {
		t.Fatalf("Claim: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("permissão = %o, quero 600", perm)
	}
}

func TestEscritaConcorrenteNaoCorrompe(t *testing.T) {
	ctx := context.Background()
	s := novoStore(t)

	const n = 20
	done := make(chan error, n)
	for i := range n {
		go func(i int) {
			_, err := s.Claim(ctx, fmt.Sprintf("key:%d", i), fmt.Sprintf("jogador-%d", i))
			done <- err
		}(i)
	}
	for range n {
		if err := <-done; err != nil {
			t.Fatalf("Claim concorrente: %v", err)
		}
	}
	for i := range n {
		if _, ok, err := s.Get(ctx, fmt.Sprintf("key:%d", i)); err != nil || !ok {
			t.Errorf("identidade %d sumiu: ok=%v err=%v", i, ok, err)
		}
	}
}
```

- [ ] **Step 2: Rodar o teste e confirmar que falha**

Run: `export PATH=$HOME/.local/go/bin:$PATH && go test ./internal/player/ -run TestClaim -v`
Expected: FAIL com `undefined: NewJSONStore`

- [ ] **Step 3: Implementar o store em JSON**

Crie `internal/player/jsonfile.go`:

```go
package player

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// JSONStore guarda identidades num arquivo JSON. Serve ao jogo local e ao
// servidor com poucas dezenas de jogadores; trocar por banco é implementar a
// mesma interface.
type JSONStore struct {
	path string
	mu   sync.Mutex
}

// NewJSONStore devolve um store apontando para path, que não precisa existir.
func NewJSONStore(path string) *JSONStore { return &JSONStore{path: path} }

type playersFile struct {
	Players map[string]Player `json:"players"`
}

// Get busca uma identidade pelo ID.
func (s *JSONStore) Get(_ context.Context, id string) (Player, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.load()
	if err != nil {
		return Player{}, false, err
	}
	p, ok := data.Players[id]
	return p, ok, nil
}

// Claim grava id→nome se o nome estiver livre. Reivindicar o próprio nome de
// novo é idempotente, porque reconectar não deve falhar.
func (s *JSONStore) Claim(_ context.Context, id, name string) (Player, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.load()
	if err != nil {
		return Player{}, err
	}
	for otherID, p := range data.Players {
		if p.Name == name && otherID != id {
			return Player{}, fmt.Errorf("%w: %q", ErrNameTaken, name)
		}
	}

	now := time.Now().UTC()
	p, existed := data.Players[id]
	if !existed {
		p = Player{ID: id, FirstSeen: now}
	}
	p.Name = name
	p.LastSeen = now
	data.Players[id] = p

	if err := s.save(data); err != nil {
		return Player{}, err
	}
	return p, nil
}

// NameTaken diz se o nome já pertence a alguma identidade.
func (s *JSONStore) NameTaken(_ context.Context, name string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.load()
	if err != nil {
		return false, err
	}
	for _, p := range data.Players {
		if p.Name == name {
			return true, nil
		}
	}
	return false, nil
}

// Touch marca a identidade como vista agora. Silencioso para ID desconhecido:
// um convidado não tem registro e isso não é erro.
func (s *JSONStore) Touch(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.load()
	if err != nil {
		return err
	}
	p, ok := data.Players[id]
	if !ok {
		return nil
	}
	p.LastSeen = time.Now().UTC()
	data.Players[id] = p
	return s.save(data)
}

// load lê o arquivo. Ausente conta como vazio: primeiro boot não é erro.
func (s *JSONStore) load() (playersFile, error) {
	data := playersFile{Players: map[string]Player{}}
	raw, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return data, nil
	}
	if err != nil {
		return data, fmt.Errorf("ler %s: %w", s.path, err)
	}
	if len(raw) == 0 {
		return data, nil
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return data, fmt.Errorf("decodificar %s: %w", s.path, err)
	}
	if data.Players == nil {
		data.Players = map[string]Player{}
	}
	return data, nil
}

// save grava em arquivo temporário e renomeia: um crash no meio da escrita
// deixa o arquivo anterior intacto em vez de um JSON truncado.
func (s *JSONStore) save(data playersFile) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("criar diretório de %s: %w", s.path, err)
	}
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("codificar identidades: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".players-*.json")
	if err != nil {
		return fmt.Errorf("criar temporário: %w", err)
	}
	defer os.Remove(tmp.Name())

	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return fmt.Errorf("permissão do temporário: %w", err)
	}
	if _, err := tmp.Write(raw); err != nil {
		tmp.Close()
		return fmt.Errorf("escrever temporário: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("fechar temporário: %w", err)
	}
	if err := os.Rename(tmp.Name(), s.path); err != nil {
		return fmt.Errorf("renomear para %s: %w", s.path, err)
	}
	return nil
}
```

- [ ] **Step 4: Rodar os testes com detector de corrida**

Run: `export PATH=$HOME/.local/go/bin:$PATH && go test ./internal/player/ -v -count=1`
Expected: PASS em tudo, sem aviso de corrida

- [ ] **Step 5: Commit**

```bash
git add internal/player/
git commit -m "feat: persistencia de identidades em json com escrita atomica

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

### Task 4: Placar

**Files:**
- Create: `internal/score/score.go`
- Create: `internal/score/jsonfile.go`
- Test: `internal/score/jsonfile_test.go`

**Interfaces:**
- Consumes: nada
- Produces: `score.Entry{PlayerID, Name string, Points int, At time.Time}`, `score.Store` (interface com `Top(ctx, n)` e `Save(ctx, Entry)`), `score.NewJSONStore(path string) *JSONStore`

- [ ] **Step 1: Escrever o teste do placar**

Crie `internal/score/jsonfile_test.go`:

```go
package score

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func novoStore(t *testing.T) *JSONStore {
	t.Helper()
	return NewJSONStore(filepath.Join(t.TempDir(), "scores.json"))
}

func TestTopVazio(t *testing.T) {
	got, err := novoStore(t).Top(context.Background(), 10)
	if err != nil {
		t.Fatalf("Top: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("Top em store vazio = %v, quero vazio", got)
	}
}

func TestSaveETopOrdenaPorPontosDesc(t *testing.T) {
	ctx := context.Background()
	s := novoStore(t)
	base := time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC)

	entradas := []Entry{
		{PlayerID: "key:a", Name: "ana", Points: 120, At: base},
		{PlayerID: "key:b", Name: "bob", Points: 300, At: base.Add(time.Minute)},
		{PlayerID: "key:c", Name: "caio", Points: 200, At: base.Add(2 * time.Minute)},
	}
	for _, e := range entradas {
		if err := s.Save(ctx, e); err != nil {
			t.Fatalf("Save(%s): %v", e.Name, err)
		}
	}

	got, err := s.Top(ctx, 10)
	if err != nil {
		t.Fatalf("Top: %v", err)
	}
	quero := []string{"bob", "caio", "ana"}
	if len(got) != len(quero) {
		t.Fatalf("Top devolveu %d entradas, quero %d", len(got), len(quero))
	}
	for i, nome := range quero {
		if got[i].Name != nome {
			t.Errorf("Top[%d].Name = %q, quero %q", i, got[i].Name, nome)
		}
	}
}

func TestTopEmpateDesempatatePeloMaisAntigo(t *testing.T) {
	ctx := context.Background()
	s := novoStore(t)
	base := time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC)

	if err := s.Save(ctx, Entry{PlayerID: "key:b", Name: "bob", Points: 100, At: base.Add(time.Hour)}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := s.Save(ctx, Entry{PlayerID: "key:a", Name: "ana", Points: 100, At: base}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := s.Top(ctx, 10)
	if err != nil {
		t.Fatalf("Top: %v", err)
	}
	if got[0].Name != "ana" {
		t.Errorf("Top[0] = %q, quero ana (chegou primeiro no empate)", got[0].Name)
	}
}

func TestTopRespeitaOLimite(t *testing.T) {
	ctx := context.Background()
	s := novoStore(t)
	for i := range 5 {
		e := Entry{PlayerID: "key:x", Name: "x", Points: i, At: time.Now().UTC()}
		if err := s.Save(ctx, e); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}

	got, err := s.Top(ctx, 3)
	if err != nil {
		t.Fatalf("Top: %v", err)
	}
	if len(got) != 3 {
		t.Errorf("Top(3) devolveu %d entradas", len(got))
	}
}

func TestSaveConcorrenteNaoPerdeEntrada(t *testing.T) {
	ctx := context.Background()
	s := novoStore(t)

	const n = 20
	done := make(chan error, n)
	for i := range n {
		go func(i int) {
			done <- s.Save(ctx, Entry{PlayerID: "key:x", Name: "x", Points: i, At: time.Now().UTC()})
		}(i)
	}
	for range n {
		if err := <-done; err != nil {
			t.Fatalf("Save concorrente: %v", err)
		}
	}
	got, err := s.Top(ctx, 100)
	if err != nil {
		t.Fatalf("Top: %v", err)
	}
	if len(got) != n {
		t.Errorf("guardou %d entradas, quero %d", len(got), n)
	}
}
```

- [ ] **Step 2: Rodar o teste e confirmar que falha**

Run: `export PATH=$HOME/.local/go/bin:$PATH && go test ./internal/score/ -v -count=1`
Expected: FAIL com `undefined: NewJSONStore`

- [ ] **Step 3: Implementar o tipo e a interface**

Crie `internal/score/score.go`:

```go
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
```

- [ ] **Step 4: Implementar o store em JSON**

Crie `internal/score/jsonfile.go`:

```go
package score

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// JSONStore guarda o placar num arquivo JSON.
type JSONStore struct {
	path string
	mu   sync.Mutex
}

// NewJSONStore devolve um store apontando para path, que não precisa existir.
func NewJSONStore(path string) *JSONStore { return &JSONStore{path: path} }

type scoresFile struct {
	Entries []Entry `json:"entries"`
}

// Top devolve as n melhores entradas. Empate desempata pela mais antiga, para
// a ordem não dançar entre leituras.
func (s *JSONStore) Top(_ context.Context, n int) ([]Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.load()
	if err != nil {
		return nil, err
	}
	entries := data.Entries
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Points != entries[j].Points {
			return entries[i].Points > entries[j].Points
		}
		return entries[i].At.Before(entries[j].At)
	})
	if n > 0 && len(entries) > n {
		entries = entries[:n]
	}
	return entries, nil
}

// Save acrescenta uma partida ao placar.
func (s *JSONStore) Save(_ context.Context, e Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := s.load()
	if err != nil {
		return err
	}
	data.Entries = append(data.Entries, e)
	return s.save(data)
}

func (s *JSONStore) load() (scoresFile, error) {
	var data scoresFile
	raw, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return data, nil
	}
	if err != nil {
		return data, fmt.Errorf("ler %s: %w", s.path, err)
	}
	if len(raw) == 0 {
		return data, nil
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return data, fmt.Errorf("decodificar %s: %w", s.path, err)
	}
	return data, nil
}

// save grava em temporário e renomeia, pelo mesmo motivo do store de
// identidades: crash no meio não deixa JSON truncado.
func (s *JSONStore) save(data scoresFile) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("criar diretório de %s: %w", s.path, err)
	}
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("codificar placar: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".scores-*.json")
	if err != nil {
		return fmt.Errorf("criar temporário: %w", err)
	}
	defer os.Remove(tmp.Name())

	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return fmt.Errorf("permissão do temporário: %w", err)
	}
	if _, err := tmp.Write(raw); err != nil {
		tmp.Close()
		return fmt.Errorf("escrever temporário: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("fechar temporário: %w", err)
	}
	if err := os.Rename(tmp.Name(), s.path); err != nil {
		return fmt.Errorf("renomear para %s: %w", s.path, err)
	}
	return nil
}
```

- [ ] **Step 5: Rodar os testes com detector de corrida**

Run: `export PATH=$HOME/.local/go/bin:$PATH && go test ./internal/score/ -v -count=1`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/score/
git commit -m "feat: placar em json ordenado por pontos

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

### Task 5: Contrato de tela, tema, teclas e fallback de tamanho

**Files:**
- Create: `internal/ui/screen.go`
- Create: `internal/ui/theme.go`
- Create: `internal/ui/keys.go`
- Create: `internal/ui/layout.go`
- Test: `internal/ui/layout_test.go`

**Interfaces:**
- Consumes: nada
- Produces: `ui.Screen` (interface), mensagens `ui.ShowScoresMsg`, `ui.StartGameMsg`, `ui.BackMsg`, `ui.QuitMsg`, `ui.NameSetMsg{Name string}`; comandos `ui.ShowScores`, `ui.StartGame`, `ui.Back`, `ui.Quit`, `ui.NameSet(string)`; `ui.Theme` com `NewTheme()`; `ui.Keys` com `ui.DefaultKeys()` e `(Keys).Help([]string) string`; `ui.Size(w, h int) (int, int)` e as constantes `ui.DefaultWidth`, `ui.DefaultHeight`, `ui.MinWidth`, `ui.MinHeight`

- [ ] **Step 1: Escrever o teste do fallback de tamanho**

Crie `internal/ui/layout_test.go`:

```go
package ui

import (
	"strings"
	"testing"
)

// Um cliente SSH pode anunciar PTY 0x0 — já aconteceu no probe do spec — e com
// largura zero nada é desenhado. O fallback é o que evita tela branca.
func TestSizeFallback(t *testing.T) {
	casos := []struct {
		nome   string
		w, h   int
		queroW int
		queroH int
	}{
		{"zero virou default", 0, 0, DefaultWidth, DefaultHeight},
		{"negativo virou default", -5, -1, DefaultWidth, DefaultHeight},
		{"abaixo do piso sobe pro piso", 10, 2, MinWidth, MinHeight},
		{"tamanho normal passa intacto", 120, 40, 120, 40},
		{"só a largura zerada", 0, 30, DefaultWidth, 30},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			w, h := Size(c.w, c.h)
			if w != c.queroW || h != c.queroH {
				t.Errorf("Size(%d, %d) = (%d, %d), quero (%d, %d)", c.w, c.h, w, h, c.queroW, c.queroH)
			}
		})
	}
}

func TestHelpJuntaAsDicas(t *testing.T) {
	got := DefaultKeys().Help([]string{"enter confirma", "esc volta"})
	if got == "" {
		t.Fatal("Help devolveu vazio")
	}
	for _, termo := range []string{"enter", "esc"} {
		if !strings.Contains(got, termo) {
			t.Errorf("Help() = %q, quero conter %q", got, termo)
		}
	}
}
```

- [ ] **Step 2: Rodar o teste e confirmar que falha**

Run: `export PATH=$HOME/.local/go/bin:$PATH && go test ./internal/ui/ -v -count=1`
Expected: FAIL com `undefined: Size`

- [ ] **Step 3: Implementar o contrato de tela**

Crie `internal/ui/screen.go`:

```go
// Package ui tem o contrato de tela, o tema e as telas do jogo. Não conhece
// persistência nem transporte: cada tela recebe só a dependência que usa.
package ui

import tea "charm.land/bubbletea/v2"

// Screen é uma tela do jogo. Devolve string, não tea.View: só o model raiz
// lida com tea.View, então as telas ficam triviais de testar e de compor.
type Screen interface {
	Init() tea.Cmd
	Update(tea.Msg) (Screen, tea.Cmd)
	View() string
	Title() string
}

// Mensagens de intenção. Uma tela diz o que quer; quem decide para onde ir é o
// model raiz, único dono do grafo de navegação. É o que permite nenhuma tela
// importar outra.
type (
	// ShowScoresMsg pede o placar.
	ShowScoresMsg struct{}
	// StartGameMsg pede uma partida nova.
	StartGameMsg struct{}
	// BackMsg sai da tela atual.
	BackMsg struct{}
	// QuitMsg encerra o jogo.
	QuitMsg struct{}
	// NameSetMsg informa que a pessoa se apresentou e o nome já está gravado.
	NameSetMsg struct{ Name string }
)

// Comandos prontos, para a tela não precisar escrever closure na mão.
func ShowScores() tea.Msg { return ShowScoresMsg{} }
func StartGame() tea.Msg  { return StartGameMsg{} }
func Back() tea.Msg       { return BackMsg{} }
func Quit() tea.Msg       { return QuitMsg{} }

// NameSet devolve um comando que anuncia o nome aceito.
func NameSet(name string) tea.Cmd {
	return func() tea.Msg { return NameSetMsg{Name: name} }
}
```

- [ ] **Step 4: Implementar tema, teclas e layout**

Crie `internal/ui/theme.go`:

```go
package ui

import "charm.land/lipgloss/v2"

// Theme concentra todo estilo num lugar, para a aparência mudar sem caçar
// cores espalhadas pelas telas. Sem Renderer: a lipgloss v2 não tem esse
// conceito, e o perfil de cor do cliente é resolvido pelo programa.
type Theme struct {
	Title    lipgloss.Style
	Item     lipgloss.Style
	Selected lipgloss.Style
	Dim      lipgloss.Style
	Error    lipgloss.Style
	Panel    lipgloss.Style
}

// NewTheme devolve o tema padrão.
func NewTheme() Theme {
	return Theme{
		Title:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#b58cff")),
		Item:     lipgloss.NewStyle().Foreground(lipgloss.Color("#d8d8d8")),
		Selected: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ffd866")),
		Dim:      lipgloss.NewStyle().Foreground(lipgloss.Color("#8a8a8a")),
		Error:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ff6b6b")),
		Panel:    lipgloss.NewStyle().Padding(1, 2),
	}
}
```

Crie `internal/ui/keys.go`:

```go
package ui

import "strings"

// Keys é o mapa de teclas, num lugar só para a barra de help nunca mentir
// sobre o que o jogo aceita.
type Keys struct {
	Up        []string
	Down      []string
	Confirm   []string
	Back      []string
	Toggle    []string
	Quit      []string
	ForceQuit []string
	theme     Theme
}

// DefaultKeys devolve o mapa do spec.
func DefaultKeys() Keys {
	return Keys{
		Up:        []string{"up", "k"},
		Down:      []string{"down", "j"},
		Confirm:   []string{"enter"},
		Back:      []string{"esc"},
		Toggle:    []string{"tab"},
		Quit:      []string{"q"},
		ForceQuit: []string{"ctrl+c"},
		theme:     NewTheme(),
	}
}

// Matches diz se a tecla pressionada está num dos conjuntos.
func (k Keys) Matches(pressed string, set []string) bool {
	for _, s := range set {
		if pressed == s {
			return true
		}
	}
	return false
}

// Help monta a barra de dicas no pé da tela.
func (k Keys) Help(hints []string) string {
	return k.theme.Dim.Render(strings.Join(hints, "  ·  "))
}
```

Crie `internal/ui/layout.go`:

```go
package ui

// Tamanhos de tela. O default entra quando o cliente não diz o tamanho, o piso
// evita layout impossível numa janela minúscula.
const (
	DefaultWidth  = 80
	DefaultHeight = 24
	MinWidth      = 40
	MinHeight     = 10
)

// Size normaliza o tamanho anunciado pelo cliente. Um PTY pode chegar 0x0 —
// visto na prática — e aí o Bubble Tea não desenharia nada.
func Size(w, h int) (int, int) {
	if w <= 0 {
		w = DefaultWidth
	}
	if h <= 0 {
		h = DefaultHeight
	}
	if w < MinWidth {
		w = MinWidth
	}
	if h < MinHeight {
		h = MinHeight
	}
	return w, h
}
```

- [ ] **Step 5: Rodar os testes e confirmar que passam**

Run: `export PATH=$HOME/.local/go/bin:$PATH && go test ./internal/ui/ -v -count=1`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/ui/
git commit -m "feat: contrato de tela, tema, teclas e fallback de tamanho

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

### Task 6: Sessão e model raiz com pilha de telas

**Files:**
- Create: `internal/app/session.go`
- Create: `internal/app/root.go`
- Test: `internal/app/root_test.go`

**Interfaces:**
- Consumes: `app.Config` (Task 1), `player.Identity`, `player.Store` (Task 2), `score.Store` (Task 4), `ui.Screen` e as mensagens (Task 5)
- Produces: `app.Session{Config, Identity, Players, Scores, Rand}`, `app.NewSession(cfg, id, players, scores) (*Session, error)`, `app.Root` implementando `tea.Model`, `app.NewRoot(sess *Session, initial ui.Screen) Root`, `(Root).Top() ui.Screen`, `(Root).Depth() int`

- [ ] **Step 1: Escrever o teste de navegação**

Crie `internal/app/root_test.go`:

```go
package app

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

// telaFake é uma tela de teste: devolve sempre o mesmo comando, para o teste
// controlar qual intenção a raiz vai receber.
type telaFake struct {
	nome string
	cmd  tea.Cmd
}

func (t telaFake) Init() tea.Cmd                       { return nil }
func (t telaFake) Update(tea.Msg) (ui.Screen, tea.Cmd) { return t, t.cmd }
func (t telaFake) View() string                        { return "tela " + t.nome }
func (t telaFake) Title() string                       { return t.nome }

func rootDeTeste(t *testing.T, inicial ui.Screen) Root {
	t.Helper()
	return NewRoot(sessaoDeTeste(t), inicial)
}

// Aplica a mensagem e, se vier comando, aplica o resultado dele também: é como
// o runtime do Bubble Tea se comporta, então o teste exercita o caminho real.
func passo(m Root, msg tea.Msg) Root {
	next, cmd := m.Update(msg)
	m = next.(Root)
	// Cap de iterações: uma tela que devolvesse comando a cada Update
	// travaria o teste em laço infinito em vez de falhar com mensagem.
	for i := 0; cmd != nil && i < 10; i++ {
		produced := cmd()
		if produced == nil {
			break
		}
		next, cmd = m.Update(produced)
		m = next.(Root)
	}
	return m
}

func TestStartGameEmpilhaAVezDeTrocar(t *testing.T) {
	m := rootDeTeste(t, telaFake{nome: "menu", cmd: ui.StartGame})
	if m.Depth() != 1 {
		t.Fatalf("profundidade inicial = %d, quero 1", m.Depth())
	}

	m = passo(m, tea.KeyPressMsg{})
	if m.Depth() != 2 {
		t.Fatalf("depois de StartGame a profundidade = %d, quero 2", m.Depth())
	}
	if got := m.Top().Title(); got != "jogo" {
		t.Errorf("topo = %q, quero jogo", got)
	}
}

func TestBackVoltaUmaTela(t *testing.T) {
	m := rootDeTeste(t, telaFake{nome: "menu", cmd: ui.ShowScores})
	m = passo(m, tea.KeyPressMsg{})
	if m.Depth() != 2 {
		t.Fatalf("profundidade = %d, quero 2", m.Depth())
	}

	m = passo(m, ui.BackMsg{})
	if m.Depth() != 1 {
		t.Errorf("depois de Back a profundidade = %d, quero 1", m.Depth())
	}
	if got := m.Top().Title(); got != "menu" {
		t.Errorf("topo = %q, quero menu", got)
	}
}

// O menu é a base da pilha: esc ali não pode encerrar a sessão por reflexo.
func TestBackNaBaseNaoDesempilha(t *testing.T) {
	m := rootDeTeste(t, telaFake{nome: "menu"})
	m = passo(m, ui.BackMsg{})
	if m.Depth() != 1 {
		t.Errorf("profundidade = %d, quero 1", m.Depth())
	}
}

// Depois de se apresentar, esc não pode levar de volta pra tela de nome.
func TestNameSetTrocaATelaSemDeixarVolta(t *testing.T) {
	m := rootDeTeste(t, telaFake{nome: "nome"})
	m = passo(m, ui.NameSetMsg{Name: "afonso"})

	if m.Depth() != 1 {
		t.Errorf("profundidade = %d, quero 1 (troca, não empilha)", m.Depth())
	}
	if got := m.Top().Title(); got != "menu" {
		t.Errorf("topo = %q, quero menu", got)
	}
	if m.Session().Identity.Name != "afonso" {
		t.Errorf("identidade da sessão = %q, quero afonso", m.Session().Identity.Name)
	}
}

func TestWindowSizeZeroCaiNoDefault(t *testing.T) {
	m := rootDeTeste(t, telaFake{nome: "menu"})

	// Primeiro um tamanho real, para o estado deixar de ser o default. Sem
	// este passo o teste passaria mesmo se o tratamento de WindowSizeMsg
	// fosse removido, porque NewRoot já começa no default — e aí uma
	// regressão que apagasse o fallback 0x0 passaria batida.
	m = passo(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	if w, h := m.Size(); w != 120 || h != 40 {
		t.Fatalf("Size() = (%d, %d), quero (120, 40)", w, h)
	}

	m = passo(m, tea.WindowSizeMsg{Width: 0, Height: 0})
	w, h := m.Size()
	if w != ui.DefaultWidth || h != ui.DefaultHeight {
		t.Errorf("Size() = (%d, %d), quero (%d, %d)", w, h, ui.DefaultWidth, ui.DefaultHeight)
	}
}

func TestViewUsaAltScreenEOTopoDaPilha(t *testing.T) {
	m := rootDeTeste(t, telaFake{nome: "menu"})
	v := m.View()
	if !v.AltScreen {
		t.Error("View().AltScreen = false, quero true")
	}
	if v.Content == "" {
		t.Error("View().Content vazio")
	}
}

// Empilhar não pode compartilhar o array de trás com a pilha anterior, senão
// uma tela empilhada depois de um Back sobrescreve o histórico.
func TestPilhaNaoCompartilhaArrayDeTras(t *testing.T) {
	m := rootDeTeste(t, telaFake{nome: "menu", cmd: ui.ShowScores})
	m = passo(m, tea.KeyPressMsg{})
	antesDoBack := m.Top().Title()
	m = passo(m, ui.BackMsg{})
	m = passo(m, ui.StartGameMsg{})

	if antesDoBack != "scores" {
		t.Fatalf("topo antes do back = %q, quero scores", antesDoBack)
	}
	if got := m.Top().Title(); got != "jogo" {
		t.Errorf("topo = %q, quero jogo", got)
	}
	if m.Depth() != 2 {
		t.Errorf("profundidade = %d, quero 2", m.Depth())
	}
}
```

- [ ] **Step 2: Escrever o helper de sessão de teste**

Crie `internal/app/testhelp_test.go`:

```go
package app

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/afonsodesenv/yatzee-terminal/internal/player"
	"github.com/afonsodesenv/yatzee-terminal/internal/score"
)

func sessaoDeTeste(t *testing.T) *Session {
	t.Helper()
	dir := t.TempDir()
	cfg := Config{DataDir: dir}
	sess, err := NewSession(
		cfg,
		player.Identity{ID: "local:teste", Kind: player.KindLocal},
		player.NewJSONStore(filepath.Join(dir, "players.json")),
		score.NewJSONStore(filepath.Join(dir, "scores.json")),
	)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	return sess
}

func ctx() context.Context { return context.Background() }
```

- [ ] **Step 3: Rodar o teste e confirmar que falha**

Run: `export PATH=$HOME/.local/go/bin:$PATH && go test ./internal/app/ -run TestStartGame -v`
Expected: FAIL com `undefined: NewRoot`

- [ ] **Step 4: Implementar a sessão**

Crie `internal/app/session.go`:

```go
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
```

- [ ] **Step 5: Implementar o model raiz**

Crie `internal/app/root.go`:

```go
package app

import (
	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui/game"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui/menu"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui/scores"
)

// Root é o model raiz: dono da pilha de telas e único lugar que conhece o
// grafo de navegação. Telas emitem intenção; a tradução em empilhar, voltar ou
// trocar acontece só aqui.
type Root struct {
	sess   *Session
	stack  []ui.Screen
	keys   ui.Keys
	width  int
	height int
}

// NewRoot monta a raiz com uma tela na base da pilha.
func NewRoot(sess *Session, initial ui.Screen) Root {
	return Root{
		sess:   sess,
		stack:  []ui.Screen{initial},
		keys:   ui.DefaultKeys(),
		width:  ui.DefaultWidth,
		height: ui.DefaultHeight,
	}
}

// Session dá acesso à sessão desta conexão.
func (m Root) Session() *Session { return m.sess }

// Top é a tela visível.
func (m Root) Top() ui.Screen { return m.stack[len(m.stack)-1] }

// Depth é o tamanho da pilha.
func (m Root) Depth() int { return len(m.stack) }

// Size devolve o tamanho já normalizado.
func (m Root) Size() (int, int) { return m.width, m.height }

// Init arranca a tela da base.
func (m Root) Init() tea.Cmd { return m.Top().Init() }

// Update roteia. A ordem importa: tamanho e saída global primeiro, intenções
// de navegação depois, e só então a mensagem desce para a tela do topo.
func (m Root) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = ui.Size(msg.Width, msg.Height)
		return m, nil

	case tea.KeyPressMsg:
		// Sai de qualquer lugar, sem passar pela tela. Vem do mapa de teclas,
		// não de um literal: keybinding mora num lugar só.
		if m.keys.Matches(msg.String(), m.keys.ForceQuit) {
			return m, tea.Quit
		}

	case ui.QuitMsg:
		return m, tea.Quit

	case ui.ShowScoresMsg:
		return m.push(scores.New(m.sess.Scores)), nil

	case ui.StartGameMsg:
		return m.push(game.New()), nil

	case ui.BackMsg:
		return m.pop(), nil

	case ui.NameSetMsg:
		// Troca em vez de empilhar: depois de se apresentar, esc não pode
		// levar de volta pra tela de nome.
		m.sess.Identity.Name = msg.Name
		return m.replace(menu.New(m.sess.Identity.Name)), nil
	}

	top := len(m.stack) - 1
	next, cmd := m.stack[top].Update(msg)
	m = m.copyStack()
	m.stack[top] = next
	return m, cmd
}

// View é o único lugar que lida com tea.View.
func (m Root) View() tea.View {
	v := tea.NewView(m.Top().View())
	v.AltScreen = true
	return v
}

// copyStack copia a pilha antes de modificar. Root é passado por valor, então
// sem a cópia duas versões do model compartilhariam o mesmo array de trás e
// uma escrita numa apareceria na outra.
func (m Root) copyStack() Root {
	next := make([]ui.Screen, len(m.stack))
	copy(next, m.stack)
	m.stack = next
	return m
}

func (m Root) push(s ui.Screen) Root {
	m = m.copyStack()
	m.stack = append(m.stack, s)
	return m
}

func (m Root) pop() Root {
	if len(m.stack) <= 1 {
		return m // base da pilha: voltar não encerra a sessão
	}
	m = m.copyStack()
	m.stack = m.stack[:len(m.stack)-1]
	return m
}

func (m Root) replace(s ui.Screen) Root {
	m = m.copyStack()
	m.stack[len(m.stack)-1] = s
	return m
}
```

Nota: este arquivo importa as telas das Tasks 7 a 10. Até elas existirem o pacote não compila — por isso o teste desta task roda junto com a Task 10. Para fechar esta task, crie os três arquivos de tela como esqueleto mínimo que satisfaz o compilador e preencha nas tasks seguintes:

Os três esqueletos, com as assinaturas exatas que `root.go` chama e os títulos
que os testes de navegação conferem. Cada um vira o arquivo completo da sua
task depois.

`internal/ui/menu/menu.go`:

```go
package menu

import (
	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

// Screen é o menu. Esqueleto: preenchido na Task 8.
type Screen struct{}

// New monta o menu sabendo quem joga.
func New(playerName string) Screen { return Screen{} }

func (s Screen) Init() tea.Cmd                       { return nil }
func (s Screen) Update(tea.Msg) (ui.Screen, tea.Cmd) { return s, nil }
func (s Screen) View() string                        { return "" }
func (s Screen) Title() string                       { return "menu" }
```

`internal/ui/scores/scores.go`:

```go
package scores

import (
	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/score"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

// Screen é o placar. Esqueleto: preenchido na Task 9.
type Screen struct{}

// New monta a tela com o store de que ela precisa.
func New(store score.Store) Screen { return Screen{} }

func (s Screen) Init() tea.Cmd                       { return nil }
func (s Screen) Update(tea.Msg) (ui.Screen, tea.Cmd) { return s, nil }
func (s Screen) View() string                        { return "" }
func (s Screen) Title() string                       { return "scores" }
```

`internal/ui/game/game.go`:

```go
package game

import (
	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

// Screen é a partida. Esqueleto: preenchido na Task 10.
type Screen struct{}

// New começa uma partida.
func New() Screen { return Screen{} }

func (s Screen) Init() tea.Cmd                       { return nil }
func (s Screen) Update(tea.Msg) (ui.Screen, tea.Cmd) { return s, nil }
func (s Screen) View() string                        { return "" }
func (s Screen) Title() string                       { return "jogo" }
```

- [ ] **Step 6: Rodar os testes de navegação**

Run: `export PATH=$HOME/.local/go/bin:$PATH && go test ./internal/app/ -v -count=1`
Expected: PASS nos sete testes de navegação

- [ ] **Step 7: Commit**

```bash
git add internal/app/ internal/ui/
git commit -m "feat: sessao e model raiz com pilha de telas

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

### Task 7: Tela de nome

**Files:**
- Create: `internal/ui/name/name.go`
- Test: `internal/ui/name/name_test.go`

**Interfaces:**
- Consumes: `player.Identity`, `player.Store`, `player.CleanName`, `player.NameRule`, `player.ErrNameTaken` (Tasks 2, 3), `ui.Screen`, `ui.NameSet`, `ui.Theme`, `ui.DefaultKeys` (Task 5)
- Produces: `name.New(id player.Identity, store player.Store) Screen` implementando `ui.Screen`, `(Screen).Value() string`, `(Screen).Error() string`

- [ ] **Step 1: Escrever o teste da tela de nome**

Crie `internal/ui/name/name_test.go`:

```go
package name

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/player"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

func storeDeTeste(t *testing.T) player.Store {
	t.Helper()
	return player.NewJSONStore(filepath.Join(t.TempDir(), "players.json"))
}

func telaDeTeste(t *testing.T, store player.Store) Screen {
	t.Helper()
	return New(player.Identity{ID: "key:abc", Kind: player.KindPubKey}, store)
}

// Digita texto tecla por tecla, como o runtime faria.
func digita(s Screen, texto string) Screen {
	for _, r := range texto {
		next, _ := s.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
		s = next.(Screen)
	}
	return s
}

func TestForcaMinusculoNaDigitacao(t *testing.T) {
	s := digita(telaDeTeste(t, storeDeTeste(t)), "AfOnSo")
	if got := s.Value(); got != "afonso" {
		t.Errorf("Value() = %q, quero afonso", got)
	}
}

func TestEnterComNomeValidoAnunciaONome(t *testing.T) {
	store := storeDeTeste(t)
	s := digita(telaDeTeste(t, store), "afonso")

	next, cmd := s.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter com nome válido não devolveu comando")
	}
	msg := cmd()
	set, ok := msg.(ui.NameSetMsg)
	if !ok {
		t.Fatalf("comando devolveu %T, quero ui.NameSetMsg", msg)
	}
	if set.Name != "afonso" {
		t.Errorf("NameSetMsg.Name = %q, quero afonso", set.Name)
	}
	if e := next.(Screen).Error(); e != "" {
		t.Errorf("erro = %q, quero vazio", e)
	}

	// O nome tem que estar gravado, não só anunciado.
	p, ok, err := store.Get(context.Background(), "key:abc")
	if err != nil || !ok {
		t.Fatalf("Get = %v, %v, %v", p, ok, err)
	}
	if p.Name != "afonso" {
		t.Errorf("nome gravado = %q, quero afonso", p.Name)
	}
}

func TestEnterComNomeInvalidoMostraErroEPreservaOTexto(t *testing.T) {
	s := digita(telaDeTeste(t, storeDeTeste(t)), "ab")

	next, cmd := s.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd != nil {
		if msg := cmd(); msg != nil {
			if _, ok := msg.(ui.NameSetMsg); ok {
				t.Fatal("nome inválido foi aceito")
			}
		}
	}
	tela := next.(Screen)
	if tela.Error() == "" {
		t.Error("nome inválido não gerou mensagem de erro")
	}
	if tela.Value() != "ab" {
		t.Errorf("Value() = %q, quero preservar ab", tela.Value())
	}
}

func TestEnterComNomeEmUsoMostraErro(t *testing.T) {
	store := storeDeTeste(t)
	if _, err := store.Claim(context.Background(), "key:outro", "afonso"); err != nil {
		t.Fatalf("Claim de preparação: %v", err)
	}

	s := digita(telaDeTeste(t, store), "afonso")
	next, _ := s.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	tela := next.(Screen)

	if !strings.Contains(tela.Error(), "uso") {
		t.Errorf("Error() = %q, quero mencionar que está em uso", tela.Error())
	}
	if tela.Value() != "afonso" {
		t.Errorf("Value() = %q, quero preservar o texto", tela.Value())
	}
}

// Convidado não reserva nome, mas também não pode se passar por registrado.
func TestConvidadoComNomeDeRegistradoERecusado(t *testing.T) {
	store := storeDeTeste(t)
	if _, err := store.Claim(context.Background(), "key:outro", "afonso"); err != nil {
		t.Fatalf("Claim de preparação: %v", err)
	}

	s := New(player.Identity{ID: "guest:1", Kind: player.KindGuest}, store)
	s = digita(s, "afonso")
	next, _ := s.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	if next.(Screen).Error() == "" {
		t.Error("convidado usou nome de jogador registrado sem erro")
	}
}

func TestConvidadoNaoEPersistido(t *testing.T) {
	store := storeDeTeste(t)
	s := New(player.Identity{ID: "guest:1", Kind: player.KindGuest}, store)
	s = digita(s, "visitante")

	next, cmd := s.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter não devolveu comando")
	}
	if _, ok := cmd().(ui.NameSetMsg); !ok {
		t.Fatal("convidado com nome livre deveria ser aceito")
	}
	_ = next

	if _, ok, err := store.Get(context.Background(), "guest:1"); err != nil || ok {
		t.Errorf("convidado foi persistido: ok=%v err=%v", ok, err)
	}
}

func TestViewMostraARegra(t *testing.T) {
	v := telaDeTeste(t, storeDeTeste(t)).View()
	if !strings.Contains(v, "16") {
		t.Errorf("View() não mostra a regra do nome:\n%s", v)
	}
}
```

- [ ] **Step 2: Rodar o teste e confirmar que falha**

Run: `export PATH=$HOME/.local/go/bin:$PATH && go test ./internal/ui/name/ -v -count=1`
Expected: FAIL com `undefined: New`

- [ ] **Step 3: Implementar a tela de nome**

Crie `internal/ui/name/name.go`:

```go
// Package name tem a tela que pergunta o nome na primeira vez.
package name

import (
	"context"
	"errors"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/player"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

// Screen pergunta o nome e só libera quando ele é válido e livre.
type Screen struct {
	id    player.Identity
	store player.Store
	input textinput.Model
	err   string
	theme ui.Theme
	keys  ui.Keys
}

// New monta a tela para uma identidade ainda sem nome.
func New(id player.Identity, store player.Store) Screen {
	ti := textinput.New()
	ti.Prompt = "> "
	ti.Placeholder = "seu-nome"
	ti.CharLimit = player.NameMaxLen
	ti.SetWidth(player.NameMaxLen + 4)
	ti.Focus()

	return Screen{
		id:    id,
		store: store,
		input: ti,
		theme: ui.NewTheme(),
		keys:  ui.DefaultKeys(),
	}
}

// Value é o texto atual do campo.
func (s Screen) Value() string { return s.input.Value() }

// Error é a mensagem de erro visível, vazia quando não há.
func (s Screen) Error() string { return s.err }

// Title identifica a tela.
func (s Screen) Title() string { return "nome" }

// Init pisca o cursor.
func (s Screen) Init() tea.Cmd { return textinput.Blink }

// Update trata o enter e delega o resto ao campo de texto.
func (s Screen) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		if s.keys.Matches(key.String(), s.keys.Confirm) {
			return s.confirm()
		}
	}

	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)

	// Baixa a caixa depois do campo processar, não interceptando a tecla:
	// assim texto colado também entra minúsculo, e a pessoa vê o nome real
	// enquanto digita em vez de descobrir a conversão no envio.
	if v := s.input.Value(); v != strings.ToLower(v) {
		s.input.SetValue(strings.ToLower(v))
	}
	return s, cmd
}

// confirm valida, reivindica e anuncia. Erro aqui não é falha do sistema: a
// tela mantém o texto e espera outro nome.
func (s Screen) confirm() (ui.Screen, tea.Cmd) {
	ctx := context.Background()

	name, err := player.CleanName(s.input.Value())
	if err != nil {
		s.err = player.NameRule()
		return s, nil
	}

	// Convidado não reserva nome, mas também não pode se passar por alguém
	// registrado no placar.
	if !s.id.Persistent() {
		taken, err := s.store.NameTaken(ctx, name)
		if err != nil {
			s.err = "não consegui verificar o nome agora"
			return s, nil
		}
		if taken {
			s.err = "esse nome já está em uso"
			return s, nil
		}
		s.err = ""
		return s, ui.NameSet(name)
	}

	if _, err := s.store.Claim(ctx, s.id.ID, name); err != nil {
		if errors.Is(err, player.ErrNameTaken) {
			s.err = "esse nome já está em uso"
		} else {
			s.err = "não consegui gravar o nome agora"
		}
		return s, nil
	}
	s.err = ""
	return s, ui.NameSet(name)
}

// View desenha o campo, a regra e o erro.
func (s Screen) View() string {
	b := &strings.Builder{}
	b.WriteString(s.theme.Title.Render("bem-vindo ao yatzee"))
	b.WriteString("\n\ncomo você quer ser chamado?\n\n")
	b.WriteString(s.input.View())
	b.WriteString("\n\n")
	b.WriteString(s.theme.Dim.Render(player.NameRule()))
	if s.err != "" {
		b.WriteString("\n")
		b.WriteString(s.theme.Error.Render(s.err))
	}
	b.WriteString("\n\n")
	b.WriteString(s.keys.Help([]string{"enter confirma", "ctrl+c sai"}))
	return s.theme.Panel.Render(b.String())
}
```

- [ ] **Step 4: Rodar os testes e confirmar que passam**

Run: `export PATH=$HOME/.local/go/bin:$PATH && go test ./internal/ui/name/ -v -count=1`
Expected: PASS nos sete testes

- [ ] **Step 5: Commit**

```bash
git add internal/ui/name/
git commit -m "feat: tela de nome com minusculo forcado e nome unico

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

### Task 8: Tela de menu

**Files:**
- Modify: `internal/ui/menu/menu.go` (substituir o esqueleto da Task 6)
- Test: `internal/ui/menu/menu_test.go`

**Interfaces:**
- Consumes: `ui.Screen`, `ui.StartGame`, `ui.ShowScores`, `ui.Quit`, `ui.Theme`, `ui.DefaultKeys` (Task 5)
- Produces: `menu.New(playerName string) Screen` implementando `ui.Screen`, `(Screen).Cursor() int`

- [ ] **Step 1: Escrever o teste do menu**

Crie `internal/ui/menu/menu_test.go`:

```go
package menu

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

func tecla(s Screen, nome string) (Screen, tea.Cmd) {
	next, cmd := s.Update(teclaMsg(nome))
	return next.(Screen), cmd
}

func teclaMsg(nome string) tea.KeyPressMsg {
	switch nome {
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "down":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	case "up":
		return tea.KeyPressMsg{Code: tea.KeyUp}
	default:
		r := rune(nome[0])
		return tea.KeyPressMsg{Code: r, Text: nome}
	}
}

func TestCursorAndaEParaNasPontas(t *testing.T) {
	s := New("afonso")
	if s.Cursor() != 0 {
		t.Fatalf("cursor inicial = %d, quero 0", s.Cursor())
	}

	s, _ = tecla(s, "down")
	if s.Cursor() != 1 {
		t.Errorf("cursor após down = %d, quero 1", s.Cursor())
	}
	s, _ = tecla(s, "j")
	if s.Cursor() != 2 {
		t.Errorf("cursor após j = %d, quero 2", s.Cursor())
	}
	s, _ = tecla(s, "down") // já no fim
	if s.Cursor() != 2 {
		t.Errorf("cursor após down no fim = %d, quero 2", s.Cursor())
	}
	s, _ = tecla(s, "k")
	if s.Cursor() != 1 {
		t.Errorf("cursor após k = %d, quero 1", s.Cursor())
	}

	// "up" literal tem que mover de verdade antes de o teste do limite
	// significar algo: se o mapa de teclas perdesse "up" e ficasse só com
	// "k", um "up" no topo continuaria "parando" em 0 por não fazer nada.
	s, _ = tecla(s, "up")
	if s.Cursor() != 0 {
		t.Errorf("cursor após up = %d, quero 0", s.Cursor())
	}

	s, _ = tecla(s, "up") // agora sim: já no topo, não deve passar do limite
	if s.Cursor() != 0 {
		t.Errorf("cursor após up no topo = %d, quero 0", s.Cursor())
	}
}

func TestEnterEmitAIntencaoDoItemSelecionado(t *testing.T) {
	casos := []struct {
		descidas int
		quero    tea.Msg
	}{
		{0, ui.StartGameMsg{}},
		{1, ui.ShowScoresMsg{}},
		{2, ui.QuitMsg{}},
	}
	for _, c := range casos {
		s := New("afonso")
		for range c.descidas {
			s, _ = tecla(s, "down")
		}
		_, cmd := tecla(s, "enter")
		if cmd == nil {
			t.Fatalf("item %d: enter não devolveu comando", c.descidas)
		}
		if got := cmd(); got != c.quero {
			t.Errorf("item %d emitiu %T, quero %T", c.descidas, got, c.quero)
		}
	}
}

func TestTeclaQSai(t *testing.T) {
	_, cmd := tecla(New("afonso"), "q")
	if cmd == nil {
		t.Fatal("q não devolveu comando")
	}
	if _, ok := cmd().(ui.QuitMsg); !ok {
		t.Errorf("q emitiu %T, quero ui.QuitMsg", cmd())
	}
}

// esc no menu não faz nada: é a base da pilha, e sair exige q ou ctrl+c.
func TestEscNaoEmiteNada(t *testing.T) {
	_, cmd := New("afonso").Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd != nil {
		t.Errorf("esc no menu devolveu comando %v", cmd())
	}
}

func TestViewMostraONomeEOsItens(t *testing.T) {
	v := New("afonso").View()
	for _, termo := range []string{"afonso", "novo jogo", "placar", "sair"} {
		if !strings.Contains(v, termo) {
			t.Errorf("View() não contém %q:\n%s", termo, v)
		}
	}
}
```

- [ ] **Step 2: Rodar o teste e confirmar que falha**

Run: `export PATH=$HOME/.local/go/bin:$PATH && go test ./internal/ui/menu/ -v -count=1`
Expected: FAIL — o esqueleto não tem `Cursor()`

- [ ] **Step 3: Implementar o menu**

Substitua `internal/ui/menu/menu.go`:

```go
// Package menu tem a tela inicial do jogo.
package menu

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

type item struct {
	label  string
	intent tea.Cmd
}

// Screen é o menu principal.
type Screen struct {
	playerName string
	items      []item
	cursor     int
	theme      ui.Theme
	keys       ui.Keys
}

// New monta o menu já sabendo quem está jogando.
func New(playerName string) Screen {
	return Screen{
		playerName: playerName,
		items: []item{
			{label: "novo jogo", intent: ui.StartGame},
			{label: "placar", intent: ui.ShowScores},
			{label: "sair", intent: ui.Quit},
		},
		theme: ui.NewTheme(),
		keys:  ui.DefaultKeys(),
	}
}

// Cursor é o índice do item selecionado.
func (s Screen) Cursor() int { return s.cursor }

// Title identifica a tela.
func (s Screen) Title() string { return "menu" }

// Init não precisa de nada.
func (s Screen) Init() tea.Cmd { return nil }

// Update move o cursor e emite a intenção do item escolhido.
func (s Screen) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return s, nil
	}
	pressed := key.String()

	switch {
	case s.keys.Matches(pressed, s.keys.Up):
		if s.cursor > 0 {
			s.cursor--
		}
	case s.keys.Matches(pressed, s.keys.Down):
		if s.cursor < len(s.items)-1 {
			s.cursor++
		}
	case s.keys.Matches(pressed, s.keys.Confirm):
		return s, s.items[s.cursor].intent
	case s.keys.Matches(pressed, s.keys.Quit):
		return s, ui.Quit
	}
	return s, nil
}

// View desenha o título, os itens e o help.
func (s Screen) View() string {
	b := &strings.Builder{}
	b.WriteString(s.theme.Title.Render("yatzee"))
	if s.playerName != "" {
		b.WriteString(s.theme.Dim.Render(fmt.Sprintf("   jogando como %s", s.playerName)))
	}
	b.WriteString("\n\n")

	for i, it := range s.items {
		if i == s.cursor {
			b.WriteString(s.theme.Selected.Render("› " + it.label))
		} else {
			b.WriteString(s.theme.Item.Render("  " + it.label))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(s.keys.Help([]string{"↑↓ move", "enter escolhe", "q sai"}))
	return s.theme.Panel.Render(b.String())
}
```

- [ ] **Step 4: Rodar os testes e confirmar que passam**

Run: `export PATH=$HOME/.local/go/bin:$PATH && go test ./internal/ui/menu/ -v -count=1`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ui/menu/
git commit -m "feat: tela de menu com intencoes de navegacao

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

### Task 9: Tela de placar

**Files:**
- Modify: `internal/ui/scores/scores.go` (substituir o esqueleto da Task 6)
- Test: `internal/ui/scores/scores_test.go`

**Interfaces:**
- Consumes: `score.Store`, `score.Entry` (Task 4), `ui.Screen`, `ui.Back`, `ui.Theme`, `ui.DefaultKeys` (Task 5)
- Produces: `scores.New(store score.Store) Screen` implementando `ui.Screen`, `scores.LoadedMsg{Entries []score.Entry, Err error}`

- [ ] **Step 1: Escrever o teste do placar**

Crie `internal/ui/scores/scores_test.go`:

```go
package scores

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/score"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

// storeFake evita I/O no teste de tela e permite simular falha de leitura.
type storeFake struct {
	entries []score.Entry
	err     error
}

func (f storeFake) Top(context.Context, int) ([]score.Entry, error) {
	return f.entries, f.err
}
func (f storeFake) Save(context.Context, score.Entry) error { return nil }

func TestInitCarregaOPlacar(t *testing.T) {
	base := time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC)
	store := storeFake{entries: []score.Entry{
		{Name: "ana", Points: 300, At: base},
	}}

	cmd := New(store).Init()
	if cmd == nil {
		t.Fatal("Init não devolveu comando de carga")
	}
	msg, ok := cmd().(LoadedMsg)
	if !ok {
		t.Fatalf("Init emitiu %T, quero LoadedMsg", cmd())
	}
	if len(msg.Entries) != 1 || msg.Entries[0].Name != "ana" {
		t.Errorf("LoadedMsg.Entries = %+v", msg.Entries)
	}
}

func TestViewListaAsEntradas(t *testing.T) {
	base := time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC)
	s := New(storeFake{})
	next, _ := s.Update(LoadedMsg{Entries: []score.Entry{
		{Name: "ana", Points: 300, At: base},
		{Name: "bob", Points: 120, At: base},
	}})

	v := next.View()
	for _, termo := range []string{"ana", "300", "bob", "120"} {
		if !strings.Contains(v, termo) {
			t.Errorf("View() não contém %q:\n%s", termo, v)
		}
	}
}

// Sem gameplay ainda não existem partidas: a tela precisa dizer isso em vez de
// mostrar um quadro vazio.
func TestViewComPlacarVazioExplica(t *testing.T) {
	s := New(storeFake{})
	next, _ := s.Update(LoadedMsg{Entries: nil})

	if !strings.Contains(next.View(), "nenhuma partida") {
		t.Errorf("View() vazia não explica o estado:\n%s", next.View())
	}
}

func TestViewComErroDeLeitura(t *testing.T) {
	s := New(storeFake{})
	next, _ := s.Update(LoadedMsg{Err: errors.New("disco pegou fogo")})

	if !strings.Contains(next.View(), "não consegui") {
		t.Errorf("View() não reporta a falha:\n%s", next.View())
	}
}

func TestEscVolta(t *testing.T) {
	_, cmd := New(storeFake{}).Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("esc não devolveu comando")
	}
	if _, ok := cmd().(ui.BackMsg); !ok {
		t.Errorf("esc emitiu %T, quero ui.BackMsg", cmd())
	}
}
```

- [ ] **Step 2: Rodar o teste e confirmar que falha**

Run: `export PATH=$HOME/.local/go/bin:$PATH && go test ./internal/ui/scores/ -v -count=1`
Expected: FAIL com `undefined: LoadedMsg`

- [ ] **Step 3: Implementar a tela de placar**

Substitua `internal/ui/scores/scores.go`:

```go
// Package scores mostra os melhores jogadores.
package scores

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/score"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

// Quantas linhas o placar mostra.
const topN = 10

// LoadedMsg traz o resultado da leitura do placar.
type LoadedMsg struct {
	Entries []score.Entry
	Err     error
}

// Screen mostra o placar. Recebe só o store de que precisa, não a sessão
// inteira: é o que mantém esta tela testável com um fake de dois métodos.
type Screen struct {
	store   score.Store
	entries []score.Entry
	err     error
	loaded  bool
	theme   ui.Theme
	keys    ui.Keys
}

// New monta a tela.
func New(store score.Store) Screen {
	return Screen{
		store: store,
		theme: ui.NewTheme(),
		keys:  ui.DefaultKeys(),
	}
}

// Title identifica a tela.
func (s Screen) Title() string { return "scores" }

// Init dispara a leitura fora da thread de render: I/O não bloqueia a UI.
func (s Screen) Init() tea.Cmd {
	store := s.store
	return func() tea.Msg {
		entries, err := store.Top(context.Background(), topN)
		return LoadedMsg{Entries: entries, Err: err}
	}
}

// Update recebe o resultado da carga e trata o esc.
func (s Screen) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case LoadedMsg:
		s.entries, s.err, s.loaded = msg.Entries, msg.Err, true
		return s, nil
	case tea.KeyPressMsg:
		if s.keys.Matches(msg.String(), s.keys.Back) {
			return s, ui.Back
		}
	}
	return s, nil
}

// View desenha a lista, o estado vazio ou a falha.
func (s Screen) View() string {
	b := &strings.Builder{}
	b.WriteString(s.theme.Title.Render("placar"))
	b.WriteString("\n\n")

	switch {
	case s.err != nil:
		b.WriteString(s.theme.Error.Render("não consegui ler o placar agora"))
	case !s.loaded:
		b.WriteString(s.theme.Dim.Render("carregando..."))
	case len(s.entries) == 0:
		b.WriteString(s.theme.Dim.Render("nenhuma partida registrada ainda"))
	default:
		for i, e := range s.entries {
			line := fmt.Sprintf("%2d. %-16s %6d", i+1, e.Name, e.Points)
			b.WriteString(s.theme.Item.Render(line))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(s.keys.Help([]string{"esc volta"}))
	return s.theme.Panel.Render(b.String())
}
```

- [ ] **Step 4: Rodar os testes e confirmar que passam**

Run: `export PATH=$HOME/.local/go/bin:$PATH && go test ./internal/ui/scores/ -v -count=1`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ui/scores/
git commit -m "feat: tela de placar com estado vazio e de falha

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

### Task 10: Tela de jogo com as duas fases

**Files:**
- Modify: `internal/ui/game/game.go` (substituir o esqueleto da Task 6)
- Create: `internal/ui/game/round.go`
- Create: `internal/ui/game/shop.go`
- Test: `internal/ui/game/game_test.go`

**Interfaces:**
- Consumes: `ui.Screen`, `ui.Back`, `ui.Theme`, `ui.DefaultKeys` (Task 5)
- Produces: `game.New() Screen` implementando `ui.Screen`, `game.Phase` (`PhaseRound`, `PhaseShop`), `(Screen).Phase() Phase`

- [ ] **Step 1: Escrever o teste das fases**

Crie `internal/ui/game/game_test.go`:

```go
package game

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

func TestComecaNaRodada(t *testing.T) {
	if got := New().Phase(); got != PhaseRound {
		t.Errorf("fase inicial = %v, quero PhaseRound", got)
	}
}

func TestTabAlternaAsFases(t *testing.T) {
	s := New()

	next, _ := s.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	s = next.(Screen)
	if s.Phase() != PhaseShop {
		t.Fatalf("depois do tab = %v, quero PhaseShop", s.Phase())
	}

	next, _ = s.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	s = next.(Screen)
	if s.Phase() != PhaseRound {
		t.Errorf("depois do segundo tab = %v, quero PhaseRound", s.Phase())
	}
}

func TestViewSegueAFase(t *testing.T) {
	// Afirma texto exclusivo de cada fase, e a ausência do texto da outra. A
	// barra de help diz "tab alterna rodada e loja", então procurar só por
	// "rodada" ou "loja" passaria em qualquer fase — inclusive se o View
	// ignorasse a fase e sempre desenhasse a rodada, que é justamente o
	// comportamento que este teste existe para garantir.
	const textoRodada = "as mecânicas de dados entram aqui"
	const textoLoja = "os itens entram aqui"

	s := New()
	v := s.View()
	if !strings.Contains(v, textoRodada) {
		t.Errorf("view da rodada não mostra o conteúdo da rodada:\n%s", v)
	}
	if strings.Contains(v, textoLoja) {
		t.Errorf("view da rodada mostra conteúdo da loja:\n%s", v)
	}

	next, _ := s.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	v = next.View()
	if !strings.Contains(v, textoLoja) {
		t.Errorf("view da loja não mostra o conteúdo da loja:\n%s", v)
	}
	if strings.Contains(v, textoRodada) {
		t.Errorf("view da loja mostra conteúdo da rodada:\n%s", v)
	}
}

// Mensagem que não é tecla não pode mexer na fase: a raiz manda WindowSizeMsg
// para a tela do topo a cada resize.
func TestMensagemNaoTeclaNaoMudaAFase(t *testing.T) {
	s := New()
	next, cmd := s.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	if got := next.(Screen).Phase(); got != PhaseRound {
		t.Errorf("fase = %v, quero PhaseRound", got)
	}
	if cmd != nil {
		t.Errorf("mensagem não-tecla devolveu comando %v", cmd())
	}
}

func TestEscVoltaProMenu(t *testing.T) {
	_, cmd := New().Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("esc não devolveu comando")
	}
	if _, ok := cmd().(ui.BackMsg); !ok {
		t.Errorf("esc emitiu %T, quero ui.BackMsg", cmd())
	}
}
```

- [ ] **Step 2: Rodar o teste e confirmar que falha**

Run: `export PATH=$HOME/.local/go/bin:$PATH && go test ./internal/ui/game/ -v -count=1`
Expected: FAIL com `undefined: PhaseRound`

- [ ] **Step 3: Implementar as duas fases**

Crie `internal/ui/game/round.go`:

```go
package game

import (
	"strings"

	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

// roundView é a fase de rodadas. Placeholder: as mecânicas entram depois, e é
// este arquivo que vai crescer.
type roundView struct{ theme ui.Theme }

func (v roundView) render() string {
	b := &strings.Builder{}
	b.WriteString(v.theme.Title.Render("rodada"))
	b.WriteString("\n\n")
	b.WriteString(v.theme.Dim.Render("as mecânicas de dados entram aqui"))
	b.WriteString("\n")
	return b.String()
}
```

Crie `internal/ui/game/shop.go`:

```go
package game

import (
	"strings"

	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

// shopView é a fase de loja. Placeholder, como a rodada.
type shopView struct{ theme ui.Theme }

func (v shopView) render() string {
	b := &strings.Builder{}
	b.WriteString(v.theme.Title.Render("loja"))
	b.WriteString("\n\n")
	b.WriteString(v.theme.Dim.Render("os itens entram aqui"))
	b.WriteString("\n")
	return b.String()
}
```

Substitua `internal/ui/game/game.go`:

```go
// Package game tem a tela de partida, dona das duas fases.
package game

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
)

// Phase é a fase da partida.
type Phase int

const (
	// PhaseRound é a fase de jogar os dados.
	PhaseRound Phase = iota
	// PhaseShop é a loja entre rodadas.
	PhaseShop
)

// Screen é a partida. Rodada e loja não são telas independentes na pilha: são
// fases, e esta tela é dona das duas, o que mantém o estado da partida num
// lugar só quando as mecânicas chegarem.
type Screen struct {
	phase Phase
	round roundView
	shop  shopView
	theme ui.Theme
	keys  ui.Keys
}

// New começa uma partida na fase de rodada.
func New() Screen {
	theme := ui.NewTheme()
	return Screen{
		phase: PhaseRound,
		round: roundView{theme: theme},
		shop:  shopView{theme: theme},
		theme: theme,
		keys:  ui.DefaultKeys(),
	}
}

// Phase é a fase atual.
func (s Screen) Phase() Phase { return s.phase }

// Title identifica a tela.
func (s Screen) Title() string { return "jogo" }

// Init não precisa de nada ainda.
func (s Screen) Init() tea.Cmd { return nil }

// Update alterna as fases e trata a volta.
func (s Screen) Update(msg tea.Msg) (ui.Screen, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return s, nil
	}
	pressed := key.String()

	switch {
	case s.keys.Matches(pressed, s.keys.Toggle):
		if s.phase == PhaseRound {
			s.phase = PhaseShop
		} else {
			s.phase = PhaseRound
		}
	case s.keys.Matches(pressed, s.keys.Back):
		return s, ui.Back
	}
	return s, nil
}

// View desenha a fase atual.
func (s Screen) View() string {
	b := &strings.Builder{}
	if s.phase == PhaseRound {
		b.WriteString(s.round.render())
	} else {
		b.WriteString(s.shop.render())
	}
	b.WriteString("\n")
	b.WriteString(s.keys.Help([]string{"tab alterna rodada e loja", "esc volta ao menu"}))
	return s.theme.Panel.Render(b.String())
}
```

- [ ] **Step 4: Rodar a suíte inteira**

Run: `export PATH=$HOME/.local/go/bin:$PATH && go test ./... && go vet ./...`
Expected: PASS em tudo, incluindo os testes de navegação da Task 6 que dependiam destas telas

- [ ] **Step 5: Commit**

```bash
git add internal/ui/game/
git commit -m "feat: tela de jogo com fases de rodada e loja

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

### Task 11: Jogo local de ponta a ponta

**Files:**
- Create: `internal/app/local.go`
- Create: `internal/app/firstscreen.go`
- Create: `cmd/yatzee/main.go`
- Create: `README.md`
- Test: `internal/app/firstscreen_test.go`

**Interfaces:**
- Consumes: tudo das tasks anteriores
- Produces: `app.FirstScreen(sess *Session) ui.Screen`, `app.PlayLocal(ctx context.Context) error`, `localIdentityFor(ctx context.Context, osUser string, store player.Store) (player.Identity, error)` (não exportada: só o pacote `app` monta identidade)

- [ ] **Step 1: Escrever o teste da primeira tela**

Crie `internal/app/firstscreen_test.go`:

```go
package app

import (
	"testing"

	"github.com/afonsodesenv/yatzee-terminal/internal/player"
)

// Identidade sem nome cai na tela de nome; com nome vai direto pro menu. É o
// comportamento que faz a segunda execução não perguntar nada.
func TestFirstScreen(t *testing.T) {
	casos := []struct {
		nome     string
		identity player.Identity
		quero    string
	}{
		{"sem nome pergunta", player.Identity{ID: "key:a", Kind: player.KindPubKey}, "nome"},
		{"com nome vai pro menu", player.Identity{ID: "key:a", Kind: player.KindPubKey, Name: "afonso"}, "menu"},
		{"convidado sem nome pergunta", player.Identity{ID: "guest:1", Kind: player.KindGuest}, "nome"},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			sess := sessaoDeTeste(t)
			sess.Identity = c.identity
			if got := FirstScreen(sess).Title(); got != c.quero {
				t.Errorf("FirstScreen().Title() = %q, quero %q", got, c.quero)
			}
		})
	}
}

func TestLocalIdentityLembraONomeGravado(t *testing.T) {
	sess := sessaoDeTeste(t)
	if _, err := sess.Players.Claim(ctx(), "local:fulano", "afonso"); err != nil {
		t.Fatalf("Claim: %v", err)
	}

	id, err := localIdentityFor(ctx(), "fulano", sess.Players)
	if err != nil {
		t.Fatalf("localIdentityFor: %v", err)
	}
	if id.ID != "local:fulano" {
		t.Errorf("ID = %q, quero local:fulano", id.ID)
	}
	if id.Kind != player.KindLocal {
		t.Errorf("Kind = %q, quero local", id.Kind)
	}
	if id.Name != "afonso" {
		t.Errorf("Name = %q, quero afonso (nome gravado)", id.Name)
	}
}

func TestLocalIdentitySemRegistroVemSemNome(t *testing.T) {
	sess := sessaoDeTeste(t)
	id, err := localIdentityFor(ctx(), "novato", sess.Players)
	if err != nil {
		t.Fatalf("localIdentityFor: %v", err)
	}
	if id.Name != "" {
		t.Errorf("Name = %q, quero vazio", id.Name)
	}
}
```

- [ ] **Step 2: Rodar o teste e confirmar que falha**

Run: `export PATH=$HOME/.local/go/bin:$PATH && go test ./internal/app/ -run TestFirstScreen -v`
Expected: FAIL com `undefined: FirstScreen`

- [ ] **Step 3: Implementar a escolha da primeira tela**

Crie `internal/app/firstscreen.go`:

```go
package app

import (
	"context"
	"fmt"

	"github.com/afonsodesenv/yatzee-terminal/internal/player"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui/menu"
	"github.com/afonsodesenv/yatzee-terminal/internal/ui/name"
)

// FirstScreen escolhe por onde a sessão começa: quem já tem nome não é
// perguntado de novo.
func FirstScreen(sess *Session) ui.Screen {
	if sess.Identity.Name == "" {
		return name.New(sess.Identity, sess.Players)
	}
	return menu.New(sess.Identity.Name)
}

// localIdentityFor monta a identidade do jogo local a partir do usuário do SO.
func localIdentityFor(ctx context.Context, osUser string, store player.Store) (player.Identity, error) {
	id := player.Identity{
		ID:   "local:" + osUser,
		Kind: player.KindLocal,
	}
	p, ok, err := store.Get(ctx, id.ID)
	if err != nil {
		return player.Identity{}, fmt.Errorf("buscar identidade local: %w", err)
	}
	if ok {
		id.Name = p.Name
	}
	return id, nil
}
```

- [ ] **Step 4: Implementar o modo local**

Crie `internal/app/local.go`:

```go
package app

import (
	"context"
	"fmt"
	"os"
	"os/user"

	tea "charm.land/bubbletea/v2"

	"github.com/afonsodesenv/yatzee-terminal/internal/player"
	"github.com/afonsodesenv/yatzee-terminal/internal/score"
)

// PlayLocal roda o jogo no terminal da própria máquina.
func PlayLocal(ctx context.Context) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(cfg.DataDir, 0o700); err != nil {
		return fmt.Errorf("criar %s: %w", cfg.DataDir, err)
	}

	players := player.NewJSONStore(cfg.PlayersPath())
	scores := score.NewJSONStore(cfg.ScoresPath())

	id, err := localIdentityFor(ctx, osUsername(), players)
	if err != nil {
		return err
	}
	sess, err := NewSession(cfg, id, players, scores)
	if err != nil {
		return err
	}
	if id.Name != "" {
		// Falha em marcar presença não justifica recusar a partida.
		_ = players.Touch(ctx, id.ID)
	}

	// WithContext liga o cancelamento ao programa. Sem isto o ctx chega
	// apenas aos stores, e um SIGTERM nao encerraria a partida local.
	prog := tea.NewProgram(NewRoot(sess, FirstScreen(sess)), tea.WithContext(ctx))
	if _, err := prog.Run(); err != nil {
		return fmt.Errorf("rodar o jogo: %w", err)
	}
	return nil
}

// osUsername identifica o jogador local. Cai num nome fixo se o SO não
// informar: pior caso, todo mundo na máquina compartilha a identidade local.
func osUsername() string {
	if u, err := user.Current(); err == nil && u.Username != "" {
		return u.Username
	}
	if v := os.Getenv("USER"); v != "" {
		return v
	}
	return "jogador"
}
```

- [ ] **Step 5: Implementar o comando**

Crie `cmd/yatzee/main.go`:

```go
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
```

Nota: `app.Serve` é da Task 12. Para fechar esta task, crie `internal/app/server.go` com o esqueleto:

```go
package app

import (
	"context"
	"errors"
)

// Serve sobe o servidor SSH. Implementado na Task 12.
func Serve(context.Context) error { return errors.New("modo servidor ainda não implementado") }
```

- [ ] **Step 6: Rodar os testes e jogar de verdade**

```bash
export PATH=$HOME/.local/go/bin:$PATH
go test ./... && go vet ./...
YATZEE_DATA_DIR=/tmp/yatzee-manual go run ./cmd/yatzee
```

Expected: PASS nos testes. No jogo: a tela pergunta o nome; digitar `Afonso` aparece como `afonso`; enter abre o menu; `placar` mostra "nenhuma partida registrada ainda"; `esc` volta; `novo jogo` abre a rodada; `tab` troca pra loja; `esc` volta ao menu; `q` sai. Rodar o comando de novo vai direto pro menu.

- [ ] **Step 7: Escrever o README**

Crie `README.md`:

```markdown
# Yatzee Terminal

Jogo de dados no terminal, com progressão e loja inspiradas em Balatro.

## Estado atual

A estrutura está de pé; as mecânicas do jogo não. O que funciona hoje:

- Identidade: o jogo pergunta seu nome na primeira vez e lembra nas seguintes
- Navegação: menu, placar e a tela de partida, com `tab` alternando rodada e loja
- Servidor SSH, com identidade por chave pública

O que **ainda não existe**: os dados, a pontuação e os itens da loja. As duas
fases da partida são telas vazias com o texto dizendo o que vem ali. O placar
funciona, mas fica vazio até haver partidas para registrar.

## Rodar

Precisa de Go 1.27 ou mais novo.

Os alvos do `make` usam `$(HOME)/.local/go/bin/go`, que é onde o Go está
instalado no WSL deste projeto. Se o seu Go estiver em outro lugar, chame o
comando direto em vez do `make`:

```bash
go run ./cmd/yatzee
```

```bash
make run     # joga local
make serve   # sobe o servidor SSH em 0.0.0.0:2222
make test    # go test ./...
```

Para jogar no servidor local:

```bash
ssh -p 2222 localhost
```

## Configuração

| Variável | Default | Uso |
|---|---|---|
| `YATZEE_SSH_HOST` | `0.0.0.0` | Interface do servidor |
| `YATZEE_SSH_PORT` | `2222` | Porta do servidor |
| `YATZEE_DATA_DIR` | `~/.local/share/yatzee` | Placar, jogadores e chave do host |
| `YATZEE_SSH_HOST_KEY` | `$YATZEE_DATA_DIR/ssh_host_ed25519` | Chave do host, gerada no primeiro boot |

## Identidade

O jogador é reconhecido pelo fingerprint da sua chave pública SSH, não pelo IP:
a chave é estável entre redes e atravessa proxy, o IP não. Na primeira conexão o
jogo pergunta o nome; nas seguintes reconhece. Quem conecta sem chave joga como
convidado e escolhe um nome a cada sessão.

O nome é único e definitivo: minúsculas, números, `-` e `_`, no máximo um
espaço, de 3 a 16 caracteres.

## Teclas

| Tecla | Efeito |
|---|---|
| `↑`/`↓`, `k`/`j` | Move a seleção |
| `enter` | Confirma |
| `esc` | Volta uma tela |
| `tab` | Alterna rodada e loja |
| `q` | Sai (no menu) |
| `ctrl+c` | Sai de qualquer lugar |

## Documentação

- Design: `docs/superpowers/specs/2026-10-13-yatzee-terminal-estrutura-design.md`
- Plano: `docs/superpowers/plans/2026-10-13-estrutura-base.md`
```

- [ ] **Step 8: Commit**

```bash
git add cmd/ internal/app/ README.md
git commit -m "feat: jogo local de ponta a ponta com identidade lembrada

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

### Task 12: Servidor SSH

**Files:**
- Modify: `internal/app/server.go` (substituir o esqueleto da Task 11)
- Test: `internal/app/server_test.go`

**Interfaces:**
- Consumes: tudo das tasks anteriores
- Produces: `app.Serve(ctx context.Context) error`, `app.IdentityFromKey(key gossh.PublicKey) player.Identity`, `app.GuestIdentity(rnd *rand.Rand) player.Identity`

- [ ] **Step 1: Escrever o teste da identidade por chave**

Crie `internal/app/server_test.go`:

```go
package app

import (
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/afonsodesenv/yatzee-terminal/internal/player"
	gossh "golang.org/x/crypto/ssh"
)

// Chave ed25519 fixa, só para o teste ter um fingerprint estável.
const chaveDeTeste = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEnTeDU2pMM3YIoG1dLS1hmVoVVfp9AX8tSCVIxH1xH1 teste@yatzee"

func TestIdentityFromKeyUsaOFingerprint(t *testing.T) {
	key, _, _, _, err := gossh.ParseAuthorizedKey([]byte(chaveDeTeste))
	if err != nil {
		t.Fatalf("parse da chave de teste: %v", err)
	}

	id := IdentityFromKey(key)
	if id.Kind != player.KindPubKey {
		t.Errorf("Kind = %q, quero pubkey", id.Kind)
	}
	if !strings.HasPrefix(id.ID, "key:SHA256:") {
		t.Errorf("ID = %q, quero prefixo key:SHA256:", id.ID)
	}
	if id.Name != "" {
		t.Errorf("Name = %q, quero vazio (o nome vem do store)", id.Name)
	}

	// Mesma chave, mesma identidade: é o que faz a pessoa ser reconhecida.
	if outro := IdentityFromKey(key); outro.ID != id.ID {
		t.Errorf("mesma chave deu IDs diferentes: %q e %q", id.ID, outro.ID)
	}
}

func TestGuestIdentityNaoPersisteENaoRepete(t *testing.T) {
	rnd := rand.New(rand.NewPCG(1, 2))

	a := GuestIdentity(rnd)
	b := GuestIdentity(rnd)

	if a.Kind != player.KindGuest {
		t.Errorf("Kind = %q, quero guest", a.Kind)
	}
	if a.Persistent() {
		t.Error("convidado marcado como persistente")
	}
	if !strings.HasPrefix(a.ID, "guest:") {
		t.Errorf("ID = %q, quero prefixo guest:", a.ID)
	}
	if a.ID == b.ID {
		t.Errorf("dois convidados com o mesmo ID: %q", a.ID)
	}
}
```

- [ ] **Step 2: Rodar o teste e confirmar que falha**

Run: `export PATH=$HOME/.local/go/bin:$PATH && go test ./internal/app/ -run TestIdentityFromKey -v`
Expected: FAIL com `undefined: IdentityFromKey`

- [ ] **Step 3: Implementar o servidor**

Substitua `internal/app/server.go`:

```go
package app

import (
	"context"
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
// estado compartilhado entre jogadores.
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
					wish.Errorln(s, "não consegui abrir o jogo:", err)
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

	go func() {
		<-ctx.Done()
		_ = srv.Close()
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
			// Nao recusa a conexao por falha de leitura — e um jogo, nao um
			// cofre — mas registra. Um players.json corrompido faria todo
			// jogador que volta ver a tela de nome outra vez, e sem este log
			// o operador nao teria sinal nenhum de que isso esta acontecendo.
			log.Printf("ler identidade %s: %v", id.ID, err)
		case ok:
			sess.Identity.Name = p.Name
			_ = players.Touch(ctx, id.ID)
		}
	}

	return NewRoot(sess, FirstScreen(sess)), nil
}
```

Imports de `server.go`: além dos listados no bloco, `"errors"` para o
`errors.Is` do `ErrServerClosed`.

- [ ] **Step 4: Rodar os testes**

Run: `export PATH=$HOME/.local/go/bin:$PATH && go test ./... && go vet ./...`
Expected: PASS em tudo

- [ ] **Step 5: Verificar o servidor de verdade**

```bash
export PATH=$HOME/.local/go/bin:$PATH
export YATZEE_DATA_DIR=/tmp/yatzee-serve
export YATZEE_SSH_HOST=127.0.0.1
export YATZEE_SSH_PORT=2222
go run ./cmd/yatzee serve
```

Noutro terminal do WSL:

```bash
ssh -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null localhost
```

Expected:
1. A chave do host aparece em `/tmp/yatzee-serve/ssh_host_ed25519` com permissão `600`.
2. A primeira conexão pergunta o nome. Digitar maiúscula aparece minúscula.
3. Navegar menu → placar → menu → jogo → `tab` → `esc` → `q` funciona igual ao local.
4. **Reconectar vai direto pro menu**, com o nome lembrado — é a prova da identidade por chave.
5. Conectar de outro terminal ao mesmo tempo abre uma sessão independente: a navegação de um não mexe na do outro.
6. Tentar o nome de outro jogador mostra o erro na tela e aceita outro nome.

Se a tela sair branca, confira o tamanho do terminal: cliente anunciando 0x0 é tratado pelo `ui.Size`, mas um terminal real deve reportar tamanho de verdade.

- [ ] **Step 6: Commit**

```bash
git add internal/app/
git commit -m "feat: servidor ssh com identidade por chave publica

Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>"
```

---

## Verificação final

Rodar depois da última task, antes de declarar a estrutura pronta.

- [ ] `export PATH=$HOME/.local/go/bin:$PATH && go test ./...` — tudo passa
- [ ] `go vet ./...` — limpo
- [ ] `go build -o bin/yatzee ./cmd/yatzee` — compila
- [ ] `grep -rn "math/rand\"" --include=*.go .` — nenhum uso do rand global (só `math/rand/v2`)
- [ ] `grep -rn "github.com/charmbracelet" --include=*.go .` — nenhum import do caminho antigo
- [ ] `grep -rn "internal/app" internal/ui/` — vazio: nenhuma tela importa o pacote app
- [ ] Os nove critérios de aceitação do spec, conferidos à mão no jogo local e por SSH
