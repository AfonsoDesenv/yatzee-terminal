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
make test-race   # go test -race ./... (precisa de compilador C: sudo apt install build-essential)
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

## Operação

- `YATZEE_DATA_DIR` precisa ser **persistente**. A chave do host SSH mora ali:
  se ela se perder, é regenerada, e todo jogador que volta recebe
  `REMOTE HOST IDENTIFICATION HAS CHANGED` e é recusado pelo próprio cliente SSH.
- Rode **um processo por `YATZEE_DATA_DIR`**. O mutex do store é por processo:
  jogar local enquanto o servidor roda sobre o mesmo diretório perde escritas em
  silêncio, porque os dois carregam, os dois modificam e o último rename vence.

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
