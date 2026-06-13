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
- Ownership verificado no service layer comparando `userID`

---

## Arquitetura

O módulo de vagas seguirá a mesma estrutura em camadas do projeto, reaproveitando a entidade `Job` existente em `internal/domain/entities/job.go` com a adição do campo `UserID` para garantir isolamento por usuário.

---

## Estrutura Técnica

### Entidades

- `Job` — `internal/domain/entities/job.go` (modificar)
  - `ID uuid.UUID`
  - `UserID uuid.UUID` (novo campo, obrigatório)
  - `Title string`
  - `RawDescription string`
  - `CreatedAt time.Time`
  - `UpdatedAt time.Time` (novo campo, opcional, usado no update)

### Serviços

- `JobServices` — `internal/application/services/job_services.go`
  - `Create(userID string, request CreateJobRequest) (JobResponse, error)`
  - `GetByID(userID, jobID string) (JobResponse, error)`
  - `GetByUserID(userID string) ([]JobResponse, error)`
  - `Update(userID, jobID string, request UpdateJobRequest) (JobResponse, error)`
  - `Delete(userID, jobID string) error`

### DTOs

- `CreateJobRequest` — `internal/application/requests/job_requests.go`
  - `Title string` (obrigatório)
  - `RawDescription string` (obrigatório)
- `UpdateJobRequest` — `internal/application/requests/job_requests.go`
  - `Title string` (obrigatório)
  - `RawDescription string` (obrigatório)
- `JobResponse` — `internal/application/responses/job_response.go`
  - `ID uuid.UUID`
  - `Title string`
  - `RawDescription string`
  - `CreatedAt time.Time`
  - `UpdatedAt time.Time`

### Repositórios

- `JobRepository` — `internal/infrastructure/repositories/job_repository.go`
  - `Create(job entities.Job) error`
  - `GetByID(id string) (entities.Job, error)`
  - `GetByUserID(userID string) ([]entities.Job, error)`
  - `Update(job entities.Job) error`
  - `Delete(id string) error`

### Endpoints

| Método | Rota | Autenticação | Descrição |
|--------|------|-------------|----------|
| POST | /v1/jobs | Protegida | Criar vaga |
| GET | /v1/jobs | Protegida | Listar vagas do usuário |
| GET | /v1/jobs/{id} | Protegida | Obter vaga por ID |
| PUT | /v1/jobs/{id} | Protegida | Atualizar vaga |
| DELETE | /v1/jobs/{id} | Protegida | Excluir vaga |

### Handlers

- `JobHandler` — `internal/apresentation/handlers/job_handlers.go`
  - `Create(w, r)` — `POST /v1/jobs`
  - `List(w, r)` — `GET /v1/jobs`
  - `GetByID(w, r)` — `GET /v1/jobs/{id}`
  - `Update(w, r)` — `PUT /v1/jobs/{id}`
  - `Delete(w, r)` — `DELETE /v1/jobs/{id}`

### Banco de Dados

Modificar a criação da tabela `jobs` (ou criá-la caso não exista):

```sql
CREATE TABLE IF NOT EXISTS jobs (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    title TEXT NOT NULL,
    raw_description TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

---

## Configuração

- Nenhuma configuração adicional necessária além das existentes
- Validação de campos obrigatórios no service layer antes de persistir

---

## Decisões Técnicas

- **Tabela `jobs` já pode existir** — usar `CREATE TABLE IF NOT EXISTS` com as novas colunas; se a tabela já existir sem `user_id` e `updated_at`, uma migração será necessária
- **JSON puro (sem multipart)** — diferente de currículos, vagas são texto puro, então usam `Content-Type: application/json` com DTOs tradicionais
- **Ownership no service** — mesmo padrão de `ResumeServices`: verificar se `job.UserID.String() == userID` em todos os métodos que acessam recurso único
- **Validação no service** — títulos e descrições vazias são rejeitadas com erro em português antes de chamar o repositório
- **UpdatedAt na resposta** — reflete a data da última alteração, atualizado pelo service no momento do update
- **Mensagens de erro em português** — "título é obrigatório", "descrição é obrigatória", "vaga não encontrada"
- **Sem listagem resumida** — diferente de currículos, vagas têm payload pequeno e não precisam de DTO separado para listagem

---

## Fluxo de criação

```
POST /v1/jobs (application/json { "title": "...", "rawDescription": "..." })
  → JobHandler.Create
    → AuthMiddleware (já garante userID no contexto)
    → decodifica JSON para CreateJobRequest
    → JobServices.Create(userID, request)
      → valida título e descrição não vazios
      → constrói entidade Job com UserID, timestamps
      → JobRepository.Create(job)
    → JobResponse { id, title, rawDescription, createdAt, updatedAt }
```
