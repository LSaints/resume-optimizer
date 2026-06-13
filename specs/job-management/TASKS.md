# TASKS.md

## Task 1

Atualizar entidade `Job` com campos `UserID` e `UpdatedAt`

### Objetivo

Modificar `internal/domain/entities/job.go` para adicionar `UserID uuid.UUID` e `UpdatedAt time.Time`

### Validação

- `Job` possui campos: `ID`, `UserID`, `Title`, `RawDescription`, `CreatedAt`, `UpdatedAt`
- Todos os campos possuem tags `json:"camelCase"`
- A struct compila sem erros

---

## Task 2

Criar/migrar tabela `jobs` no banco de dados

### Objetivo

Adicionar a criação da tabela `jobs` na função `createTables` em `internal/infrastructure/data/db.go` com as colunas `id`, `user_id`, `title`, `raw_description`, `created_at`, `updated_at`

### Validação

- A função `createTables` cria a tabela `jobs` com todas as colunas necessárias
- A chave estrangeira `user_id` referencia `users(id)`
- A aplicação inicializa sem erros com a nova tabela

---

## Task 3

Criar DTOs de vaga

### Objetivo

Criar `CreateJobRequest`, `UpdateJobRequest` em `internal/application/requests/job_requests.go` e `JobResponse` em `internal/application/responses/job_response.go`

### Validação

- `CreateJobRequest` possui campos `Title` e `RawDescription` com tags `json:"camelCase"`
- `UpdateJobRequest` possui campos `Title` e `RawDescription` com tags `json:"camelCase"`
- `JobResponse` possui campos `ID`, `Title`, `RawDescription`, `CreatedAt`, `UpdatedAt` com tags `json:"camelCase"`
- Structs seguem o padrão de nomenclatura e localização do projeto

---

## Task 4

Criar JobRepository

### Objetivo

Implementar `JobRepository` em `internal/infrastructure/repositories/job_repository.go` com métodos CRUD usando raw SQL

### Validação

- `NewJobRepository(db *sql.DB) *JobRepository` existe
- `Create(job entities.Job) error` insere registro no banco
- `GetByID(id string) (entities.Job, error)` retorna vaga por ID
- `GetByUserID(userID string) ([]entities.Job, error)` retorna vagas do usuário
- `Update(job entities.Job) error` atualiza `title`, `raw_description` e `updated_at`
- `Delete(id string) error` remove registro do banco
- `GetByID` retorna erro se o ID não existir
- `GetByUserID` retorna slice vazio se usuário não tiver vagas

---

## Task 5

Criar JobServices com validação e ownership

### Objetivo

Implementar `JobServices` em `internal/application/services/job_services.go` orquestrando validação e persistência

### Validação

- `NewJobServices(repo *JobRepository) *JobServices` existe
- `Create(userID string, request CreateJobRequest) (JobResponse, error)`:
  - Valida título e descrição não vazios
  - Cria entidade com UUID, UserID e timestamps
  - Persiste e retorna `JobResponse`
  - Retorna erro se campos obrigatórios estiverem vazios
- `GetByID(userID, jobID string) (JobResponse, error)`:
  - Retorna vaga apenas se o `userID` for o dono
  - Retorna erro se não encontrada ou se não pertence ao usuário
- `GetByUserID(userID string) ([]JobResponse, error)`:
  - Retorna lista de vagas do usuário
- `Update(userID, jobID string, request UpdateJobRequest) (JobResponse, error)`:
  - Atualiza título, descrição e `updated_at`
  - Retorna erro se vaga não pertencer ao usuário
- `Delete(userID, jobID string) error`:
  - Remove apenas se a vaga pertencer ao usuário

---

## Task 6

Criar JobHandler com endpoints

### Objetivo

Implementar `JobHandler` em `internal/apresentation/handlers/job_handlers.go` com métodos para cada operação CRUD

### Validação

- `NewJobHandler(service *JobServices) *JobHandler` existe
- `Create(w, r)` — lê `userID` do contexto, decodifica JSON, valida, chama `service.Create`, retorna 201
- `List(w, r)` — lê `userID` do contexto, chama `service.GetByUserID`, retorna array JSON
- `GetByID(w, r)` — lê `userID` do contexto e `id` da URL via `r.PathValue("id")`, chama `service.GetByID`
- `Update(w, r)` — lê `userID` do contexto e `id` da URL, decodifica JSON, chama `service.Update`
- `Delete(w, r)` — lê `userID` do contexto e `id` da URL, chama `service.Delete`, retorna 204
- Erros retornam mensagens em português com status HTTP adequado (400, 404, 500)
- Content-Type `application/json` é definido em todas as respostas

---

## Task 7

Registrar rotas de vagas e conectar dependências

### Objetivo

Atualizar `internal/apresentation/routes/routes.go` para instanciar e registrar todas as dependências do módulo de vagas

### Validação

- `JobRepository`, `JobServices` e `JobHandler` são instanciados em ordem
- Rotas são registradas protegidas pelo `AuthMiddleware`:
  - `POST /v1/jobs`
  - `GET /v1/jobs`
  - `GET /v1/jobs/{id}`
  - `PUT /v1/jobs/{id}`
  - `DELETE /v1/jobs/{id}`
- A aplicação compila sem erros
- O servidor inicia e aceita requisições nas novas rotas

---

## Task 8

Testar fluxo completo de CRUD de vagas

### Objetivo

Validar a feature ponta a ponta, incluindo cenários de sucesso e erro

### Validação

- Autenticar usuário e criar vaga com título e descrição → retorna dados completos (201)
- Listar vagas do usuário → retorna lista com vagas cadastradas (200)
- Visualizar vaga específica → retorna dados completos (200)
- Atualizar título e descrição de vaga existente → dados atualizados (200)
- Excluir vaga → registro removido (204), consulta posterior retorna (404)
- Tentar criar vaga com título vazio → erro (400)
- Tentar criar vaga com descrição vazia → erro (400)
- Tentar acessar vaga de outro usuário → erro (404)
- Todas as operações sem token → erro (401)
