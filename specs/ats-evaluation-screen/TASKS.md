# TASKS.md

## Task 1

Criar tipos TypeScript para avaliação ATS

### Objetivo

Criar `src/types/evaluation.ts` com interfaces `EvaluateRequest`, `EvaluationResponse` e `EvaluationSummaryResponse` refletindo os contratos da API.

### Validação

- Arquivo `src/types/evaluation.ts` existe
- `EvaluateRequest` possui campo `jobId: string`
- `EvaluationResponse` possui campos: `id`, `resumeId`, `jobId`, `score`, `summary`, `details`, `createdAt`
- `EvaluationSummaryResponse` possui campos: `id`, `resumeId`, `jobId`, `score`, `summary`, `createdAt` (sem `details`)
- Tipos seguem o padrão de nomenclatura dos arquivos existentes em `src/types/`

---

## Task 2

Criar service de avaliação ATS

### Objetivo

Criar `src/services/evaluationService.ts` com funções `evaluate`, `listByResume` e `getByID` que chamam os endpoints da API via `api.ts`.

### Validação

- `evaluate(resumeId, jobId)` chama `POST /v1/resumes/{resumeID}/evaluate` com body `{ jobId }`
- `listByResume(resumeId)` chama `GET /v1/resumes/{resumeID}/evaluations`
- `getByID(resumeId, evaluationId)` chama `GET /v1/resumes/{resumeID}/evaluations/{evaluationID}`
- Todas as funções retornam Promises tipadas com os tipos do Task 1
- O service segue o mesmo padrão de `optimizationService.ts`

---

## Task 3

Criar página de avaliação ATS (EvaluatePage)

### Objetivo

Implementar `src/pages/EvaluatePage.tsx` com seleção de currículo e vaga, disparo de avaliação e redirecionamento para o resultado.

### Validação

- `/evaluate` carrega lista de currículos via `resumeService.list()` e vagas via `jobService.list()`
- Exibe Select para currículo com options carregadas
- Exibe Select para vaga com options carregadas
- Botão "Avaliar" fica desabilitado até que ambos os selects tenham valor selecionado
- Botão exibe "Avaliando..." com spinner durante a chamada
- Após sucesso, redireciona para `/evaluations/{resumeId}/{evaluationId}`
- Em caso de erro da API, exibe mensagem amigável em português
- Estados: loading (skeleton), erro, formulário preenchido
- Segue o mesmo padrão visual e de código de `OptimizePage.tsx`

---

## Task 4

Criar página de resultado de avaliação (EvaluationResultPage)

### Objetivo

Implementar `src/pages/EvaluationResultPage.tsx` que exibe o score, resumo e detalhamento de uma avaliação específica.

### Validação

- `/evaluations/:resumeId/:evaluationId` carrega avaliação via `evaluationService.getByID()`
- Exibe score (0-10) em destaque visual — número grande com cor semântica (verde para alta pontuação, amarelo para média, vermelho para baixa)
- Exibe resumo textual da avaliação
- Exibe detalhamento completo com pontos fortes e oportunidades de melhoria
- Exibe metadados: nome do currículo, título da vaga (buscados via `resumeService.get()` e `jobService.get()`), data formatada em português
- Botão "Nova avaliação" redireciona para `/evaluate`
- Botão "Ver histórico" redireciona para `/resumes/{resumeId}/evaluations`
- Estados: loading (skeleton), erro, avaliação não encontrada (404)
- Score sem detalhes é exibido mesmo se `details` estiver vazio (apenas oculta a seção)

---

## Task 5

Criar página de histórico de avaliações (EvaluationHistoryPage)

### Objetivo

Implementar `src/pages/EvaluationHistoryPage.tsx` listando todas as avaliações de um currículo com score e resumo.

### Validação

- `/resumes/:id/evaluations` carrega avaliações via `evaluationService.listByResume()`
- Cada card exibe: score (em badge ou destaque), resumo textual, título da vaga (enriquecido via `jobService.list()`), data formatada
- Cada card é clicável e redireciona para `/evaluations/{resumeId}/{evaluationId}`
- Botão de voltar na parte superior
- Estado vazio exibe "Nenhuma avaliação encontrada" com link para `/evaluate`
- Estados: loading (skeleton cards), erro com mensagem em português
- Segue o mesmo padrão visual e de código de `OptimizationHistoryPage.tsx`

---

## Task 6

Registrar rotas e adicionar navegação

### Objetivo

Atualizar `src/App.tsx` para registrar as novas rotas de avaliação e adicionar link "Avaliar" na navegação principal.

### Validação

- Rotas registradas dentro do `ProtectedRoute`:
  - `<Route path="/evaluate" element={<EvaluatePage />} />`
  - `<Route path="/evaluations/:resumeId/:evaluationId" element={<EvaluationResultPage />} />`
  - `<Route path="/resumes/:id/evaluations" element={<EvaluationHistoryPage />} />`
- Link "Avaliar" adicionado na navegação do Header (ao lado de "Otimizar")
- Navegação entre todas as páginas funciona (incluindo links "Ver histórico" e "Nova avaliação")
- Rotas protegidas redirecionam para `/login` se não autenticado
- `npm run build` executa sem erros

---

## Task 7

Testar fluxo completo de avaliação ATS

### Objetivo

Validar o fluxo ponta a ponta da avaliação ATS, incluindo todos os cenários de sucesso e erro.

### Validação

- Navegar para `/evaluate` → selects carregam currículos e vagas do usuário
- Selecionar currículo e vaga → clicar "Avaliar" → aguardar loading → redirecionar para resultado com score, resumo e detalhes
- Score exibido está entre 0 e 10 com até uma casa decimal
- Cores semânticas do score refletem o valor (alta/média/baixa)
- Navegar para o histórico de avaliações do currículo → lista exibe todas as avaliações realizadas
- Clicar em uma avaliação no histórico → abre a página de detalhes completa
- Selecionar currículo ou vaga inexistente via URL → exibe erro amigável
- Acessar rotas protegidas sem token → redireciona para `/login`
- Build de produção (`npm run build`) executa sem erros
- Interface é responsiva em mobile (375px) e desktop (1440px)
