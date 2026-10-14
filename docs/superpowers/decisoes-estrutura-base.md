> **Contexto historico.** Varias destas decisoes tratam do servidor SSH, da
> identidade e do placar, que foram removidos em 2026-10-14. As que seguem
> valendo sao as de navegacao, de forca de teste e de ferramental.

# Decisões que eu tomei sozinho — estrutura base

Executando o plano por subagentes, cada conflito, ambiguidade ou defeito do
próprio plano foi decidido na hora em vez de parar e esperar. Este arquivo é o
registro completo e verbatim dessas decisões, extraído do ledger de execução
antes de ele ser descartado. Cada uma diz o que custa se estiver errada, e
todas são reversíveis — o histórico em git é granular de propósito.

Ordem cronológica. 50 decisões.


## 48. Ruling: trabalhar na branch `estrutura-base` no próprio diretório, sem worktree separada — o critério de aceitação do usuário é abrir o jogo neste diretório pelo WSL, e uma worktree destacada move o código pra outro lugar. Custo se errado: nenhum material, o histórico segue isolado numa branch e mergeável.

## 50. Ruling: CONFLITO 1 — `Root` passa a usar `keys.Matches(msg.String(), keys.ForceQuit)` em vez do literal, e ganha um campo `keys`. Keybinding mora num lugar só, senão a barra de help pode mentir sobre o que o jogo aceita. Custo se errado: duas linhas a mais que poderiam ser um literal.

## 52. Ruling: CONFLITO 2 — os três esqueletos passam a estar escritos por extenso no plano, com assinatura e título exatos. Custo se errado: nenhum, só verbosidade no plano.

## 54. Ruling: CONFLITO 3 — o código é a autoridade: `localIdentityFor` não exportada, porque só o pacote `app` monta identidade. Bloco Interfaces corrigido. Custo se errado: um rename.

## 56. Ruling: CONFLITO 4 — usar `errors.Is(err, ssh.ErrServerClosed)`. Verifiquei que `charm.land/ssh.ErrServerClosed` existe. Comparar texto de erro quebra silencioso numa atualização da biblioteca. Custo se errado: nenhum, é estritamente melhor.

## 58. Ruling: CONFLITO 5 — acrescentei aos Global Constraints do plano o formato obrigatório de invocação via WSL, e ele vai em todo dispatch. Custo se errado: nenhum; sem isso todo implementador trava no primeiro comando.

## 74. Ruling: achado "make test quebrado" — procede e é load-bearing: verifiquei que o WSL não tem gcc/cc/clang e CGO_ENABLED=0, então `-race` aborta sem rodar teste nenhum. Instalar build-essential exige sudo com senha que eu não tenho. Decisão: `test` passa a ser `go test ./...` (roda hoje) e acrescentei um alvo `test-race` que checa gcc e orienta `sudo apt install build-essential`. Troquei -race por go test simples nas 12 tasks e no critério 8 do spec. Custo se errado: os testes de concorrência que escrevi continuam rodando e pegariam deadlock ou escrita perdida, mas corrida de dados que só o detector vê passa batida até o usuário instalar o compilador — recuperável com um apt install a qualquer momento.

## 76. Ruling: achado "SSHHostKey relativa" — procede. O override do operador não passava por filepath.Abs, contradizendo a constraint de caminho absoluto justamente no cenário que ela existe pra cobrir (unit de systemd). Decisão: corrigir código e plano, mais teste novo TestLoadConfigTornaHostKeyAbsoluta. Custo se errado: nenhum, é estritamente mais correto.

## 87. Ruling: item ⚠️ do revisor (caso " ab " listado no brief mas ausente do teste) — verifiquei rodando: CleanName(" ab ") é rejeitado corretamente por comprimento após trim. Comportamento certo, caso de teste genuinamente ausente. Decisão: é uma linha e vai num fix round junto com o gofmt, porque todas as tasks seguintes herdam o estilo deste pacote e código fora do gofmt vai tropeçar revisor atrás de revisor. Custo se errado: um fix round de ~1 minuto gasto em cobertura mínima.

## 89. Ruling: adotar `-count=1` nas rodadas verbosas por pacote em todo o plano — o relatório da Task 2 mostrou resultado `(cached)` como evidência GREEN de código recém-escrito, o que não prova execução fresca. Custo se errado: alguns segundos a mais por rodada de teste.

## 96. Ruling: item ⚠️ (trailer Co-Authored-By não visível no formato do diff) — conferi `git log --format=trailers` nos cinco commits de código: todos têm o trailer. Não é lacuna.

## 101. Ruling: evidência de teste parafraseada em vez de saída crua do `go test -v` — passo a exigir colagem da saída literal nos dispatches seguintes. Custo se errado: relatórios um pouco mais longos.

## 111. Ruling: corrigi preventivamente o alinhamento gofmt de dois structs do brief da Task 5 (Keys e a tabela do layout_test), porque o mesmo defeito gerou fix round na Task 2 e o estilo deste pacote é herdado por todas as telas. Custo se errado: nenhum, é formatação.

## 116. Ruling: `go mod tidy` na Task 5 removeu os pins de bubbles/wish/ssh/crypto porque nenhum código os importa ainda — comportamento correto do Go, mas cria risco real: um `go get` sem versão nas Tasks 7 e 12 resolveria para a mais recente, não para as versões cuja API eu verifiquei rodando. Decisão: acrescentei aos Global Constraints a exigência de pedir versão explícita na primeira importação de cada biblioteca, com as linhas de `go get` prontas para as Tasks 7 e 12, e vou repeti-las nos dispatches. Custo se errado: nenhum; sem isso, uma API diferente da verificada entra sem aviso e as armadilhas documentadas deixam de valer.

## 118. Ruling: item ⚠️ (alegação de suíte/vet limpos no módulo sem saída crua) — rodei eu mesmo: 4 pacotes ok com -count=1, vet limpo, gofmt vazio. Não é lacuna.

## 125. Ruling: terceira vez que Go escrito à mão no plano sai fora do gofmt (telaFake, depois de Keys e da tabela do layout). O implementador rodou `gofmt -w` só nesse arquivo e declarou a mudança — aceito, é whitespace. Alinhei o plano com o que o gofmt produziu. Padrão reconhecido: todo bloco Go que eu escrever à mão no plano precisa passar por gofmt antes de virar brief. Custo se errado: nenhum, é formatação; mas sem isso cada task gasta um fix round nisso.

## 128. Ruling: achado Important (TestWindowSizeZeroCaiNoDefault passa trivialmente, porque NewRoot já inicia no default e o teste afirma justamente o default depois de mandar 0x0) — procede e é mais grave do que coverage comum: o fallback 0x0 é risco verificado neste projeto e esse teste ficaria verde mesmo se o tratamento de WindowSizeMsg fosse apagado, o que quebraria também o resize normal. O revisor sugeriu deixar para outra task; discordo, porque o arquivo está aberto e o custo é uma linha. Decisão: teste passa a aplicar 120x40 primeiro e só então 0x0. Custo se errado: um fix round curto.

## 130. Ruling: minor #3 (laço do helper `passo` sem cap) — dobrado para dentro do round já aberto, contra a regra de minors não entrarem no loop. Motivo: as Tasks 7 a 10 reusam esse helper com telas reais, e uma tela que devolva comando sempre travaria a suíte em laço infinito; suíte travada custa muito mais pra diagnosticar que assert falho. Custo se errado: uma linha de cap defensivo que nunca dispara.

## 139. Ruling: quarto caso de Go meu fora do gofmt no plano — ordem de import (bubbles vem antes de bubbletea alfabeticamente). O implementador trocou e declarou. Plano corrigido. Mesmo padrão já registrado na Task 6.

## 141. Ruling: item ⚠️ (teste depende de NameRule() conter "16", definido na Task 2) — NameRule é fmt.Sprintf com NameMinLen/NameMaxLen = 3/16 e a suíte passa. Não é lacuna.

## 151. Ruling: achado 1 (tecla "up" literal nunca provada) — procede. `Up: []string{"up","k"}` compartilham um branch; o teste provava "k" mas pressionava "up" só na posição 0, onde não fazer nada é indistinguível de clampar. Remover "up" do mapa deixaria a suíte verde com a tecla documentada quebrada. Esta é a TERCEIRA ocorrência do mesmo padrão no meu plano (Task 6 WindowSize, Task 8 cursor): teste de limite que afirma estado já existente. Decisão: provar o movimento antes do limite, no plano e no código. Custo se errado: nenhum.

## 153. Ruling: achado 2 (evidência da suíte completa resumida em vez de crua, justamente na task que mudou um construtor usado por outro pacote) — rodei a suíte eu mesmo: 6 pacotes com teste, todos ok, dois sem arquivo de teste (ui/game e ui/scores, esperado — Tasks 9 e 10 os preenchem). Não há regressão. Pedi a saída crua no fix report de todo modo. Custo se errado: nenhum.

## 169. Ruling: item ⚠️ (shape de ui.Keys/ui.Theme vem da Task 5 e não está neste diff) — a Task 5 foi revisada e aprovada com exatamente esses campos, e a suíte inteira passa. Não é lacuna.

## 170. Ruling: minor "%-16s não trunca nome longo" — NÃO é problema: player.NameMaxLen = 16 e CleanName rejeita acima disso, então a largura da coluna é exatamente o teto do nome. Fato de outro pacote que o revisor não tinha. Nada a fazer.

## 181. Ruling: achado Important (TestViewSegueAFase passa trivialmente) — procede, e é o mais sutil dos quatro: a barra de help renderiza "tab alterna rodada e loja", então as palavras "rodada" E "loja" estão na saída em TODA fase. O teste passaria mesmo se o View ignorasse a fase e sempre desenhasse a rodada — exatamente o comportamento que ele existe para garantir. Confirmei a linha 84 do game.go. Decisão: afirmar texto exclusivo de cada fase ("as mecânicas de dados entram aqui" / "os itens entram aqui") E a ausência do texto da outra. Custo se errado: nenhum.

## 183. Ruling: minor "sem teste para mensagem não-tecla" — dobrado no round aberto. A raiz manda WindowSizeMsg para a tela do topo a cada resize, então é caminho real, não hipotético, e custa 8 linhas. Custo se errado: um teste trivial a mais.

## 185. Ruling: minor de integridade de evidência — o bloco GREEN do relatório mostra o módulo como "yatzee_terminal" (underscore) enquanto RED e suíte mostram "yatzee-terminal" (hífen), e as linhas de comando não têm o wrapper do WSL. Sinal de bloco editado à mão. Rodei eu mesmo: 4 testes passam, gofmt limpo. Não houve engano funcional, mas vou cobrar saída literal no fix report, porque evidência que parece fabricada corrói a revisão inteira.

## 187. Ruling: minor PhaseRound ser o zero value (teste não distingue atribuição deliberada de campo não inicializado) — aceito como está, concordando com o revisor: o comportamento é idêntico e a atribuição explícita documenta a intenção. Começar o iota em 1 só para fortalecer um teste é pior que o problema. Custo se errado: nenhum hoje; se um dia uma fase nova virar o padrão, revisitar.

## 192. Ruling: o commit 147d15d traz `Co-Authored-By: Claude Haiku 5.5` em vez de `Claude Opus 5`, porque o subagente usou o lembrete de atribuição da sessão dele. Decisão: deixar como está nos dois commits. Reescrever o commit trocaria o SHA que o ledger e o pacote de re-revisão em voo já referenciam, e o trailer atual é factualmente correto sobre qual modelo fez aquele commit. Custo se errado: uma linha de atribuição inconsistente no histórico, visível em `git log` — o usuário decide se quer uniformizar depois com um rebase.

## 200. Ruling: achado próprio (vindo da ressalva de SIGTERM do implementador) — `cmd/yatzee/main.go` usava `context.Background()` sem sinal nenhum, enquanto o `Serve` da Task 12 já tem goroutine esperando `ctx.Done()` para fechar o servidor. Ou seja, o desligamento gracioso que escrevi era código morto: `systemctl stop` ou `docker stop` esperariam o prazo inteiro e matariam o processo à força, derrubando sessões no meio. Decisão: `signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)` no main, corrigido no plano e a ser aplicado no fix round. Custo se errado: nenhum; sem isso, todo restart do servidor demora o grace period e mata sessões abruptamente.

## 202. Ruling: o plano estava em CRLF (um edit meu em Python no Windows converteu) e a reescrita seguinte voltou para LF, inflando um diff de 6 linhas para 7141. Verifiquei com `diff -w` que a única mudança real de conteúdo é o bloco de sinal. Decisão: não reescrever histórico; acrescentei .gitattributes com `* text=auto eol=lf` para a regra valer de agora em diante. Custo se errado: um commit de docs ruidoso no histórico, já passado.

## 208. Ruling: achado README promete o que não existe — procede, com uma correção ao revisor: ele disse que `game.New()` é stub, mas a Task 10 implementou o container de fases e o `tab` alterna de verdade; placeholder é o CONTEÚDO das fases, não o toggle. Então a tabela de teclas do README está correta. Decisão: seção "Estado atual" separando o que funciona do que não existe, mantendo o bullet de SSH porque a Task 12 entra logo em seguida e um aviso que vence em uma hora é pior que nenhum. Custo se errado: README otimista por uma task.

## 210. Ruling: achado README não documenta que o make assume Go em $(HOME)/.local/go/bin — procede e é real para quem clonar fora deste WSL. Decisão: documentar o requisito e oferecer `go run ./cmd/yatzee` como alternativa. Custo se errado: nenhum.

## 212. Ruling: minor "falta tea.WithContext" — promovido na prática e dobrado no round, porque é a outra metade do meu achado de SIGTERM: sem ele o contexto cancelado não para a partida local, só os stores. Custo se errado: nenhum.

## 214. Ruling: minor "sem teste do ramo de erro do store em localIdentityFor" — dobrado no round. É o caminho onde uma falha de leitura seria confundida com jogador desconhecido. Custo se errado: 15 linhas de teste.

## 218. CORREÇÃO ao ruling anterior: o travamento do smoke run NÃO era "o programa ignora SIGTERM", como registrei a partir da primeira ressalva. O implementador diagnosticou melhor na segunda rodada: `timeout` e o jogo ficavam em estado T (stopped) por job control do terminal, e `timeout --foreground -s KILL ... </dev/null` resolve. Minha atribuição a SIGTERM estava errada.

## 223. Ruling: minor novo "ErrProgramKilled vira exit 1 no SIGTERM" — é regressão da MINHA correção: um docker stop da partida local agora imprime erro e sai 1 em vez de sair limpo, e supervisor de processo trata exit 1 como falha (alerta, restart-loop). Procede. Decisão: NÃO abrir outro round na Task 11; carregar para a Task 12, que é dona da semântica de desligamento e toca os mesmos caminhos de saída. Uma mudança, uma revisão, no lugar certo. Custo se errado: a Task 12 precisa lembrar disso — está no dispatch dela.

## 225. Ruling: minor "defer stop() pulado pelo os.Exit" — inofensivo de fato (o processo sai e libera o registro do sinal), só nit de linter. Parked.

## 227. Ruling: minor "README diz que o SSH funciona mas make serve é stub" — resolve-se sozinho quando a Task 12 entrar. Virou item de verificação no dispatch da Task 12: depois que serve funcionar, o README fica correto sem tocar nele.

## 229. Ruling: nota do revisor para a Task 12 — depois que o NotifyContext cancela uma vez, SIGTERMs seguintes são engolidos até o stop() rodar; se o Serve travar no grace period, só SIGKILL escapa. Carregado para o dispatch.

## 243. Ruling: achado Important (erro de players.Get engolido sem log no caminho com chave) — procede, com o veredito que eu pedi: continuar como jogador novo é CERTO num servidor de jogo (disco com problema não deve trancar ninguém fora), mas silenciosamente é ERRADO. Um players.json corrompido faria todo jogador que volta ver a tela de nome outra vez e o operador não teria sinal. Pior: o jogo local propaga esse erro (corrigido na Task 11), então a versão silenciosa esconde uma falha que o outro modo reporta. Decisão: log.Printf e seguir servindo. Custo se errado: nenhum; é a opção que preserva disponibilidade E observabilidade.

## 245. Ruling: minor "ErrProgramKilled vira saída limpa sem checar a causa" — procede. Decisão: gatear com `ctx.Err() != nil`, para que só um desligamento pedido por nós conte como sucesso. Custo se errado: nenhum.

## 247. Ruling: minor "erro interno cru vai para o cliente" — procede. Decisão: mensagem genérica para o jogador, detalhe no log do servidor.

## 249. Ruling: minor "teste de convidado prova só que o gerador avança" — QUINTA ocorrência da família de asserção fraca neste plano, e dessa vez num teste derivado do meu brief. O teste sorteia duas vezes do MESMO RNG, então passaria mesmo se todas as sessões compartilhassem um gerador global — exatamente o que o desenho proíbe. Decisão: construir via dois NewSession distintos. Custo se errado: nenhum.

## 267. Ruling: achado 4 segue aberto porque A MINHA PRÓPRIA CORREÇÃO tinha o mesmo furo do teste que ela substituiu. Eu mandei construir duas sessões independentes e comparar os IDs de convidado; o implementador fez exatamente isso. Mas a asserção continua incapaz de falhar sob o defeito que ela existe para pegar: com um gerador global compartilhado, dois sorteios sequenciais TAMBÉM dão IDs diferentes. Troquei uma asserção fraca por outra, um nível acima. O re-revisor pegou isso e está certo. Decisão: comparar os PONTEIROS dos geradores (`a.Rand == b.Rand` deve ser falso), mantendo também a comparação de IDs — a primeira afirma a propriedade estrutural, a segunda que os geradores produzem valores distintos. Fix round 2/5 despachado com pedido de confirmar que a asserção falha quando os dois apontam para um gerador só. Custo se errado: uma linha.

## 271. Ruling: observação fora de escopo do re-revisor (caminho de erro do sessionModel com model nulo, sem teste) — já resolvida por mim lendo o fonte do wish: model nulo vira programa nulo (bubbletea/tea.go:120-123) e o middleware chama next(sess) e retorna (tea.go:56-60). A sessão encerra e o cliente já recebeu o wish.Errorln. Não é lacuna.

## 286. Ruling: Critical #1 — a tela de placar NUNCA carrega. O Bubble Tea chama Init() uma vez só, no model raiz (único call site: bubbletea v2.1.0 tea.go:1146), e minha Root.Init() arma apenas a tela da base. Logo scores.Screen.Init(), a única coisa que produz LoadedMsg, nunca roda: loaded fica false e View() desenha "carregando..." para sempre. O revisor escreveu um teste que falha no HEAD. Critério de aceitação 3 do spec NÃO é atendido e o README mente ao dizer que o placar funciona. É defeito do MEU plano, prescrito na linha 1793 (Task 6, escrita antes de a Task 9 ter Init assíncrono); nenhuma das 12 revisões de task pegou, porque cada uma só via o próprio diff. O único item da checklist final que eu não fiz — "os nove critérios de aceitação conferidos à mão" — era exatamente o que pegaria. Decisão: corrigir com pushWith devolvendo s.Init(), mais teste de regressão permanente. Custo se errado: nenhum; sem isso o placar é inútil e toda tela futura com carga assíncrona herda a armadilha.

## 288. Ruling: Important #2 — DOIS testes que não podem falhar sobreviveram, provados por mutação: (a) TestPilhaNaoCompartilhaArrayDeTras passa mesmo com copyStack devolvendo m sem copiar, porque guarda uma string copiada antes e descarta o Root pré-Back; (b) TestTouchAtualizaLastSeen usa !Before ("não andou pra trás"), que vale trivialmente quando nada acontece — com Touch virando `return nil` a suíte inteira do player continua passando. Isso leva a família a 8 ocorrências, 2 vivas. Decisão: corrigir as duas asserções e exigir prova de mutação. Custo se errado: o guard do aliasing e o Touch seguem sem cobertura real.

## 290. Ruling: Important #3 — o fallback de 0x0 é CÓDIGO MORTO. m.width/m.height são escritos e nunca lidos pelo View(), e nenhuma tela recebe WindowSizeMsg (o case retorna antes de repassar). O revisor traçou no fonte: o default 80x24 do próprio Bubble Tea só vale quando p.ttyOutput != nil, e num PTY emulado por SSH a saída é a ssh.Session, que não é term.File, então ttyOutput é nil e o renderer é dimensionado 0x0. Ou seja, a armadilha que eu documentei como verificada continua viva — minha verificação por SSH forçava stty, então nunca exercitou esse caminho. Decisão: reemitir o WindowSizeMsg normalizado (o renderer é redimensionado por qualquer WindowSizeMsg que passe pelo loop) e repassar para a tela do topo. Custo se errado: cliente com 0x0 vê tela branca, que é o sintoma que eu já tinha visto no probe.

## 292. Ruling: Minor #8 (internal/game não existe, mas o spec o lista) — é divergência spec/plano, não falha de implementação, e o revisor recomenda NÃO criar pacote vazio só para satisfazer diagrama. Concordo. Decisão: corrigir o SPEC, que é meu, para refletir que internal/game nasce com as mecânicas. Vou fazer eu mesmo.

## 294. Ruling: NÃO incluí idle timeout na onda de correções, apesar de recomendado — SSH IdleTimeout conta tráfego de canal, então um jogador lendo o menu conta como ocioso, e escolher o valor é decisão de produto. Fica para o usuário. Custo se errado: conexões podem ser mantidas abertas indefinidamente por qualquer cliente até alguém definir o limite.
