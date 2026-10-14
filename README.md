# Yatzee Terminal

Jogo de dados no terminal, com progressão e loja inspiradas em Balatro.

## Rodar

Precisa de Go 1.27 ou mais novo.

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
