# Auditoria de `head.Dependencies`: CDN, fallback e opções funcionais

Data: 2026-07-26

Base: `origin/main` em `196b3ff517bcc9d6644caf7c429764d7daef15e9`

Branch: `codex/head-dependency-options`

Status: implementação, revisão independente e validação concluídas; pronta para
release.

## Objetivo

Tornar `head.Dependencies()` um default forte para consumidores humanos e
agênticos: todos os runtimes exigidos pelos componentes devem funcionar sem
wiring manual, URLs externas devem ser versionadas, uma falha de download da
CDN deve usar o asset local correspondente, e aplicações com requisitos
próprios precisam de escape hatches explícitos.

## Achados

### 1. O conjunto completo não era completo

`components/textinput` e exemplos públicos emitem `x-mask`, mas
`head.Dependencies()` não carregava Alpine Mask. O site mascarava a lacuna ao
manter tags manuais na sua própria layout. Um consumidor que seguisse apenas a
API e a documentação recebia um campo com atributo de máscara sem o runtime que
o implementa.

Decisão: Alpine Mask passa a fazer parte do conjunto completo. O conjunto
minimal permanece CSS, Alpine core, HTMX e combobox.

### 2. Trocar `src` no elemento que falhou não é fallback

O primeiro protótipo escutava `error` e mudava o `src` do mesmo `<script>`.
Chromium atualizava o DOM, mas não iniciava uma segunda execução: o elemento já
tinha uma requisição encerrada. A inspeção visual do markup não revelaria a
falha.

Decisão: o fallback sempre cria um novo elemento `<script>`.

### 3. Scripts dinâmicos exigem uma ordem explícita

Alpine Collapse, Focus e Mask devem registrar seus listeners antes de Alpine
core emitir `alpine:init`. Elementos inseridos dinamicamente também não herdam
automaticamente a semântica de ordem de scripts `defer`. Além disso, HTMX 2.0.8
usa `DOMContentLoaded` quando o documento ainda não está completo; uma carga
dinâmica tardia no estado `interactive` pode perder esse evento.

Decisão: um único loader externo carrega sequencialmente Collapse, Focus, Mask,
Alpine, HTMX e combobox. HTMX aguarda `window.load` quando inserido pelo loader,
garantindo que seu bootstrap use o caminho de documento completo.

### 4. A demo anterior não exercitava a API

`/components/dependencies` documentava `@head.Dependencies()`, mas a layout do
site já carregava seu próprio runtime local. O smoke test provava que a página
funcionava, não que o helper documentado funcionava nem que o fallback era real.

Decisão: a cobertura nova usa um documento independente renderizado com a API
pública. O servidor de teste devolve 503 para todas as cinco URLs primárias e o
navegador precisa provar Collapse, Focus, Mask, Alpine, HTMX e navegação de
combobox usando os fallbacks embedded.

## Contrato implementado

O default continua curto:

```templ
@head.Dependencies()
```

Ele monta CSS e loader locais, tenta URLs exatas do unpkg e cai para os mesmos
arquivos versionados servidos por `assets.Handler()`. O estado público
`window.goshtosoDependencies.ready` permite sincronizar JavaScript da aplicação.
O loader emite eventos de fallback, sucesso completo e erro terminal.

As opções funcionais são:

- `WithLocalRuntime()` para zero requisições CDN;
- `WithDependencyCDNURL` e `WithDependencyLocalURL` por runtime;
- `WithDependencyIntegrity` para preservar ou substituir SHA-384 SRI;
- `WithoutLocalFallback()`;
- `WithoutDependency()` quando a aplicação é dona do runtime;
- `WithStylesheetURL`, `WithComboboxURL` e `WithLoaderURL` para mounts próprios.

Hashes SHA-384 são gerados dos mesmos bytes embedded e aplicados tanto à CDN
quanto ao fallback. `templ.WithNonce` é propagado ao loader e a cada script que
ele cria. Não há JavaScript inline nem event handler inline. O runtime Alpine
padrão ainda exige as permissões de CSP que seu modelo de avaliação e mutação
de estilo necessita.

## Evidência preservada

- testes unitários travam URLs CDN e fallback, ordem, conjunto minimal, zero
  value, opções, modo local e nonce;
- o gerador publica constantes locais e CDN a partir do mesmo
  `assets/js/runtime/versions.json`;
- o teste Playwright `TestDependenciesCDNFailureLoadsOrderedLocalFallback`
  provoca cinco falhas primárias e verifica os seis comportamentos de runtime;
- a aplicação de referência usa o default CDN-first e seu teste trava tanto as
  URLs primárias quanto os fallbacks; ausência de CDN não é tratada como mérito.

## Snags registrados

- presença de um `src` local no DOM não prova que o browser executou fallback;
- uma demo que carrega runtime por outro caminho pode dar falsa confiança à API;
- componentes podem emitir diretivas opcionais que o head helper esquece de
  carregar;
- carga dinâmica e `defer` não têm contratos de ordem equivalentes;
- um `x-mask` fora de uma raiz Alpine (`x-data`) não inicializa, mesmo com o
  plugin presente;
- trocar a versão da CDN sem trocar a URL local pode criar fallback incompatível;
- uma falha de rede recuperada ainda pode aparecer no console de rede do
  navegador; o contrato elimina erros JavaScript não tratados, não oculta o
  diagnóstico HTTP.

## Gates

Resultados finais antes do release:

- `go test ./... -count=1`: passou;
- testes não-E2E do módulo `site/`: passaram;
- testes do módulo externo `examples/application-patterns`: passaram;
- `go test ./site/tests/e2e/... -count=1 -timeout 15m`: passou em 328,059 s;
- `golangci-lint run` nos módulos root e `site/`: zero issues;
- `templ generate`, `go run ./cmd/skillgen`, `go run ./cmd/vendorgen -check`
  e `git diff --check`: passaram sem drift;
- os cinco recursos CDN default foram baixados e comparados byte a byte com os
  arquivos embedded usados para gerar SRI;
- revisão independente final: nenhum achado restante ou bloqueador de release.

## Resultado do release

- PR [#187](https://github.com/araihu/goshtoso/pull/187) integrado em `main`;
- commit de merge e alvo da tag: `6e1b94a473d3e6903347c75955b126b980abde32`;
- tag anotada `v0.0.13` publicada e resolvida para o mesmo commit;
- workflow de release concluído com sucesso e release público criado;
- `styles.css` e `goshtoso-theme.css` publicados como assets do release;
- descoberta remota confirmou as cinco skills, incluindo a versão atualizada de
  `using-goshtoso`;
- follow-up pós-tag fixa o módulo do site em `v0.0.13` e habilita o link de
  comparação do changelog, que retornava 404 antes da tag existir.
