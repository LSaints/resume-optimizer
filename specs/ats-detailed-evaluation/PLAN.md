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
- Google AI Studio (Gemini) para análise via prompt
- React 19 + TypeScript 6 + Vite 8 (frontend)

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
- Ownership validado no service layer antes de qualquer operação
- CSS Modules + design tokens para estilização
- Tipos TypeScript em `src/types/` refletem responses da API
- Páginas em `src/pages/<Name>Page.tsx` com `*.module.css`

### State Before This Feature

O sistema prompt do `AtsScoringServices` já solicita à IA o JSON completo com breakdown, matchedKeywords, missingKeywords e recommendations. No entanto:

- O parser `parseEvaluationResponse` extrai apenas `score`, `summary` e `details`, ignorando os demais campos
- A entidade `AtsEvaluation` não possui campos para breakdown, keywords ou recommendations
- A tabela `ats_evaluations` não possui colunas para esses dados
- `AtsEvaluationResponse` não inclui breakdown, keywords ou recommendations
- A frontend `EvaluationResultPage` exibe apenas score, summary e details

---

## Arquitetura

A feature enriquece o pipeline existente de avaliação ATS nos mesmos moldes da arquitetura em camadas. Nenhuma nova rota, entidade ou serviço será criado — apenas os artefatos existentes serão estendidos.

---

## Estrutura Técnica

### Entidades (`internal/domain/entities/ats_evaluation.go`)

Novos campos na struct `AtsEvaluation`:

- `BreakdownKeywordMatch float64` `json:"breakdownKeywordMatch"` — subtotal do critério (máx 3.0)
- `BreakdownTechnical float64` `json:"breakdownTechnical"` — subtotal do critério (máx 2.5)
- `BreakdownExperience float64` `json:"breakdownExperience"` — subtotal do critério (máx 2.0)
- `BreakdownImpact float64` `json:"breakdownImpact"` — subtotal do critério (máx 1.5)
- `BreakdownReadability float64` `json:"breakdownReadability"` — subtotal do critério (máx 1.0)
- `MatchedKeywords string` `json:"matchedKeywords"` — JSON array de strings (ex: `["Go","SQL"]`)
- `MissingKeywords string` `json:"missingKeywords"` — JSON array de strings
- `Recommendations string` `json:"recommendations"` — JSON array de strings

Os campos de breakdown são `float64` para armazenamento direto no SQLite REAL.
Os campos de lista (`MatchedKeywords`, `MissingKeywords`, `Recommendations`) são armazenados como `string` contendo JSON serializado, seguindo o padrão de `RawResponse`.

### Banco de Dados

Novas colunas na tabela `ats_evaluations`:

```sql
ALTER TABLE ats_evaluations ADD COLUMN breakdown_keyword_match REAL DEFAULT 0;
ALTER TABLE ats_evaluations ADD COLUMN breakdown_technical REAL DEFAULT 0;
ALTER TABLE ats_evaluations ADD COLUMN breakdown_experience REAL DEFAULT 0;
ALTER TABLE ats_evaluations ADD COLUMN breakdown_impact REAL DEFAULT 0;
ALTER TABLE ats_evaluations ADD COLUMN breakdown_readability REAL DEFAULT 0;
ALTER TABLE ats_evaluations ADD COLUMN matched_keywords TEXT DEFAULT '[]';
ALTER TABLE ats_evaluations ADD COLUMN missing_keywords TEXT DEFAULT '[]';
ALTER TABLE ats_evaluations ADD COLUMN recommendations TEXT DEFAULT '[]';
```

Para novas instalações, o `CREATE TABLE` em `db.go` deve incluir essas colunas desde o início.

### DTOs

**`AtsEvaluationResponse`** — novos campos:

```go
type AtsEvaluationResponse struct {
    // campos existentes...
    BreakdownKeywordMatch  float64  `json:"breakdownKeywordMatch"`
    BreakdownTechnical     float64  `json:"breakdownTechnical"`
    BreakdownExperience   float64  `json:"breakdownExperience"`
    BreakdownImpact       float64  `json:"breakdownImpact"`
    BreakdownReadability  float64  `json:"breakdownReadability"`
    MatchedKeywords       []string `json:"matchedKeywords"`
    MissingKeywords       []string `json:"missingKeywords"`
    Recommendations       []string `json:"recommendations"`
}
```

`AtsEvaluationSummaryResponse` — sem alterações (breakdown e listas são dados detalhados, não necessários na listagem).

### Parser (`parseEvaluationResponse`)

O parser atual deve ser atualizado para extrair os novos campos da resposta JSON da IA.

Estrutura interna para unmarshal:

```go
type atsScoreResponse struct {
    Score              float64           `json:"score"`
    Summary            string            `json:"summary"`
    Details            string            `json:"details"`
    Breakdown          atsBreakdown      `json:"breakdown"`
    MatchedKeywords    []string          `json:"matchedKeywords"`
    MissingKeywords    []string          `json:"missingKeywords"`
    Recommendations    []string          `json:"recommendations"`
}

type atsBreakdown struct {
    KeywordMatch          float64 `json:"keywordMatch"`
    TechnicalCompatibility float64 `json:"technicalCompatibility"`
    ProfessionalExperience float64 `json:"professionalExperience"`
    ImpactAndResults      float64 `json:"impactAndResults"`
    AtsReadability        float64 `json:"atsReadability"`
}
```

Validações adicionais no parser:
- Cada subtotal deve estar entre 0 e seu máximo (3.0, 2.5, 2.0, 1.5, 1.0 respectivamente)
- Se breakdown não for enviado, subtotais default 0 — sem quebrar compatibilidade retroativa

### Serviço (`AtsScoringServices`)

O método `Evaluate` deve ser atualizado para:
1. Passar os novos campos extraídos pelo parser para a entidade `AtsEvaluation`
2. Serializar `MatchedKeywords`, `MissingKeywords` e `Recommendations` como strings JSON para persistência
3. Preencher `toResponse` com os novos campos, desserializando as strings JSON de volta para `[]string`
4. Fallback silencioso: se o JSON da IA não incluir breakdown/keywords/recommendations, usar valores zero (compatibilidade com respostas antigas)

### Repositório (`AtsEvaluationRepository`)

O INSERT deve incluir as novas colunas. O SELECT deve ler as novas colunas.

### Endpoints

Nenhuma alteração nas rotas existentes:

| Método | Rota | Descrição |
|--------|------|-----------|
| POST | /v1/resumes/{resumeID}/evaluate | Avaliar currículo (agora com dados enriquecidos) |
| GET | /v1/resumes/{resumeID}/evaluations | Listar (sem breakdown) |
| GET | /v1/resumes/{resumeID}/evaluations/{evaluationID} | Obter avaliação (agora com breakdown, keywords, recomendações) |

### Frontend — Tipos TypeScript (`src/types/evaluation.ts`)

```typescript
export interface AtsBreakdown {
  keywordMatch: number
  technicalCompatibility: number
  professionalExperience: number
  impactAndResults: number
  atsReadability: number
}

export interface EvaluationResponse {
  id: string
  resumeId: string
  jobId: string
  score: number
  summary: string
  details: string
  breakdownKeywordMatch: number
  breakdownTechnical: number
  breakdownExperience: number
  breakdownImpact: number
  breakdownReadability: number
  matchedKeywords: string[]
  missingKeywords: string[]
  recommendations: string[]
  createdAt: string
}
```

### Frontend — Página de Resultado (`EvaluationResultPage.tsx`)

A página existente deve ser expandida para incluir, abaixo do score geral:

1. **Breakdown por critério** — seção com 5 barras de progresso horizontais, cada uma com:
   - Nome do critério (ex: "Correspondência de Palavras-chave")
   - Peso do critério (ex: "30%")
   - Nota obtida / nota máxima (ex: "2.6 / 3.0")
   - Barra de progresso preenchida proporcionalmente
   - Cor semântica baseada na proporção (verde ≥ 70%, amarelo ≥ 40%, vermelho < 40%)

2. **Palavras-chave** — duas seções lado a lado (ou empilhadas em mobile):
   - "Palavras-chave Encontradas" — tags verdes
   - "Palavras-chave Ausentes" — tags vermelhas

3. **Recomendações** — lista numerada com ícone de "💡" ou "→" no início de cada item

### Frontend — Service (`src/services/evaluationService.ts`)

Sem alterações — os mesmos endpoints retornam agora mais dados no response.

### Tratamento de Compatibilidade Retroativa

Avaliações criadas antes desta feature terão:
- `breakdownKeywordMatch = 0`, `breakdownTechnical = 0`, etc.
- `matchedKeywords = []`, `missingKeywords = []`, `recommendations = []`

A página de resultado deve continuar funcionando nesses casos, simplesmente ocultando as seções de breakdown, keywords e recomendações se todos os valores forem zero/vazios.

---

## Configuração

Nenhuma configuração adicional. Reutiliza as mesmas variáveis de ambiente (`GEMINI_API_KEY`, `GEMINI_MODEL`).

---

## Decisões Técnicas

- **ALTER TABLE com DEFAULT**: usado para não quebrar banco existente; avaliações antigas recebem valores default (0 / `[]`)
- **Listas como strings JSON no banco**: evita criar tabelas N:N para palavras-chave, mantendo simplicidade; o Go serializa/desserializa na camada de serviço/response
- **Breakdown armazenado em colunas individuais**: facilita consultas SQL futuras (ex: média de keywordMatch por vaga); serialização ocorre apenas no response
- **Compatibilidade retroativa**: o parser trata ausência dos novos campos como zero/empty; o frontend oculta seções sem dados
- **Navegação inalterada**: nenhuma nova rota; as páginas existentes de avaliação e resultado são enriquecidas

---

## Assumptions

- O system prompt existente já solicita o JSON completo com todos os campos — nenhuma alteração no prompt é necessária, apenas no parser e armazenamento
- A IA do Google Gemini já retorna os campos de breakdown, keywords e recommendations (conforme solicitado no prompt atual)
