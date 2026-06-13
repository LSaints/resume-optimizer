# TASKS.md

## Task 1

Adicionar dependência de leitura de `.env`

### Objetivo

Adicionar o pacote `github.com/joho/godotenv` ao `go.mod` para carregar variáveis de ambiente do arquivo `.env` na inicialização do servidor

### Validação

- `go mod tidy` executa sem erros
- O pacote consta no `go.sum`

---

## Task 2

Criar arquivo `.env` com variáveis da API Google

### Objetivo

Criar o arquivo `backend/.env` com as variáveis `GEMINI_API_KEY` e `GEMINI_MODEL`, e adicionar `.env` ao `.gitignore` para não versionar a chave

### Validação

- Arquivo `backend/.env` existe com as variáveis documentadas
- `.gitignore` contém `.env` e o arquivo não aparece em `git status`
- A aplicação consegue ler `os.Getenv("GEMINI_API_KEY")` após carregar o `.env`

---

## Task 3

Atualizar `main.go` para carregar `.env` na inicialização

### Objetivo

No `cmd/api/main.go`, adicionar a chamada `godotenv.Load()` (ou leitura manual) antes de iniciar o servidor, para que as variáveis de ambiente estejam disponíveis

### Validação

- `godotenv.Load()` é chamado antes de `RegisterRoutes()`
- O servidor inicializa sem erros com o `.env` presente
- O servidor inicializa sem erros mesmo sem o `.env` (apenas avisa que não encontrou)

---

## Task 4

Enriquecer entidade `ResumeOptimized` com novos campos

### Objetivo

Adicionar os campos `JobID`, `SystemPrompt`, `UserPrompt` e `TypstContent` à entidade `ResumeOptimized` em `internal/domain/entities/resume_optimized.go`

### Validação

- `ResumeOptimized` possui campos: `ID`, `ResumeID`, `JobID`, `SystemPrompt`, `UserPrompt`, `RawText`, `TypstContent`, `CreatedAt`
- Todos os campos têm tipos e tags JSON corretos (`json:"camelCase"`)
- O código compila sem erros

---

## Task 5

Criar tabela `resumes_optimized` no banco de dados

### Objetivo

Adicionar a criação da tabela `resumes_optimized` na função `createTables` em `internal/infrastructure/data/db.go`

### Validação

- A função `createTables` cria a tabela com as colunas: `id`, `resume_id`, `job_id`, `system_prompt`, `user_prompt`, `raw_text`, `typst_content`, `created_at`
- As chaves estrangeiras `resume_id` e `job_id` estão definidas
- A aplicação inicializa sem erros com a nova tabela

---

## Task 6

Criar OptimizationRepository

### Objetivo

Implementar `OptimizationRepository` em `internal/infrastructure/repositories/optimization_repository.go` com métodos CRUD usando raw SQL

### Validação

- `NewOptimizationRepository(db *sql.DB) *OptimizationRepository` existe
- `Create(opt entities.ResumeOptimized) error` insere registro no banco
- `GetByID(id string) (entities.ResumeOptimized, error)` retorna otimização por ID
- `GetByResumeID(resumeID string) ([]entities.ResumeOptimized, error)` retorna otimizações de um currículo ordenadas por `created_at DESC`
- `Delete(id string) error` remove registro do banco
- `GetByID` retorna erro se o ID não existir
- `GetByResumeID` retorna slice vazio se não houver otimizações

---

## Task 7

Criar DTOs de otimização

### Objetivo

Criar structs de request e response para as operações de otimização

### Validação

- `OptimizeResumeRequest` em `internal/application/requests/optimize_request.go` com campo `JobID string` e tag `json:"jobId"`
- `OptimizeResponse` em `internal/application/responses/optimize_response.go` com campos: `ID`, `ResumeID`, `JobID`, `TypstContent`, `CreatedAt` e tags `json:"camelCase"`
- `OptimizeSummaryResponse` no mesmo arquivo com campos: `ID`, `ResumeID`, `JobID`, `CreatedAt` (sem `TypstContent`)
- Structs seguem o padrão de nomenclatura e localização do projeto

---

## Task 8

Criar GeminiClient para comunicação com a API Google AI Studio

### Objetivo

Implementar `GeminiClient` em `internal/application/services/gemini_client.go` responsável por fazer a requisição HTTP para a API do Google AI Studio

### Validação

- `NewGeminiClient() *GeminiClient` existe
- `SendPrompt(systemPrompt, userPrompt string) (string, error)` existe
- Monta requisição HTTP para `https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent?key={apiKey}`
- Lê `GEMINI_API_KEY` e `GEMINI_MODEL` das variáveis de ambiente
- Usa timeout de 60 segundos via `context.WithTimeout`
- Retorna erro "serviço de IA não configurado" se `GEMINI_API_KEY` estiver vazia
- Retorna erro se a API retornar status code diferente de 200
- Retorna o texto gerado pela IA em caso de sucesso

---

## Task 9

Criar OptimizationServices

### Objetivo

Implementar `OptimizationServices` em `internal/application/services/optimization_services.go` orquestrando validação, montagem de prompt, chamada à IA e persistência

### Validação

- `NewOptimizationServices(optRepo *OptimizationRepository, resumeRepo *ResumeRepository, jobRepo *JobRepository, gemini *GeminiClient) *OptimizationServices` existe
- `Optimize(userID, resumeID, jobID string) (OptimizeResponse, error)`:
  - Busca currículo e valida ownership → erro "currículo não encontrado" se não pertencer ao usuário
  - Busca vaga e valida ownership → erro "vaga não encontrada" se não pertencer ao usuário
  - Monta system prompt (constante) + user prompt com texto do currículo e descrição da vaga
  - Chama `GeminiClient.SendPrompt`
  - Extrai bloco Typst da resposta (remove ```typst e ``` se presentes)
  - Persiste via `OptimizationRepository.Create`
  - Retorna `OptimizeResponse` com o Typst gerado
- `GetByResumeID(userID, resumeID string) ([]OptimizeSummaryResponse, error)`:
  - Retorna lista resumida sem `TypstContent`
- `GetByID(userID, optimizationID string) (OptimizeResponse, error)`:
  - Retorna otimização apenas se o currículo pai pertencer ao usuário
  - Erro "otimização não encontrada" se não encontrada

---

## Task 10

Criar OptimizationHandler com endpoints

### Objetivo

Implementar `OptimizationHandler` em `internal/apresentation/handlers/optimization_handlers.go` com métodos para cada operação de otimização

### Validação

- `NewOptimizationHandler(service *OptimizationServices) *OptimizationHandler` existe
- `Optimize(w, r)` — lê `userID` do contexto, `resumeID` da URL via `r.PathValue("resumeID")`, decodifica JSON body, chama `service.Optimize`, retorna 201
- `ListByResume(w, r)` — lê `userID` do contexto e `resumeID` da URL, chama `service.GetByResumeID`, retorna 200 com array JSON
- `GetByID(w, r)` — lê `userID` do contexto, `resumeID` e `optimizationID` da URL, chama `service.GetByID`, retorna 200
- Erros retornam mensagens em português com status HTTP adequado (400, 401, 404, 500, 502)
- Content-Type `application/json` é definido em todas as respostas

---

## Task 11

Registrar rotas de otimização e conectar dependências

### Objetivo

Atualizar `internal/apresentation/routes/routes.go` para instanciar e registrar todas as dependências do módulo de otimização

### Validação

- `OptimizationRepository`, `GeminiClient`, `OptimizationServices` e `OptimizationHandler` são instanciados em ordem
- Rotas são registradas protegidas pelo `AuthMiddleware`:
  - `POST /v1/resumes/{resumeID}/optimize`
  - `GET /v1/resumes/{resumeID}/optimizations`
  - `GET /v1/resumes/{resumeID}/optimizations/{optimizationID}`
- A aplicação compila sem erros
- O servidor inicia e aceita requisições nas novas rotas

---

## Task 12

Testar fluxo completo de otimização de currículo

### Objetivo

Validar a feature ponta a ponta, incluindo cenários de sucesso e erro

### Validação

- Autenticar usuário, criar currículo e vaga, e otimizar → retorna 201 com conteúdo Typst
- Listar otimizações de um currículo → retorna lista sem `typstContent` (200)
- Visualizar otimização específica → retorna dados completos com `typstContent` (200)
- Tentar otimizar com currículo de outro usuário → erro 404
- Tentar otimizar com vaga de outro usuário → erro 404
- Tentar otimizar sem `jobId` no body → erro 400
- Tentar otimizar com `resumeID` inexistente → erro 404
- Tentar otimizar com `jobID` inexistente → erro 404
- Tentar otimizar sem token → erro 401
- Tentar otimizar sem `GEMINI_API_KEY` configurada → erro 500
