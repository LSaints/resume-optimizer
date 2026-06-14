# TASKS.md

## Task 1

Adicionar colunas de breakdown, keywords e recomendações na tabela `ats_evaluations`

### Objetivo

Adicionar colunas para armazenar os novos campos de avaliação detalhada na tabela `ats_evaluations` em `internal/infrastructure/data/db.go`.

### Validação

- `CREATE TABLE IF NOT EXISTS ats_evaluations` em `db.go` inclui as colunas:
  - `breakdown_keyword_match REAL DEFAULT 0`
  - `breakdown_technical REAL DEFAULT 0`
  - `breakdown_experience REAL DEFAULT 0`
  - `breakdown_impact REAL DEFAULT 0`
  - `breakdown_readability REAL DEFAULT 0`
  - `matched_keywords TEXT DEFAULT '[]'`
  - `missing_keywords TEXT DEFAULT '[]'`
  - `recommendations TEXT DEFAULT '[]'`
- A aplicação inicializa sem erros com banco novo
- A aplicação inicializa sem erros com banco existente (colunas já existentes não causam erro — usar `ALTER TABLE` não é necessário se o `CREATE TABLE IF NOT EXISTS` já incluir as colunas, mas para banco existente pode ser necessário adicionar `ALTER TABLE` com verificação)

---

## Task 2

Atualizar entidade `AtsEvaluation` com novos campos

### Objetivo

Adicionar campos de breakdown, matchedKeywords, missingKeywords e recommendations na struct `AtsEvaluation` em `internal/domain/entities/ats_evaluation.go`.

### Validação

- `AtsEvaluation` possui novos campos:
  - `BreakdownKeywordMatch float64`
  - `BreakdownTechnical float64`
  - `BreakdownExperience float64`
  - `BreakdownImpact float64`
  - `BreakdownReadability float64`
  - `MatchedKeywords string`
  - `MissingKeywords string`
  - `Recommendations string`
- Todos os campos possuem tags JSON no formato `json:"camelCase"`
- O código compila sem erros

---

## Task 3

Atualizar repositório `AtsEvaluationRepository` para lidar com as novas colunas

### Objetivo

Atualizar os métodos `Create`, `GetByID` e `GetByResumeID` em `internal/infrastructure/repositories/ats_evaluation_repository.go` para incluir as novas colunas no INSERT e SELECT.

### Validação

- `Create` insere as 8 novas colunas no INSERT
- `GetByID` lê as 8 novas colunas no Scan
- `GetByResumeID` lê as 8 novas colunas no Scan
- O código compila sem erros
- Criação e leitura de avaliação retornam os novos campos corretamente

---

## Task 4

Atualizar parser da resposta da IA para extrair breakdown, keywords e recomendações

### Objetivo

Atualizar a struct `atsScoreResponse` e a função `parseEvaluationResponse` em `internal/application/services/ats_scoring_services.go` para extrair breakdown, matchedKeywords, missingKeywords e recommendations do JSON retornado pela IA.

### Validação

- `atsScoreResponse` inclui campos `Breakdown`, `MatchedKeywords`, `MissingKeywords`, `Recommendations`
- `parseEvaluationResponse` retorna breakdown (5 float64), matchedKeywords ([]string), missingKeywords ([]string), recommendations ([]string) além dos campos existentes
- Cada subtotal do breakdown é validado individualmente contra seu máximo permitido (3.0, 2.5, 2.0, 1.5, 1.0)
- Se breakdown ou listas não estiverem presentes no JSON, retorna valores zero/default sem erro
- Testes unitários do parser com JSON completo, JSON parcial e JSON sem os novos campos

---

## Task 5

Atualizar `AtsEvaluationResponse` com novos campos

### Objetivo

Adicionar campos de breakdown (individuais), matchedKeywords, missingKeywords e recommendations na struct `AtsEvaluationResponse` em `internal/application/responses/evaluation_response.go`.

### Validação

- `AtsEvaluationResponse` possui:
  - `BreakdownKeywordMatch float64` `json:"breakdownKeywordMatch"`
  - `BreakdownTechnical float64` `json:"breakdownTechnical"`
  - `BreakdownExperience float64` `json:"breakdownExperience"`
  - `BreakdownImpact float64` `json:"breakdownImpact"`
  - `BreakdownReadability float64` `json:"breakdownReadability"`
  - `MatchedKeywords []string` `json:"matchedKeywords"`
  - `MissingKeywords []string` `json:"missingKeywords"`
  - `Recommendations []string` `json:"recommendations"`
- `AtsEvaluationSummaryResponse` permanece inalterado (sem os novos campos)
- O código compila sem erros

---

## Task 6

Atualizar `AtsScoringServices` para preencher e retornar os novos campos

### Objetivo

Atualizar o método `Evaluate` em `internal/application/services/ats_scoring_services.go` para serializar as listas como JSON strings na entidade, e atualizar `toResponse` para desserializar as strings JSON de volta para `[]string` e preencher os campos de breakdown.

### Validação

- `Evaluate` serializa `matchedKeywords`, `missingKeywords` e `recommendations` como JSON strings ao criar a entidade
- `Evaluate` preenche os 5 campos de breakdown na entidade
- `toResponse` desserializa as strings JSON para `[]string` no response
- `toResponse` preenche os 5 campos de breakdown no response
- Se as strings JSON forem inválidas ou vazias, retorna slice vazio sem erro
- O código compila sem erros

---

## Task 7

Atualizar tipos TypeScript no frontend

### Objetivo

Atualizar `src/types/evaluation.ts` com os novos campos da resposta da API, incluindo a interface `AtsBreakdown`.

### Validação

- `EvaluationResponse` possui os campos: `breakdownKeywordMatch`, `breakdownTechnical`, `breakdownExperience`, `breakdownImpact`, `breakdownReadability` (number)
- `EvaluationResponse` possui os campos: `matchedKeywords`, `missingKeywords`, `recommendations` (string[])
- `EvaluationSummaryResponse` permanece inalterado
- `npm run build` executa sem erros

---

## Task 8

Adicionar seção de breakdown visual na página de resultado

### Objetivo

Expandir `src/pages/EvaluationResultPage.tsx` e seu CSS module para exibir uma seção de breakdown com barras de progresso horizontais para cada critério, exibindo nome, peso, nota obtida/máxima e barra colorida.

### Validação

- Se breakdown tiver valores > 0, exibe seção com 5 barras de progresso:
  - "Correspondência de Palavras-chave" (30%) — nota / 3.0
  - "Compatibilidade Técnica" (25%) — nota / 2.5
  - "Experiência Profissional" (20%) — nota / 2.0
  - "Impacto e Resultados" (15%) — nota / 1.5
  - "Legibilidade ATS" (10%) — nota / 1.0
- Cada barra tem cor semântica: verde (≥ 70%), amarelo (≥ 40%), vermelho (< 40%)
- Se todos os breakdowns forem zero, a seção não é exibida (compatibilidade retroativa)
- `npm run build` executa sem erros

---

## Task 9

Adicionar seções de palavras-chave e recomendações na página de resultado

### Objetivo

Expandir `src/pages/EvaluationResultPage.tsx` e seu CSS module para exibir listas de palavras-chave encontradas (tags verdes), palavras-chave ausentes (tags vermelhas) e recomendações (lista numerada).

### Validação

- Se `matchedKeywords` não estiver vazia, exibe seção "Palavras-chave Encontradas" com tags em verde
- Se `missingKeywords` não estiver vazia, exibe seção "Palavras-chave Ausentes" com tags em vermelho
- Se `recommendations` não estiver vazia, exibe seção "Recomendações" com lista numerada
- Se alguma lista estiver vazia, sua seção correspondente não é exibida
- Layout responsivo: em mobile as seções de keywords empilham verticalmente
- `npm run build` executa sem erros

---

## Task 10

Testar fluxo completo de avaliação detalhada

### Objetivo

Validar a feature ponta a ponta: desde a requisição de avaliação até a exibição dos dados enriquecidos na página de resultado, incluindo compatibilidade retroativa com avaliações antigas.

### Validação

- Avaliar currículo → retorna 201 com `AtsEvaluationResponse` contendo:
  - `breakdownKeywordMatch`, `breakdownTechnical`, `breakdownExperience`, `breakdownImpact`, `breakdownReadability` com valores entre 0 e seus máximos
  - `matchedKeywords` como array não vazio de strings
  - `missingKeywords` como array de strings
  - `recommendations` como array não vazio de strings
  - `score` igual à soma dos breakdowns (aproximadamente, tolerância de 0.1)
- Página de resultado exibe:
  - Score geral em destaque
  - Breakdown com 5 barras de progresso com cores semânticas
  - Tags de palavras-chave encontradas em verde
  - Tags de palavras-chave ausentes em vermelho
  - Lista numerada de recomendações
- Avaliação existente (sem breakdown) é exibida sem as novas seções (sem erro)
- Listagem (`GET /v1/resumes/{resumeID}/evaluations`) retorna apenas campos resumidos (sem breakdown)
- `npm run build` executa sem erros
- `go build ./cmd/api` executa sem erros
