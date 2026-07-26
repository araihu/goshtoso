# Auditoria de qualidade de aplicações Goshtoso

Data da auditoria: 2026-07-25

Base auditada: `origin/main` em `bd8edd1c3d9baa188654b93e7f049dd94d414c69`

Status: implementação e validação concluídas em `codex/agent-quality-improvements`

Este documento é o checkpoint permanente da investigação sobre por que agentes
conseguem integrar Goshtoso, mas nem sempre produzem aplicações bonitas e úteis
sem receber o site da biblioteca, Manja ou `tks-console` como referência. Ele
preserva evidências, decisões e critérios de aceite para que retomadas não
reduzam o escopo nem repitam a investigação.

## Veredito executivo

Goshtoso já resolve bem instalação, assets locais, componentes tipados e
interatividade. O principal problema é de transferência: a orientação pública
leva o consumidor até um componente funcionando, enquanto os melhores produtos
possuem uma camada própria de shell, hierarquia, navegação, estados, CSS e
testes que não aparece como contrato reutilizável da biblioteca.

Mais temas ou mais componentes atômicos não corrigem essa lacuna. A ordem
aprovada é:

1. publicar padrões de aplicação e critérios visuais;
2. tornar exemplos completos extraíveis;
3. validar o aprendizado em um módulo externo sem acesso ao source do site;
4. promover para componentes apenas os contratos que continuarem repetidos.

A jornada de adoção recebeu **26/40, Aceitável**. A meta do benchmark após as
melhorias é pelo menos **30/40**, sem source-dives nem utilities que falhem
silenciosamente.

A reavaliação final atingiu **36/40, +10 pontos**, por meio de um app externo
novo construído apenas com a superfície pública permitida. O resultado superou
a meta sem source-dive direto, sem asset remoto, sem erro de console e sem
overflow de página na matriz visual. A pureza de contexto é qualificada porque
o harness injetou memórias de trabalhos anteriores; essa contaminação foi
declarada e nenhum source ou artefato citado por ela foi aberto ou reutilizado.

## Método e fontes

A auditoria combinou cinco linhas independentes de evidência:

- revisão da skill pública `using-goshtoso`, `docs/USAGE.md`, catálogo, demos,
  exemplos e regras internas de design;
- assessment independente de design e descoberta;
- detector de anti-patterns aplicado a 106 arquivos `.templ`;
- sonda cega em um módulo externo chamado Sourceboard;
- inspeção de consumidores reais: Manja e a subtree `tks-console` de
  `guilycst/totvs-work`.

Limite original: não havia backend de navegador disponível durante a auditoria
inicial. A implementação encerrou essa lacuna com inspeção real em 390 px e
1440 px, Goshtoso/Minimal, light/dark e loading/empty/error/success.

## Design Health Score — baseline

| # | Heurística | Nota | Questão principal |
|---|---|---:|---|
| 1 | Visibilidade do status | 3/4 | Bons estados funcionais, mas nenhum checklist público diz quando a página está pronta. |
| 2 | Correspondência com tarefas reais | 3/4 | A linguagem de API é precisa; a navegação não começa por tarefas como operações, detalhe e wizard. |
| 3 | Controle e liberdade | 3/4 | Existem starter, temas e exemplos; falta um caminho claro para adaptar uma composição. |
| 4 | Consistência e padrões | 2/4 | Inventários de temas, packages e exemplos apresentam drift. |
| 5 | Prevenção de erros | 3/4 | Integração é bem protegida; composição e limite do CSS embutido não são. |
| 6 | Reconhecimento, não lembrança | 2/4 | O consumidor cruza skill, catálogo, exemplo, pkg.go.dev e source. |
| 7 | Flexibilidade e eficiência | 2/4 | Bons escape hatches técnicos, mas sem recipes copiáveis ou contratos de página. |
| 8 | Design estético e minimalista | 2/4 | Componentes coerentes, porém a composição pública repete grids e cards de mesmo peso. |
| 9 | Diagnóstico e recuperação | 3/4 | Troubleshooting técnico é bom; classes ausentes e composição fraca não têm diagnóstico equivalente. |
| 10 | Ajuda e documentação | 3/4 | Cobertura ampla no nível de API, pouca orientação para aplicações completas. |
| **Total** |  | **26/40** | **Fundação sólida, transferência insuficiente.** |

## Reavaliação final — 36/40

A sonda externa Control Room aplicou a mesma família de dez heurísticas. O
relatório completo preserva fontes, comandos, screenshots temporários, gates e
ressalvas em `docs/audits/2026-07-25-blind-agent-probe.md`.

| # | Heurística | Baseline | Final | Evidência de mudança |
|---|---|---:|---:|---|
| 1 | Visibilidade do status | 3/4 | 4/4 | State matrix, Skeleton, badges, live region e progresso do workflow. |
| 2 | Correspondência com tarefas reais | 3/4 | 4/4 | Recipes e app operacional usam serviços, owners, SLO, canary e rollback. |
| 3 | Controle e liberdade | 3/4 | 3/4 | Retry, Back/Continue por POST e histórico funcionam; navegação global mobile continua app-owned. |
| 4 | Consistência e padrões | 2/4 | 4/4 | Inventários centralizados, method routes, landmarks e componentes de composição. |
| 5 | Prevenção de erros | 3/4 | 3/4 | Atributos nativos e defaults seguros melhoraram; a sonda não exigia ciclo completo de validação. |
| 6 | Reconhecimento, não lembrança | 2/4 | 4/4 | Skill progressiva, quatro recipes e referências geradas mantêm contratos visíveis. |
| 7 | Flexibilidade e eficiência | 2/4 | 3/4 | Slots, attrs e URLs diretas funcionam; uma UI branded ainda exige CSS próprio relevante. |
| 8 | Design estético e minimalista | 2/4 | 4/4 | Superfície dominante, hierarquia contida, índice editorial e ausência de card soup. |
| 9 | Diagnóstico e recuperação | 3/4 | 4/4 | Empty/error acionáveis, retry, preservação de contexto e matriz visual explícita. |
| 10 | Ajuda e documentação | 3/4 | 3/4 | A orientação foi suficiente sem source; restam sharp edges P2 documentados abaixo. |
| **Total** |  | **26/40** | **36/40** | **Meta de 30/40 superada por 6 pontos.** |

O score final não afirma perfeição. A sonda precisou de 570 linhas de CSS
app-owned para sua identidade, descobriu que `table.Row.Link` combinado com uma
ação Button pode gerar controles interativos aninhados e precisou passar
`MainAttrs` para tornar o alvo do skip link focável. Esses pontos permanecem
como backlog P2, não como falhas ocultas.

## O que já funciona

### Integração operacional

A skill pública cobre módulo Go, `assets.Handler()`, `head.Dependencies()`,
imports, geração templ, estratégia CSS e falhas comuns. Esse contrato permite
uma primeira renderização previsível sem CDN e sem abrir o source da biblioteca.

### Componentes e comportamento

Na base auditada, o catálogo possuía 42 páginas, 74 primitives e 222 seções de
demonstração. Sete apps completos demonstravam HTMX, OOB, SSE, WebSocket,
cookies, IndexedDB, validação e undo. Goshtoso não é um kit visual vazio.

### Direção de design

`DESIGN.md` já registra hierarquia tipográfica, papéis semânticos de cor,
navegação, busca, densidade e anti-patterns. Esse conhecimento está no repositório
do produtor, mas não no caminho normal do consumidor.

## Evidência dos consumidores reais

### Manja

Manja usa Goshtoso, mas sustenta PRODUCT/DESIGN próprios, templates extensos e
aproximadamente 945 linhas de CSS. O produto demonstra que componentes ajudam,
mas a gramática de página ainda é específica do app e invisível ao consumidor
da biblioteca.

### `guilycst/totvs-work/tks-console`

A subtree auditada foi verificada contra o upstream `cloud104/tks-console`:

- monorepo `guilycst/totvs-work` em `8cbd1eaca7bc0c848c613659dec8cc28916fe609`;
- subtree com tree `a7368e9c00bc9f506ca344e3a5787e85be967e65`;
- upstream `cloud104/tks-console` em
  `f079108fd3d0dd6ffbb7dce7bb912c5f9d6f2e00`, com o mesmo tree;
- Goshtoso fixado no commit `06edaa228e71a214ab94885500577c44098a1e88`.

Inventário sem os arquivos Go gerados por templ:

| Superfície | Quantidade |
|---|---:|
| Arquivos `.templ` | 42 |
| Linhas `.templ` | 12.164 |
| Templates de páginas | 26 arquivos, 8.707 linhas |
| Templates de componentes locais | 16 arquivos, 3.457 linhas |
| Atributos `class=` | 1.537 |
| Chamadas renderizadas Goshtoso | 139 |
| Arquivos-fonte que importam Goshtoso | 39 |
| Caminhos Goshtoso distintos | 21 |
| CSS específico do app | 2.161 linhas |
| CSS Goshtoso extraído | 8.260 linhas |
| JavaScript próprio | 418 linhas |
| Testes E2E | 111, além de `TestMain` |

A melhor usabilidade do `tks-console` vem de decisões repetíveis:

- shell persistente com um único scroll container;
- sidebar, topnav, breadcrumbs e busca global com teclado;
- filtros sticky e densos acima de tabelas;
- navegação HTMX com URL, título e OOB sincronizados;
- skeleton, vazio, erro com retry e conteúdo carregado;
- hierarquia orientada a prioridade, não a cards decorativos;
- flows longos com steps, validação server-side e review;
- microcopy específica e testes de jornadas.

O melhor artefato de transferência encontrado foi
`tks-console/docs/plans/addon-install-form.md`: ele define usuário, ação,
fluxo, inventário de componentes, estados, copy e acessibilidade antes do
código. O formato deve ser generalizado nas recipes públicas.

O console também registra atritos reais: runtime montado manualmente,
implementações locais ainda ativas de form, modal, banner e toast, tabs locais
em uma página central e um pipeline CSS próprio. Isso é evidência de fronteiras
que precisam ser documentadas ou graduadas, não prova de que o uso básico já
garante o mesmo resultado.

## Controle: sonda cega Sourceboard

Um agente externo construiu em 9 minutos e 30 segundos um dashboard operacional
com Navbar, métricas, Table, atividade, formulário e estados loading, empty,
error e success. O resultado usou 11 packages Goshtoso, zero CSS autoral e zero
source-dives. Geração, build, testes, rotas, POSTs e assets locais passaram.

A sonda reproduziu dois gaps centrais:

1. oito utilities plausíveis não existiam no CSS embutido:
   `max-w-7xl`, `xl:grid-cols-4`, `lg:col-span-2`, `min-h-64`,
   `sm:text-4xl`, `first:pt-0`, `last:pb-0`, `min-w-[220px]` e
   `sm:col-span-2`;
2. Card não oferecia conteúdo de body arbitrário, forçando o feed principal
   para o slot `Footer`.

O app final funcionou, mas a composição foi enfraquecida depois de consultar o
stylesheet servido. Esse é exatamente o tipo de falha silenciosa que o contrato
público deve impedir.

## Diagnóstico priorizado

### P1. A skill termina onde começa a composição

Depois do primeiro componente, faltam rotas para App Shell, Operations List,
Detail Workspace e Multi-step Workflow. A referência gerada tem 1.777 linhas e
é orientada a declarações, não a tarefas.

### P1. O contrato de CSS permite falha silenciosa

O CSS embutido é confiável para o markup conhecido da biblioteca. Ele não é um
build Tailwind geral para classes arbitrárias passadas via `RootClass` e attrs.
Essa fronteira não estava explícita e não havia uma lista estável de utilities
garantidas para as recipes.

### P1. Exemplos completos não eram extraíveis

O índice mostrava seis dos sete apps, omitindo Wizard. Não expunha objetivo,
complexidade, components, estados, source map, entrypoint nem captura.

### P2. Falta uma camada pequena de composição estável

Manja, `tks-console` e a sonda repetem shell, page header, toolbar, estados
empty/loading/error e regiões de detalhe. Os candidatos aprovados para promoção,
após as recipes, são `AppShell`, `PageHeader`, `Toolbar`, `EmptyState` e
`Skeleton`. Primitives genéricas como Stack e Grid ficam fora até existir mais
evidência.

### P2. Drift reduz a confiança

Na base auditada havia:

- comentários e páginas divergindo entre 13, 15 e 16 temas;
- opção TOTVS sem definição correspondente;
- Wizard ausente de `/examples` apesar de registrado nas outras superfícies;
- `/components/dependencies` derivando o package inexistente
  `components/dependencies`, embora a API pública viva em `components/head`.

## Entregas aprovadas

### Orientação pública

- expandir `using-goshtoso` com a seção “From first component to application”;
- criar `references/application-patterns.md`;
- criar `references/visual-acceptance.md`;
- documentar a fronteira do CSS embutido ao lado dos hooks de classe;
- usar exemplos de `http.ServeMux` com padrões qualificados por método;
- remover orientação exclusiva de maintainers da página consumer-facing.

### Quatro recipes canônicas

1. **App Shell:** topnav, sidebar, skip link, breadcrumbs, uma região rolável e
   comportamento mobile.
2. **Operations List:** header, ação principal, busca, filtros, Table, loading,
   empty e error.
3. **Detail Workspace:** identidade, status, tabs, ações e rail de detalhes.
4. **Multi-step Workflow:** Steps, Form, validação server-side, review e envio.

Cada recipe deve declarar problema, estrutura DOM, components, estados,
responsividade, acessibilidade, CSS exigido, mapa de arquivos, source link e
critérios de conclusão.

### Exemplos e benchmark

- transformar `/examples` em inventário extraível e incluir Wizard;
- fornecer um módulo externo autocontido que não importe `site/`;
- verificar que as classes exclusivas do app são emitidas ou deliberadamente
  pertencem ao build Tailwind do consumidor;
- rerodar o mesmo tipo de sonda sem acesso ao source de Goshtoso, Manja ou
  `tks-console`.

### Componentes candidatos

Promover apenas após as recipes comprovarem repetição:

- `AppShell`;
- `PageHeader`;
- `Toolbar`;
- `EmptyState`;
- `Skeleton`;
- body arbitrário em Card ou uma alternativa de Panel, se a repetição persistir.

## Critérios de aceite do benchmark

- módulo Go externo novo, sem import de `site/`;
- zero source-dives na biblioteca durante a tarefa;
- build, testes e assets verdes;
- layout útil em 390 px e 1440 px;
- light, dark e tema Minimal;
- loading, empty, error e success;
- nenhum erro de console;
- nenhum achado P1 de acessibilidade;
- score de design de pelo menos 30/40;
- nenhuma utility necessária ausente silenciosamente.

## Ledger de implementação

Este ledger deve ser atualizado quando cada fase terminar. “Feito” exige teste,
não apenas arquivos presentes.

| Entrega | Estado inicial | Evidência de conclusão |
|---|---|---|
| Relatório versionado | feito | este arquivo |
| Roteamento progressivo na skill | feito | `TestAgentSkillRoutesFromIntegrationToApplicationPatterns` |
| Application patterns reference | feito | quatro patterns e state matrix testados |
| Visual acceptance reference | feito | matriz de viewport/tema/modo e checks automatizados testados |
| Contrato de CSS embutido | feito | nove utilities safelisted e presentes em `assets/styles.css` |
| Drift de temas/TOTVS/package | feito | 15 opções contra CSS, TOTVS removido e Dependencies → `head` |
| Índice extraível com sete apps | feito | sequência editorial com imagens, Wizard, complexity, components, states e source links testados |
| Quatro recipes canônicas | feito | `/docs/application-patterns` com previews, contratos 390/1440, source maps, SEO, busca e E2E |
| Benchmark externo | feito | módulo integrado; build/testes verdes; matriz 390/1440, temas, modos e quatro estados sem overflow ou erros de console |
| Componentes promovidos | feito | AppShell, PageHeader, Toolbar, EmptyState, Skeleton e Card Body; APIs, demos, catálogo, skillgen e E2E completo |
| Sonda cega final | feito | Control Room externo, zero source-dive direto, gates verdes e 36/40; qualificação de memória declarada |
| Reavaliação final | feito | 36/40 contra 26/40; aumento de 10 pontos com a mesma régua |

## Tarefas paralelas sob o control plane

Cada tarefa opera em um worktree próprio e entrega um commit para revisão e
integração nesta branch. A sessão principal continua responsável por conflitos,
regeneração, testes finais e autoria do resultado consolidado.

| Tarefa | Thread | Escopo | Estado |
|---|---|---|---|
| Benchmark externo | `019f9b9b-334f-7430-a575-b5e926c7566c` | `examples/application-patterns` autocontido | integrado, revisado e arquivado |
| Recipes públicas | `019f9b9b-3351-7180-bba6-4ab77609fcb1` | `/docs/application-patterns`, preview, SEO e testes | integrado, revisado e arquivado |
| Kit de composição | `019f9b9b-3351-7180-bba6-4ad8292435a0` | AppShell, PageHeader, Toolbar, EmptyState, Skeleton e Card Body | integrado, revisado e arquivado |
| Sonda cega Control Room | `019f9bc4-aefb-7692-9747-cc2789f81fcf` | app temporário externo e relatório de aceitação | integrado, revisado e arquivado |

## Snags já registrados

- `RootClass` e attrs aceitam nomes de classes que podem não existir no CSS
  embutido. Isso precisa ser tratado como contrato, não como surpresa do build.
- O site é um módulo Go separado; verificações nele devem usar o workspace local
  ou `GOWORK=off` conforme o tipo de teste.
- Arquivos `*_templ.go`, `assets/styles.css` e
  `assets/goshtoso-theme.css` são gerados e não podem ser editados à mão.
- Mudanças de API exigem `go run ./cmd/skillgen` além de `templ generate`.
- A auditoria original não teve navegador disponível; não declarar qualidade
  visual final sem screenshots e inspeção real.
- A primeira verificação `GOWORK=off` encontrou o starter já quebrado: ele usava
  `table.Config.PaginationID()`, mas a implementação mantinha `paginationID()`
  privado apesar do comentário público. O contrato de IDs de fragmento passou a
  exigir `TbodyID`, `TheadID` e `PaginationID` exportados e testados.
- O benchmark externo precisou abrir o source para confirmar
  `table.Cell.Component`, `table.LinkMode`, seletores de tema e a importação
  injetada pelo templ. A referência pública agora contém os quatro contratos.
- `button.Button` não aceitava `name`/`value` nativos, o que impedia um
  formulário multi-ação idiomático. `button.WithAttrs` agora cobre atributos
  nativos e o benchmark prova Back/Continue com POSTs reais.
- Na recipe pública, `tabs.Config.ID` identificava os painéis ARIA, não um root
  de DOM contratual; o E2E passou a escopar pelo preview, sem acoplar-se a uma
  estrutura privada. A tabela também precisou de largura mínima explícita para
  cumprir a promessa de rolagem horizontal em 390 px.
- Após integrar o kit, a recipe e a demo AppShell colidiram no helper templ
  privado `appShellCreateAction`. Os helpers da recipe receberam prefixo
  `applicationPattern` e o build integrado ganhou uma regressão real que os
  slices isolados não podiam detectar.
- A matriz final em 390 px encontrou overflow de página tanto na tabela de
  success quanto no skeleton de loading do benchmark. Grid e flex ancestors
  passaram a usar `min-width: 0`, a track usa `minmax(0, 1fr)` e o loading
  empilha no breakpoint mobile; o teste de assets protege a contenção.
- A sonda final mostrou que `table.Row.Link` com `table.Row.Actions` pode
  produzir controles interativos aninhados. Até haver prevenção no componente
  ou orientação específica, apps devem escolher linha clicável ou ação Button,
  não ambas.
- `AppShell` permite tornar o alvo do skip link focável via `MainAttrs`, mas não
  aplica `tabindex="-1"` automaticamente. Alinhar o default ao contrato público
  é uma oportunidade P2.
- O harness obrigou a sonda a receber memórias de resultados anteriores. Ela
  manteve zero source-dive direto e não reutilizou artefatos, mas o relatório
  qualifica corretamente a alegação de cegueira absoluta.

## Decisão de encerramento

O trabalho está encerrado porque o benchmark externo satisfez os critérios, a
sonda independente superou a meta, o ledger foi atualizado e os gates finais
passaram:

- `templ generate`, `just css` e `go run ./cmd/skillgen` sem drift;
- `go fix ./...` nos dois módulos sem alterações;
- root e site com `golangci-lint run`: zero issues;
- build do site e build/vet/lint do benchmark externo;
- testes de root, site, starter e benchmark;
- E2E completo: `ok github.com/araihu/goshtoso/site/tests/e2e` em 308,731 s;
- inspeção real de recipes, componentes e benchmark, sem erros de console;
- sonda cega final: 36/40 e zero source-dive direto.

Os três follow-ups P2 preservados são reduzir o CSS app-owned necessário para
branding, prevenir/documentar linha clicável com ação interativa e considerar
`tabindex="-1"` como default do alvo principal de AppShell. Nenhum deles reabre
os P1s que motivaram esta auditoria.
