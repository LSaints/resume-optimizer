# PLAN.md

## Existing Context

### Architecture

Clean Architecture (Layered)

```
cmd/api/
internal/
  apresentation/   -> HTTP layer (handlers, routes, middleware)
  application/     -> Use cases (services, DTOs)
  domain/          -> Business entities
  infrastructure/  -> Data access (db, repositories)
```

### Stack

- Go 1.26 (stdlib `net/http`, `database/sql`, `encoding/json`)
- SQLite via `mattn/go-sqlite3`
- UUID via `google/uuid`
- React 19 + TypeScript 6 + Vite 8

### Existing Conventions

- Concrete structs with constructor functions (`New<Type>`)
- No interfaces, no DI framework
- Raw SQL queries
- Versioned routes: `/v1/<resource>`
- Folder naming: `apresentation/` (Portuguese spelling)
- Files: `snake_case.go`
- DTOs: `<Action><Entity>Request` / `<Entity>Response`
- Messages in Portuguese
- Dependencies wired manually in `routes.go`
- `userID` injetado no contexto da requisição via `AuthMiddleware`

### Estado Atual

- `TypstRenderer` frontend exibe apenas código-fonte bruto em `<pre>`
- `TypstViewerPage` mostra metadados + código bruto de uma otimização
- `OptimizePage` permite selecionar currículo + vaga e disparar otimização
- Não existe página de histórico de otimizações
- Backend não possui serviço de renderização Typst

---

## Arquitetura

O módulo de renderização estenderá a camada de application com um novo serviço que converte código Typst em SVG utilizando o binário `typst` via `os/exec`. O SVG será servido pelo backend e renderizado no frontend como imagem inline. A entidade `ResumeOptimized` não será modificada. Uma nova rota de renderização será adicionada. O frontend ganhará um componente de visualização com alternância entre renderização e código-fonte, além de uma nova página de histórico com thumbnails.

---

## Estrutura Técnica

### Entidades

- `ResumeOptimized` — `internal/domain/entities/resume_optimized.go` (já existe, sem alterações)

### Serviços

- `TypstRenderService` — `internal/application/services/typst_render_service.go`
  - `RenderToSVG(typstContent string) (string, error)`
    - Escreve o conteúdo Typst em arquivo temporário (`*.typ`)
    - Executa `typst compile input.typ output.svg --format svg`
    - Lê o SVG gerado e retorna como string
    - Timeout de 15 segundos via `context.WithTimeout`
    - Remove arquivos temporários ao final (defer)
    - Retorna erro "renderizador typst nao disponivel" se `typst` não estiver no PATH
    - Retorna erro "erro ao renderizar documento" se a compilação falhar
  - `RenderToPDF(typstContent string) ([]byte, error)`
    - Mesmo fluxo, formato PDF
    - Retorna os bytes do PDF para download

### DTOs

- `RenderResponse` — `internal/application/responses/render_response.go`
  - `SvgContent string` `json:"svgContent"` — SVG inline renderizado

### Repositórios

Nenhum novo repositório necessário. O `OptimizationRepository` já existente será usado para buscar o conteúdo Typst.

### Endpoints

| Método | Rota | Autenticação | Descrição |
|---|---|---|---|
| GET | `/v1/optimizations/{optimizationID}/render` | Protegida | Retorna SVG renderizado da otimização |
| GET | `/v1/optimizations/{optimizationID}/render/pdf` | Protegida | Download do PDF renderizado |

### Handlers

- `RenderHandler` — `internal/apresentation/handlers/render_handlers.go`
  - `RenderSVG(w, r)` — `GET /v1/optimizations/{optimizationID}/render`
    - Lê `optimizationID` da URL via `r.PathValue("optimizationID")`
    - Lê `userID` do contexto
    - Busca a otimização via `OptimizationServices.GetByID` (já valida ownership)
    - Chama `TypstRenderService.RenderToSVG(optimization.TypstContent)`
    - Retorna 200 com `RenderResponse` contendo SVG inline
    - Em caso de erro do renderizador, retorna 502
  - `RenderPDF(w, r)` — `GET /v1/optimizations/{optimizationID}/render/pdf`
    - Mesmo fluxo, retorna PDF com `Content-Type: application/pdf`
    - Header `Content-Disposition: attachment; filename="curriculo-otimizado.pdf"`

### Integração com Serviços Existentes

- `RenderHandler` depende de `OptimizationServices` (já existente) e `TypstRenderService` (novo)
- `OptimizationServices.GetByID` já valida ownership do usuário — reutilizado

---

## Configuração

### Dependência de Sistema

- Binário `typst` deve estar instalado no servidor e acessível via PATH
- Versão mínima: Typst 0.12+
- Instalação recomendada: `cargo install typst-cli` ou `brew install typst`

### Variáveis de Ambiente

Nenhuma nova variável de ambiente necessária.

---

## Frontend

### Componentes

- `TypstRenderer` (substituído — `frontend/src/components/TypstRenderer.tsx`)
  - Props: `{ content: string }`
  - Estado interno: `viewMode: "rendered" | "source"`
  - Busca o SVG do backend via `GET /v1/optimizations/{id}/render`
  - Exibe SVG inline quando em modo `rendered`
  - Exibe código-fonte formatado quando em modo `source`
  - Botão de alternância entre os modos
  - Botão "Copiar código" (visível apenas no modo source)
  - Botão "Baixar PDF" (visível no modo rendered)
  - Loading spinner enquanto o SVG carrega
  - Tratamento de erro se a renderização falhar

### Novas Páginas

- `OptimizationHistoryPage` — `frontend/src/pages/OptimizationHistoryPage.tsx`
  - Rota: `/resumes/:id/optimizations`
  - Lista todas as otimizações de um currículo
  - Cada item exibe: thumbnail SVG renderizado, nome do currículo, título da vaga, data
  - Clique no item navega para `/optimizations/:resumeId/:optimizationId`
  - Botão para excluir otimização
  - Botão "Nova otimização" volta para `/optimize`

### Páginas Alteradas

- `TypstViewerPage` (modificada)
  - Agora usa o `TypstRenderer` atualizado (com renderização SVG)
  - Mantém metadados e ações existentes

### Novas Rotas

| Path | Página | Proteção |
|---|---|---|
| `/resumes/:id/optimizations` | `OptimizationHistoryPage` | Sim (ProtectedRoute) |

### Novos Serviços

- `renderService` — `frontend/src/services/renderService.ts`
  - `getRenderSVG(optimizationID: string): Promise<string>` — busca SVG renderizado
  - `getDownloadPDFURL(optimizationID: string): string` — retorna URL para download

### Novos Tipos

- `RenderResponse` — `frontend/src/types/render.ts`
  - `{ svgContent: string }`

---

## Decisões Técnicas

- **Renderização server-side via CLI**: utilizar o binário `typst` via `os/exec` por ser a abordagem mais estável e evitar dependência WASM pesada no frontend. O SVG gerado é inline e exibido diretamente no HTML.
- **Sem cache de renderização**: por simplicidade inicial, cada requisição renderiza sob demanda. Cache pode ser adicionado futuramente salvando o SVG no banco ou em disco.
- **Reuso da validação de ownership**: o `RenderHandler` reutiliza `OptimizationServices.GetByID` que já valida que a otimização pertence ao usuário logado, evitando duplicação de lógica.
- **SVG como formato primário**: SVG é inline, escalável, e pode ser estilizado via CSS. PDF é oferecido como download.
- **Temporary files em diretório padrão**: usar `os.CreateTemp` para garantir isolamento entre requisições e limpeza automática via `defer os.Remove`.
- **Timeout de 15s**: a compilação Typst é rápida para documentos de currículo (tipicamente < 5s). 15s é um limite seguro.
- **Dependência `typst` documentada**: o binário deve ser instalado no ambiente de produção. A ausência retorna erro 502 com mensagem clara.

---

## Fluxo de renderização

```
GET /v1/optimizations/{optimizationID}/render
  → RenderHandler.RenderSVG
    → AuthMiddleware (já garante userID no contexto)
    → OptimizationServices.GetByID(userID, optimizationID)
      → valida ownership da otimização
    → TypstRenderService.RenderToSVG(optimization.TypstContent)
      → escreve .typ em temp file
      → exec.Command("typst", "compile", input, output, "--format", "svg")
      → lê output SVG
      → deleta temp files
      → retorna SVG string
    → retorna RenderResponse { svgContent: "<svg>...</svg>" }

Frontend:
  TypstRenderer.render() faz fetch para /v1/optimizations/{id}/render
  → exibe SVG inline no DOM
  → se falhar, mostra mensagem de erro + fallback para código-fonte
```

---

## Dependências a adicionar

### Backend

- Nenhuma nova dependência Go. `os/exec` é stdlib.

### Frontend

- Nenhuma nova dependência npm. Fetch nativo + SVG inline.

### Sistema

- Binário `typst-cli` instalado (via cargo, brew, ou package manager)
