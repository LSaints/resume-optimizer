# PLAN.md

## Existing Context

### Architecture

Clean Architecture (Layered) — SPA frontend com React 19 + TypeScript 6 + Vite 8

### Stack

- **Frontend**: React 19, TypeScript 6, Vite 8, react-router-dom 7
- **Backend**: Go 1.26, API REST em `/v1/*`
- **Auth**: JWT via header `Authorization: Bearer <token>`, armazenado em `localStorage`
- **Estilização**: CSS Modules + CSS custom properties (design tokens)

### Existing Conventions

- Componentes próprios sem bibliotecas de UI (shadcn, MUI, etc.)
- CSS Modules com `*.module.css`
- Design tokens em `src/styles/tokens.css`
- Serviços centralizados via `src/services/api.ts` com funções `get`, `post`, `put`, `del`
- Tipos TypeScript em `src/types/` refletem as responses da API
- Páginas em `src/pages/<Name>Page.tsx` com `*.module.css` correspondente
- `AuthContext` para estado global de autenticação; demais estados são locais da página
- Rotas protegidas via `ProtectedRoute` com redirecionamento para `/login`
- Estados de loading com shimmer/skeleton, estados de erro com mensagens em português
- Navegação entre páginas via `useNavigate` do `react-router-dom`

---

## Arquitetura

A feature de avaliação ATS segue o mesmo padrão das páginas de otimização existentes. Serão criadas:

1. **Tipos TypeScript** — interfaces para request e responses da API de avaliação
2. **Service** — funções que chamam os endpoints de avaliação via `api.ts`
3. **Página de Avaliação** — seleção de currículo + vaga + disparo de avaliação
4. **Página de Resultado** — exibição do score, resumo e detalhamento
5. **Página de Histórico** — listagem de avaliações de um currículo
6. **Rotas** — registro das novas páginas no `<App>`

O backend já está implementado com os endpoints:
- `POST /v1/resumes/{resumeID}/evaluate`
- `GET /v1/resumes/{resumeID}/evaluations`
- `GET /v1/resumes/{resumeID}/evaluations/{evaluationID}`

---

## Estrutura Técnica

### Tipos TypeScript (`src/types/`)

**`evaluation.ts`** — novo arquivo:

```typescript
interface EvaluateRequest {
  jobId: string
}

interface EvaluationResponse {
  id: string
  resumeId: string
  jobId: string
  score: number
  summary: string
  details: string
  createdAt: string
}

interface EvaluationSummaryResponse {
  id: string
  resumeId: string
  jobId: string
  score: number
  summary: string
  createdAt: string
}
```

### Serviços (`src/services/`)

**`evaluationService.ts`** — novo arquivo:

| Função | Descrição | Endpoint |
|---|---|---|
| `evaluate(resumeId, jobId)` | Dispara avaliação ATS | `POST /v1/resumes/{resumeID}/evaluate` |
| `listByResume(resumeId)` | Lista avaliações de um currículo | `GET /v1/resumes/{resumeID}/evaluations` |
| `getByID(resumeId, evaluationId)` | Obtém avaliação específica | `GET /v1/resumes/{resumeID}/evaluations/{evaluationID}` |

### Páginas (`src/pages/`)

| Página | Rota | Descrição |
|---|---|---|
| `EvaluatePage` | `/evaluate` | Selecionar currículo + vaga e disparar avaliação |
| `EvaluationResultPage` | `/evaluations/:resumeId/:evaluationId` | Exibir resultado da avaliação |
| `EvaluationHistoryPage` | `/resumes/:id/evaluations` | Listar histórico de avaliações de um currículo |

#### EvaluatePage

Similar à `OptimizePage` existente:

- Carrega lista de currículos e vagas do usuário via services
- Select de currículo (carregado via `resumeService.list()`)
- Select de vaga (carregado via `jobService.list()`)
- Botão "Avaliar" desabilitado até que ambos sejam selecionados
- Botão mostra "Avaliando..." com spinner durante a chamada
- Após sucesso, redireciona para `/evaluations/{resumeId}/{evaluationId}`
- Estados: loading (skeleton), erro, formulário

#### EvaluationResultPage

Exibe o resultado completo de uma avaliação:

- Pontuação numérica de 0 a 10 (destacada visualmente — ex: badge, gauge ou número grande)
- Resumo textual da avaliação
- Detalhamento com pontos fortes e oportunidades de melhoria
- Metadados: nome do currículo, título da vaga, data da avaliação
- Botão "Nova avaliação" redireciona para `/evaluate`
- Botão "Ver histórico" redireciona para `/resumes/{resumeId}/evaluations`
- Estados: loading (skeleton), erro, não encontrado

#### EvaluationHistoryPage

Similar à `OptimizationHistoryPage` existente:

- Lista avaliações de um currículo com: score (destacado), resumo, vaga associada, data
- Cada item é clicável e redireciona para `/evaluations/{resumeId}/{evaluationId}`
- Botão de voltar para a página anterior
- Estado vazio com link para `/evaluate`
- Estados: loading (skeleton), erro, lista vazia

### Rotas (`src/App.tsx`)

Adicionar no bloco protegido:

```tsx
<Route path="/evaluate" element={<EvaluatePage />} />
<Route path="/evaluations/:resumeId/:evaluationId" element={<EvaluationResultPage />} />
<Route path="/resumes/:id/evaluations" element={<EvaluationHistoryPage />} />
```

### Navegação (`Header.tsx` ou `Layout.tsx`)

Adicionar link "Avaliar" na navegação principal, consistente com os links existentes (Dashboard, Currículos, Vagas, Otimizar).

---

## Fluxos

### Avaliação

```
Usuário → /evaluate
  → carrega lista de currículos e vagas do usuário
  → seleciona um currículo e uma vaga em selects
  → clica "Avaliar"
  → evaluationService.evaluate(resumeId, jobId)
  → POST /v1/resumes/{resumeID}/evaluate
  → exibe loading durante a chamada (pode levar até 60s)
  → redireciona para /evaluations/{resumeId}/{evaluationId}
```

### Visualização de resultado

```
Usuário → /evaluations/{resumeId}/{evaluationId}
  → evaluationService.getByID(resumeId, evaluationId)
  → exibe score (0-10) em destaque
  → exibe resumo textual
  → exibe detalhamento (pontos fortes + oportunidades)
  → botões: "Nova avaliação", "Ver histórico"
```

### Histórico

```
Usuário → /resumes/{id}/evaluations
  → evaluationService.listByResume(resumeId)
  → lista cards com score, resumo, vaga, data
  → clica em um card → /evaluations/{resumeId}/{evaluationId}
```

---

## Componentes Existentes Reutilizados

- `Button` — ações primárias e secundárias
- `Select` — seleção de currículo e vaga
- `LoadingSpinner` — estado de carregamento
- `Modal` — confirmação de exclusão (se aplicável)

Nenhum novo componente de UI precisa ser criado. Os componentes existentes atendem às necessidades.

---

## Decisões Técnicas

- **Reuso do padrão de otimização**: as páginas de avaliação seguem a mesma estrutura, nomeação e comportamento das páginas de otimização já implementadas, garantindo consistência visual e de código
- **Sem Details na listagem**: `EvaluationSummaryResponse` exclui `details` para manter a listagem leve, mesma abordagem usada em `OptimizeSummaryResponse`
- **Score em destaque**: a pontuação é o elemento central da avaliação e deve ter destaque visual (tamanho grande, cor semântica) para rápida interpretação
- **Todas as páginas protegidas**: as rotas ficam dentro do `ProtectedRoute`, herdando a verificação de autenticação
- **Link "Avaliar" na navegação principal**: posicionado ao lado de "Otimizar" no Header, indicando que é uma funcionalidade complementar
