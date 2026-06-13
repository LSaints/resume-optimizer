# TASKS.md

## Task 1

Criar entidade `AtsEvaluation`

### Objetivo

Criar a struct `AtsEvaluation` em `internal/domain/entities/ats_evaluation.go` com todos os campos necessários para representar uma avaliação ATS

### Validação

- Arquivo `internal/domain/entities/ats_evaluation.go` existe
- Struct possui campos: `ID uuid.UUID`, `ResumeID uuid.UUID`, `JobID uuid.UUID`, `Score float64`, `Summary string`, `Details string`, `RawResponse string`, `CreatedAt time.Time`
- Todos os campos têm tags JSON no formato `json:"camelCase"`
- O código compila sem erros

---

## Task 2

Criar tabela `ats_evaluations` no banco de dados

### Objetivo

Adicionar a criação da tabela `ats_evaluations` na função `createTables` em `internal/infrastructure/data/db.go`

### Validação

- A função `createTables` cria a tabela com as colunas: `id`, `resume_id`, `job_id`, `score`, `summary`, `details`, `raw_response`, `created_at`
- As chaves estrangeiras `resume_id` e `job_id` estão definidas
- A aplicação inicializa sem erros com a nova tabela

---

## Task 3

Criar DTOs de avaliação ATS

### Objetivo

Criar structs de request e response para as operações de avaliação ATS

### Validação

- `EvaluateResumeRequest` em `internal/application/requests/evaluate_request.go` com campo `JobID string` e tag `json:"jobId"`
- `AtsEvaluationResponse` em `internal/application/responses/evaluation_response.go` com campos: `ID`, `ResumeID`, `JobID`, `Score`, `Summary`, `Details`, `CreatedAt` e tags `json:"camelCase"`
- `AtsEvaluationSummaryResponse` no mesmo arquivo com campos: `ID`, `ResumeID`, `JobID`, `Score`, `Summary`, `CreatedAt` (sem `Details`)
- Structs seguem o padrão de nomenclatura e localização do projeto

---

## Task 4

Criar `AtsEvaluationRepository`

### Objetivo

Implementar `AtsEvaluationRepository` em `internal/infrastructure/repositories/ats_evaluation_repository.go` com métodos CRUD usando raw SQL

### Validação

- `NewAtsEvaluationRepository(db *sql.DB) *AtsEvaluationRepository` existe
- `Create(eval entities.AtsEvaluation) error` insere registro no banco
- `GetByID(id string) (entities.AtsEvaluation, error)` retorna avaliação por ID
- `GetByResumeID(resumeID string) ([]entities.AtsEvaluation, error)` retorna avaliações de um currículo ordenadas por `created_at DESC`
- `Delete(id string) error` remove registro do banco
- `GetByID` retorna erro se o ID não existir
- `GetByResumeID` retorna slice vazio se não houver avaliações

---

## Task 5

Criar system prompt e parser de resposta para scoring ATS

### Objetivo

Definir o system prompt para scoring ATS como constante e implementar o parser da resposta JSON da IA no serviço

### Validação

- System prompt ATS definido como constante em `internal/application/services/ats_scoring_services.go`
- Prompt instrui a IA a retornar JSON com `score`, `summary` e `details`
- Função `parseEvaluationResponse(raw string) (score float64, summary, details string, error)` implementada
- Parser trata resposta com e sem formatação markdown (código JSON)
- Parser retorna erro se `score` estiver fora do intervalo 0-10
- Parser retorna erro se JSON for inválido ou campos obrigatórios estiverem ausentes

---

## Task 6

Criar `AtsScoringServices`

### Objetivo

Implementar `AtsScoringServices` em `internal/application/services/ats_scoring_services.go` orquestrando validação, montagem de prompt, chamada à IA e persistência

### Validação

- `NewAtsScoringServices(evalRepo *AtsEvaluationRepository, resumeRepo *ResumeRepository, jobRepo *JobRepository, gemini *GeminiClient) *AtsScoringServices` existe
- `Evaluate(userID, resumeID, jobID string) (AtsEvaluationResponse, error)`:
  - Retorna erro "currículo não encontrado" se currículo não pertencer ao usuário
  - Retorna erro "vaga não encontrada" se vaga não pertencer ao usuário
  - Monta system prompt ATS + user prompt com texto do currículo e descrição da vaga
  - Chama `GeminiClient.SendPrompt`
  - Parseia resposta JSON e valida score entre 0 e 10
  - Persiste via `AtsEvaluationRepository.Create`
  - Retorna `AtsEvaluationResponse` com score, summary e details
- `ListByResume(userID, resumeID string) ([]AtsEvaluationSummaryResponse, error)`:
  - Verifica ownership do currículo antes de listar
  - Retorna lista resumida sem `Details`
- `GetByID(userID, evaluationID string) (AtsEvaluationResponse, error)`:
  - Retorna avaliação apenas se o currículo pai pertencer ao usuário
  - Retorna erro "avaliação não encontrada" se não encontrada

---

## Task 7

Criar `AtsScoringHandler` com endpoints

### Objetivo

Implementar `AtsScoringHandler` em `internal/apresentation/handlers/ats_scoring_handlers.go` com métodos para cada operação de avaliação ATS

### Validação

- `NewAtsScoringHandler(service *AtsScoringServices) *AtsScoringHandler` existe
- `Evaluate(w, r)` — lê `userID` do contexto, `resumeID` da URL via `r.PathValue("resumeID")`, decodifica JSON body, chama `service.Evaluate`, retorna 201
- `ListByResume(w, r)` — lê `userID` do contexto e `resumeID` da URL, chama `service.ListByResume`, retorna 200 com array JSON
- `GetByID(w, r)` — lê `userID` do contexto, `resumeID` e `evaluationID` da URL, chama `service.GetByID`, retorna 200
- Erros retornam mensagens em português com status HTTP adequado (400, 401, 404, 500, 502)
- Content-Type `application/json` é definido em todas as respostas

---

## Task 8

Registrar rotas de avaliação ATS e conectar dependências

### Objetivo

Atualizar `internal/apresentation/routes/routes.go` para instanciar e registrar todas as dependências do módulo de avaliação ATS

### Validação

- `AtsEvaluationRepository`, `AtsScoringServices` e `AtsScoringHandler` são instanciados em ordem
- Rotas são registradas protegidas pelo `AuthMiddleware`:
  - `POST /v1/resumes/{resumeID}/evaluate`
  - `GET /v1/resumes/{resumeID}/evaluations`
  - `GET /v1/resumes/{resumeID}/evaluations/{evaluationID}`
- A aplicação compila sem erros
- O servidor inicia e aceita requisições nas novas rotas

---

## Task 9

Testar fluxo completo de avaliação ATS

### Objetivo

Validar a feature ponta a ponta, incluindo cenários de sucesso e erro

### Validação

- Autenticar usuário, criar currículo e vaga, e avaliar → retorna 201 com `score`, `summary` e `details`
- Score retornado está entre 0 e 10
- Listar avaliações de um currículo → retorna lista sem `details` (200)
- Visualizar avaliação específica → retorna dados completos com `details` (200)
- Tentar avaliar com currículo de outro usuário → erro 404
- Tentar avaliar com vaga de outro usuário → erro 404
- Tentar avaliar sem `jobId` no body → erro 400
- Tentar avaliar com `resumeID` inexistente → erro 404
- Tentar avaliar com `jobID` inexistente → erro 404
- Tentar avaliar sem token → erro 401
- Tentar avaliar sem `GEMINI_API_KEY` configurada → erro 500
