# Yatzee Terminal — Jogo Local: Estrutura

**Data:** 2026-10-14
**Status:** Implementado
**Substitui:** `2026-10-13-yatzee-terminal-estrutura-design.md`, que desenhava um
servidor SSH com identidade por chave pública e placar persistido. Essa ideia
foi abandonada: o jogador vai baixar os fontes e rodar no próprio terminal.

## Objetivo

Jogo de dados estilo Yahtzee com progressão e loja inspiradas em Balatro,
rodando no terminal da máquina de quem joga. Esta entrega é só a **estrutura**:
menu, tela de partida e a navegação entre elas. As mecânicas vêm depois.

## Escopo

**Existe hoje:**

- Menu com dois itens: novo jogo e sair
- Tela de partida com duas fases placeholder — rodada e loja — alternadas por `tab`
- Navegação completa: menu → jogo → menu
- Testes de navegação e de layout

**Não existe, e é a próxima conversa:**

- Os dados, a pontuação e as categorias
- Economia e itens da loja

**Deliberadamente fora:** servidor, rede, contas, nome de jogador, placar e
qualquer persistência. O jogo não grava nada em disco.

## Stack

| Peça | Escolha |
|---|---|
| Linguagem | Go 1.27.2 |
| TUI | `charm.land/bubbletea/v2` v2.1.0 |
| Estilo | `charm.land/lipgloss/v2` v2.0.6 |

Duas dependências diretas, e nada além do stdlib fora delas.

## Arquitetura

```
cmd/yatzee ──> internal/app ──> internal/ui ──> internal/ui/{menu,game}
```

- **`internal/ui`** — o contrato de tela, as mensagens de intenção, o tema, o
  mapa de teclas e a normalização de tamanho.
- **`internal/ui/menu`**, **`internal/ui/game`** — as duas telas.
- **`internal/app`** — o model raiz (dono da pilha de telas) e o ponto de entrada.
- **`cmd/yatzee`** — o binário.

### A regra que sustenta tudo: nenhuma tela conhece outra tela

Uma tela emite uma **intenção** (`ui.StartGameMsg`, `ui.BackMsg`, `ui.QuitMsg`)
e o model raiz é o único lugar que conhece o grafo de navegação, traduzindo
intenção em empilhar ou voltar.

```go
type Screen interface {
    Init() tea.Cmd
    Update(tea.Msg) (Screen, tea.Cmd)
    View() string // só a raiz devolve tea.View
    Title() string
}
```

Duas consequências práticas: qualquer tela se testa com nada além de teclas e
inspeção do comando devolvido, e nenhum pacote sob `internal/ui/` importa
`internal/app` — o que seria ciclo, já que `app` importa as telas.

O menu não importa a tela de jogo. É por isso que o vocabulário é de intenção
em vez de uma mensagem genérica carregando a próxima tela dentro: com a tela
dentro da mensagem, o grafo de navegação volta a se espalhar por todas elas.

### Telas

```
Menu ──> Game ──┬── rodada
                └── loja      [tab alterna, esc volta ao menu]
```

Rodada e loja **não** são telas independentes na pilha: são fases de uma
partida. A tela de jogo é dona das duas e escolhe qual exibir pela fase atual,
o que mantém o estado da partida num lugar só quando as mecânicas chegarem.
`round.go` é o arquivo em que os dados vão crescer, `shop.go` o dos itens.

### Teclas

| Tecla | Efeito | Onde |
|---|---|---|
| `↑`/`↓`, `k`/`j` | Move a seleção | Menu |
| `enter` | Confirma | Menu |
| `tab` | Alterna rodada e loja | Jogo |
| `esc` | Volta ao menu | Jogo |
| `q` | Sai | Menu |
| `ctrl+c` | Sai de qualquer lugar | Global |

`esc` no menu não faz nada: é a base da pilha, e sair exige `q` ou `ctrl+c`,
para ninguém encerrar o jogo por reflexo de voltar.

## Três detalhes que parecem supérfluos e não são

Herdados da fase anterior, mantidos porque custam pouco e guardam algo real.
Cada um tem teste que **falha** quando o código que ele guarda é quebrado —
verificado por mutação, não por confiança.

1. **`copyStack` antes de cada mutação da pilha.** `Root` é passado por valor,
   então duas versões do model compartilhariam o mesmo array de trás, e
   empilhar depois de um `esc` sobrescreveria o histórico da versão anterior.
2. **`pushWith` devolve o `Init()` da tela empilhada.** O Bubble Tea chama
   `Init` uma única vez, no model raiz; sem isto, uma tela que carrega algo
   fica presa para sempre no estado inicial. Foi exatamente o defeito que
   deixou a antiga tela de placar em "carregando" para sempre.
3. **`ui.Size` normaliza tamanho zerado para 80x24** e a raiz reemite a
   mensagem corrigida, porque o renderer só se redimensiona com um
   `WindowSizeMsg` que passe pelo laço do Bubble Tea. Com largura zero nada é
   desenhado. O tamanho normalizado é ponto fixo, então não há laço.

## Critérios de aceitação

1. `make run` (ou `go run ./cmd/yatzee`) abre o menu.
2. `novo jogo` abre a tela de partida.
3. `tab` alterna rodada e loja.
4. `esc` volta ao menu.
5. `q` no menu encerra devolvendo o terminal ao normal.
6. `go test ./...` passa; `go vet ./...` limpo; `gofmt -l` vazio.
