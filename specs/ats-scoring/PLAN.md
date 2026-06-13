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
- Serviços externos isolados em clientes próprios (ex.: `GeminiClient`)
- Ownership validado no service layer antes de qualquer operação

---

## Arquitetura

O módulo de avaliação ATS seguirá a mesma estrutura em camadas do projeto. Uma nova entidade `AtsEvaluation` será criada em `internal/domain/entities/`. Um novo serviço `AtsScoringServices` orquestrará a lógica de avaliação, reutilizando o `GeminiClient` existente (ou a ser implementado) para comunicação com a API do Google AI Studio com um system prompt específico para scoring ATS.

---

## Estrutura Técnica

### Entidades

- `AtsEvaluation` — `internal/domain/entities/ats_evaluation.go`
  - `ID uuid.UUID` — identificador único
  - `ResumeID uuid.UUID` — FK para o currículo avaliado
  - `JobID uuid.UUID` — FK para a vaga usada como referência
  - `Score float64` — pontuação de 0 a 10 (uma casa decimal)
  - `Summary string` — resumo textual da avaliação
  - `Details string` — detalhamento com pontos fortes e oportunidades de melhoria
  - `RawResponse string` — resposta bruta da IA (para auditoria)
  - `CreatedAt time.Time`

### Serviços

- `AtsScoringServices` — `internal/application/services/ats_scoring_services.go`
  - `Evaluate(userID, resumeID, jobID string) (AtsEvaluationResponse, error)`
    - Busca currículo e valida ownership
    - Busca vaga e valida ownership
    - Monta system prompt específico para scoring ATS
    - Monta user prompt com texto do currículo + descrição da vaga
    - Chama `GeminiClient.SendPrompt`
    - Parseia a resposta da IA para extrair score, summary e details
    - Persiste o resultado em `ats_evaluations`
    - Retorna `AtsEvaluationResponse`
  - `ListByResume(userID, resumeID string) ([]AtsEvaluationSummaryResponse, error)`
    - Lista histórico de avaliações de um currículo
  - `GetByID(userID, evaluationID string) (AtsEvaluationResponse, error)`
    - Retorna uma avaliação específica com todos os detalhes

### GeminiClient (já especificado no módulo `resume-optimization`)

Reutiliza o mesmo `GeminiClient` em `internal/application/services/gemini_client.go`:
- `SendPrompt(systemPrompt, userPrompt string) (string, error)`
- Leitura de `GEMINI_API_KEY` e `GEMINI_MODEL` de variáveis de ambiente
- Timeout de 60 segundos

### System Prompt para Scoring ATS

Definido como constante no serviço `AtsScoringServices`:

```
Você é um especialista em recrutamento e seleção com vasto conhecimento em ATS (Applicant Tracking Systems).

Sua função é analisar a compatibilidade de um currículo com uma descrição de vaga e gerar uma avaliação objetiva.

Analise os seguintes critérios:
1. Correspondência de palavras-chave — presença de termos técnicos, ferramentas e habilidades mencionadas na vaga
2. Adequação de experiência — anos de experiência, nível hierárquico, setor de atuação
3. Compatibilidade de habilidades — hard skills e soft skills requeridas vs. apresentadas
4. Estrutura e formatação — organização, clareza, seções bem definidas
5. Resultados mensuráveis — presença de métricas, realizações quantificáveis
6. Legibilidade para ATS — uso de formato limpo, sem tabelas ou gráficos

Retorne APENAS um JSON válido no seguinte formato, sem formatação markdown:

{
  "score": 7.5,
  "summary": "Resumo da avaliação em português",
  "details": "Detalhamento com pontos fortes e oportunidades de melhoria em português"
}

A pontuação deve ser um número entre 0 e 10, com no máximo uma casa decimal.
```

### DTOs

- `EvaluateResumeRequest` — `internal/application/requests/evaluate_request.go`
  - `JobID string` `json:"jobId"` — ID da vaga para basear a avaliação
- `AtsEvaluationResponse` — `internal/application/responses/evaluation_response.go`
  - `ID string` `json:"id"`
  - `ResumeID string` `json:"resumeId"`
  - `JobID string` `json:"jobId"`
  - `Score float64` `json:"score"`
  - `Summary string` `json:"summary"`
  - `Details string` `json:"details"`
  - `CreatedAt string` `json:"createdAt"`
- `AtsEvaluationSummaryResponse` — mesmo arquivo
  - `ID string` `json:"id"`
  - `ResumeID string` `json:"resumeId"`
  - `JobID string` `json:"jobId"`
  - `Score float64` `json:"score"`
  - `Summary string` `json:"summary"`
  - `CreatedAt string` `json:"createdAt"`
  - Usado na listagem (sem `Details`)

### Repositórios

- `AtsEvaluationRepository` — `internal/infrastructure/repositories/ats_evaluation_repository.go`
  - `Create(eval entities.AtsEvaluation) error`
  - `GetByID(id string) (entities.AtsEvaluation, error)`
  - `GetByResumeID(resumeID string) ([]entities.AtsEvaluation, error)` — ordenado por `created_at DESC`
  - `Delete(id string) error`

### Endpoints

| Método | Rota | Autenticação | Descrição |
|---|---|---|---|
| POST | /v1/resumes/{resumeID}/evaluate | Protegida | Avaliar currículo contra uma vaga |
| GET | /v1/resumes/{resumeID}/evaluations | Protegida | Listar histórico de avaliações do currículo |
| GET | /v1/resumes/{resumeID}/evaluations/{evaluationID} | Protegida | Obter avaliação específica |

### Handlers

- `AtsScoringHandler` — `internal/apresentation/handlers/ats_scoring_handlers.go`
  - `Evaluate(w, r)` — `POST /v1/resumes/{resumeID}/evaluate`
    - Lê `userID` do contexto via `r.Context().Value(middleware.UserIDKey).(string)`
    - Lê `resumeID` da URL via `r.PathValue("resumeID")`
    - Decodifica body JSON com `JobID`
    - Chama `atsScoringServices.Evaluate(userID, resumeID, jobID)`
    - Retorna 201 Created com `AtsEvaluationResponse`
  - `ListByResume(w, r)` — `GET /v1/resumes/{resumeID}/evaluations`
    - Lê `userID` do contexto e `resumeID` da URL
    - Chama `atsScoringServices.ListByResume(userID, resumeID)`
    - Retorna 200 com `[]AtsEvaluationSummaryResponse`
  - `GetByID(w, r)` — `GET /v1/resumes/{resumeID}/evaluations/{evaluationID}`
    - Lê `userID` do contexto, `resumeID` e `evaluationID` da URL
    - Chama `atsScoringServices.GetByID(userID, evaluationID)`
    - Retorna 200 com `AtsEvaluationResponse`

### Banco de Dados

Nova tabela `ats_evaluations`:

```sql
CREATE TABLE IF NOT EXISTS ats_evaluations (
    id TEXT PRIMARY KEY,
    resume_id TEXT NOT NULL,
    job_id TEXT NOT NULL,
    score REAL NOT NULL,
    summary TEXT NOT NULL,
    details TEXT NOT NULL,
    raw_response TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (resume_id) REFERENCES resumes(id),
    FOREIGN KEY (job_id) REFERENCES jobs(id)
);
```

---

## Configuração

Reutiliza as mesmas variáveis de ambiente do módulo de otimização:

- `GEMINI_API_KEY` — chave de API do Google AI Studio (obrigatória)
- `GEMINI_MODEL` — modelo a ser usado (opcional, default `gemini-2.0-flash`)

Nenhuma configuração adicional é necessária.

### Tratamento de Erros

| Situação | HTTP Status | Mensagem |
|---|---|---|
| Currículo não encontrado ou não pertence ao usuário | 404 | "currículo não encontrado" |
| Vaga não encontrada ou não pertence ao usuário | 404 | "vaga não encontrada" |
| Chave da API não configurada | 500 | "serviço de IA não configurado" |
| Erro na API do Google | 502 | "erro ao processar avaliação" |
| JSON inválido no body | 400 | "json inválido" |
| Campo `jobId` ausente | 400 | "jobId é obrigatório" |
| Sem token de autenticação | 401 | "token ausente" |

---

## Decisões Técnicas

- **Reuso do GeminiClient**: o mesmo cliente HTTP para API do Google AI Studio usado na otimização será reutilizado, alterando apenas o system prompt
- **Resposta estruturada da IA**: a IA deve retornar um JSON com `score`, `summary` e `details`, permitindo parse direto sem extração complexa
- **Score como `float64`**: permite valores como 7.5, garantindo precisão de uma casa decimal
- **Validação de score**: o serviço deve validar que o score retornado pela IA está entre 0 e 10 antes de persistir
- **Histórico preservado**: cada avaliação gera um novo registro, permitindo comparar diferentes versões do currículo contra a mesma vaga
- **Sem Details na listagem**: `AtsEvaluationSummaryResponse` exclui `Details` para respostas leves
- **Ownership via currículo**: a validação de ownership da avaliação é feita indiretamente — a avaliação pertence ao usuário se o currículo pai pertencer ao usuário

---

## Fluxo de avaliação

```
POST /v1/resumes/{resumeID}/evaluate (JSON: { "jobId": "..." })
  → AtsScoringHandler.Evaluate
    → AuthMiddleware (já garante userID no contexto)
    → decodifica body → extrai jobID
    → AtsScoringServices.Evaluate(userID, resumeID, jobID)
      → ResumeRepository.GetByID(resumeID) → valida ownership
      → JobRepository.GetByID(jobID) → valida ownership
      → monta system prompt ATS (constante)
      → monta user prompt (currículo + vaga)
      → GeminiClient.SendPrompt(systemPrompt, userPrompt)
        → POST /v1/models/{model}:generateContent?key={apiKey}
        → retorna JSON com score, summary, details
      → parseia JSON da resposta
      → valida score entre 0 e 10
      → AtsEvaluationRepository.Create(evaluation)
      → retorna AtsEvaluationResponse { id, resumeId, jobId, score, summary, details, createdAt }
```

---

## Dependências a adicionar

Nenhuma. Reutiliza dependências já existentes ou já especificadas (`GeminiClient`, `google/uuid`, `mattn/go-sqlite3`).
