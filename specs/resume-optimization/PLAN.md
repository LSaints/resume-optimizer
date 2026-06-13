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

---

## Arquitetura

O módulo de otimização seguirá a mesma estrutura em camadas do projeto. A entidade `ResumeOptimized` já existe em `internal/domain/entities/` e será enriquecida com novos campos. Um novo serviço `OptimizationServices` será responsável por montar o prompt, chamar a API do Google AI Studio e processar a resposta. A comunicação com a API externa será feita via `net/http` padrão, sem dependências adicionais.

---

## Estrutura Técnica

### Entidades

- `ResumeOptimized` — `internal/domain/entities/resume_optimized.go` (será enriquecida)
  - `ID uuid.UUID` — identificador único
  - `ResumeID uuid.UUID` — FK para o currículo original
  - `JobID uuid.UUID` — FK para a vaga usada na otimização
  - `SystemPrompt string` — system prompt utilizado
  - `UserPrompt string` — prompt montado com currículo + vaga
  - `RawText string` — texto original retornado pela IA (resposta bruta)
  - `TypstContent string` — conteúdo Typst limpo extraído da resposta
  - `CreatedAt time.Time`

### Serviços

- `OptimizationServices` — `internal/application/services/optimization_services.go`
  - `Optimize(userID, resumeID, jobID string) (OptimizeResponse, error)`
    - Busca currículo e vaga no banco (validando ownership)
    - Monta system prompt + user prompt
    - Chama API do Google AI Studio (Gemini)
    - Extrai o conteúdo Typst da resposta
    - Persiste o resultado em `resumes_optimized`
    - Retorna resposta com o Typst gerado
  - `GetByResumeID(userID, resumeID string) ([]OptimizeResponse, error)`
    - Lista histórico de otimizações de um currículo
  - `GetByID(userID, optimizationID string) (OptimizeResponse, error)`
    - Retorna uma otimização específica

- `GeminiClient` — `internal/application/services/gemini_client.go`
  - `SendPrompt(systemPrompt, userPrompt string) (string, error)`
    - Monta requisição HTTP para a API Google AI Studio
    - Define headers e corpo JSON
    - Processa resposta e extrai texto gerado
    - Lê `GEMINI_API_KEY` de variável de ambiente
    - Usa modelo `gemini-2.0-flash` por padrão
    - Timeout de 60 segundos

### System Prompt

O system prompt será definido como uma constante no serviço de otimização:

```
Você é um especialista em otimização de currículos com vasto conhecimento em recrutamento e seleção, ATS (Applicant Tracking Systems) e mercado de trabalho.

Sua função é reescrever currículos para maximizar a compatibilidade com a vaga desejada, respeitando o nível de senioridade e exigência da vaga.

Regras:
1. Analise o currículo original e a descrição da vaga fornecidos
2. Identifique o nível ATS da vaga (entry-level, mid-level, senior, expert)
3. Reestruture o currículo em linguagem Typst, organizando seções de forma profissional
4. Destaque palavras-chave da vaga no currículo
5. Use linguagem e profundidade condizentes com o nível ATS identificado
6. Mantenha a veracidade das informações — nunca invente experiências ou habilidades
7. Priorize realizações mensuráveis e resultados concretos
8. Otimize o formato para ser legível tanto por humanos quanto por sistemas ATS

Retorne APENAS o código Typst, sem explicações adicionais.
```

### DTOs

- `OptimizeResumeRequest` — `internal/application/requests/optimize_request.go`
  - `JobID string` `json:"jobId"` — ID da vaga para basear a otimização
- `OptimizeResponse` — `internal/application/responses/optimize_response.go`
  - `ID string` `json:"id"`
  - `ResumeID string` `json:"resumeId"`
  - `JobID string` `json:"jobId"`
  - `TypstContent string` `json:"typstContent"`
  - `CreatedAt string` `json:"createdAt"`
- `OptimizeSummaryResponse` — mesmo arquivo
  - `ID string` `json:"id"`
  - `ResumeID string` `json:"resumeId"`
  - `JobID string` `json:"jobId"`
  - `CreatedAt string` `json:"createdAt"`
  - Usado na listagem (sem `TypstContent`)

### Repositórios

- `OptimizationRepository` — `internal/infrastructure/repositories/optimization_repository.go`
  - `Create(opt entities.ResumeOptimized) error`
  - `GetByID(id string) (entities.ResumeOptimized, error)`
  - `GetByResumeID(resumeID string) ([]entities.ResumeOptimized, error)` — ordenado por `created_at DESC`
  - `Delete(id string) error`

### Endpoints

| Método | Rota | Autenticação | Descrição |
|---|---|---|---|
| POST | /v1/resumes/{resumeID}/optimize | Protegida | Otimizar currículo com base em uma vaga |
| GET | /v1/resumes/{resumeID}/optimizations | Protegida | Listar histórico de otimizações do currículo |
| GET | /v1/resumes/{resumeID}/optimizations/{optimizationID} | Protegida | Obter otimização específica |

### Handlers

- `OptimizationHandler` — `internal/apresentation/handlers/optimization_handlers.go`
  - `Optimize(w, r)` — `POST /v1/resumes/{resumeID}/optimize`
    - Lê `resumeID` da URL via `r.PathValue("resumeID")`
    - Decodifica body JSON com `JobID`
    - Chama `optimizationServices.Optimize(userID, resumeID, jobID)`
    - Retorna 201 Created com `OptimizeResponse`
  - `ListByResume(w, r)` — `GET /v1/resumes/{resumeID}/optimizations`
    - Lê `resumeID` da URL
    - Chama `optimizationServices.GetByResumeID(userID, resumeID)`
    - Retorna 200 com `[]OptimizeSummaryResponse`
  - `GetByID(w, r)` — `GET /v1/resumes/{resumeID}/optimizations/{optimizationID}`
    - Lê `resumeID` e `optimizationID` da URL
    - Chama `optimizationServices.GetByID(userID, optimizationID)`
    - Retorna 200 com `OptimizeResponse`

### Banco de Dados

Nova tabela `resumes_optimized`:

```sql
CREATE TABLE IF NOT EXISTS resumes_optimized (
    id TEXT PRIMARY KEY,
    resume_id TEXT NOT NULL,
    job_id TEXT NOT NULL,
    system_prompt TEXT NOT NULL,
    user_prompt TEXT NOT NULL,
    raw_text TEXT NOT NULL,
    typst_content TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (resume_id) REFERENCES resumes(id),
    FOREIGN KEY (job_id) REFERENCES jobs(id)
);
```

---

## Configuração

### Variáveis de Ambiente

Arquivo `.env` na raiz do projeto `backend/`:

```env
GEMINI_API_KEY=your-google-ai-studio-api-key
GEMINI_MODEL=gemini-2.0-flash
```

- `GEMINI_API_KEY` — chave de API do Google AI Studio (obrigatória)
- `GEMINI_MODEL` — modelo a ser usado (opcional, default `gemini-2.0-flash`)

### Leitura de .env

Adicionar biblioteca `github.com/joho/godotenv` para carregar `.env` automaticamente no `main.go`, ou implementar leitura manual via `os.ReadFile` + parser simples.

### Tratamento de Erros

| Situação | HTTP Status | Mensagem |
|---|---|---|
| Currículo não encontrado ou não pertence ao usuário | 404 | "currículo não encontrado" |
| Vaga não encontrada ou não pertence ao usuário | 404 | "vaga não encontrada" |
| Chave da API não configurada | 500 | "serviço de IA não configurado" |
| Erro na API do Google | 502 | "erro ao processar otimização" |
| JSON inválido no body | 400 | "json inválido" |
| Sem token de autenticação | 401 | "token ausente" |

---

## Decisões Técnicas

- **API Google AI Studio**: chamada HTTP direta via `net/http` para `https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent?key={apiKey}`, sem SDK externo
- **Modelo padrão**: `gemini-2.0-flash` por ser rápido e com boa qualidade para tasks de texto
- **System prompt versionado**: definido como constante no código, facilitando manutenção e versionamento
- **Extração do Typst**: a resposta da IA pode vir com markdown ```typst ... ```, o serviço deve extrair apenas o bloco Typst
- **Timeout**: 60 segundos para a chamada externa, via `context.WithTimeout`
- **Isolamento por usuário**: validação de ownership de currículo e vaga antes de processar
- **Histórico preservado**: cada otimização gera um novo registro, permitindo comparação entre versões
- **Sem RawText na listagem**: `OptimizeSummaryResponse` exclui `TypstContent` para respostas leves

---

## Fluxo de otimização

```
POST /v1/resumes/{resumeID}/optimize (JSON: { "jobId": "..." })
  → OptimizationHandler.Optimize
    → AuthMiddleware (já garante userID no contexto)
    → decodifica body → extrai jobID
    → OptimizationServices.Optimize(userID, resumeID, jobID)
      → ResumeRepository.GetByID(resumeID) → valida ownership
      → JobRepository.GetByID(jobID) → valida ownership
      → monta system prompt (constante)
      → monta user prompt (currículo + vaga)
      → GeminiClient.SendPrompt(systemPrompt, userPrompt)
        → GET /v1/models/{model}:generateContent?key={apiKey}
        → retorna texto gerado
      → extrai bloco Typst da resposta
      → OptimizationRepository.Create(resumeOptimized)
      → retorna OptimizeResponse { id, resumeId, jobId, typstContent, createdAt }
```

---

## Dependências a adicionar

- `github.com/joho/godotenv` (ou leitura manual de `.env`)
```
