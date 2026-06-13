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

O módulo de currículos seguirá a mesma estrutura em camadas do projeto, adicionando um serviço de extração de texto (`TextExtractor`) responsável por processar PDF e DOCX, e reutilizando a entidade `Resume` já existente em `internal/domain/entities/`.

---

## Estrutura Técnica

### Entidades

- `Resume` — já existe em `internal/domain/entities/resume.go`
  - `ID uuid.UUID`
  - `UserID uuid.UUID`
  - `OriginalName string`
  - `RawText string`
  - `UploadedAt time.Time`

### Serviços

- `ResumeServices` — `internal/application/services/resume_services.go`
  - `Create(userID, originalName string, file io.Reader) (ResumeResponse, error)`
  - `GetByID(userID, resumeID string) (ResumeResponse, error)`
  - `GetByUserID(userID string) ([]ResumeSummaryResponse, error)`
  - `Update(userID, resumeID, originalName string, file io.Reader) (ResumeResponse, error)`
  - `Delete(userID, resumeID string) error`
- `TextExtractor` — `internal/application/services/text_extractor.go`
  - `ExtractText(filename string, file io.Reader) (string, error)`
  - Detecta o formato pela extensão e aplica o parser adequado

### DTOs

- `CreateResumeRequest` — `multipart/form-data` com campo `file`
- `UpdateResumeRequest` — `multipart/form-data` com campo `file`
- `ResumeResponse` — `internal/application/responses/resume_response.go`
  - `ID, UserID, OriginalName, RawText, UploadedAt`
- `ResumeSummaryResponse` — `internal/application/responses/resume_response.go`
  - `ID, UserID, OriginalName, UploadedAt` (sem `RawText`)
  - Usado na listagem para evitar tráfego excessivo de dados

### Repositórios

- `ResumeRepository` — `internal/infrastructure/repositories/resume_repository.go`
  - `Create(resume entities.Resume) error`
  - `GetByID(id string) (entities.Resume, error)`
  - `GetByUserID(userID string) ([]entities.Resume, error)`
  - `Update(resume entities.Resume) error`
  - `Delete(id string) error`

### Endpoints

| Método | Rota | Autenticação | Descrição |
|---|---|---|---|
| POST | /v1/resumes | Protegida | Criar currículo (upload de arquivo) |
| GET | /v1/resumes | Protegida | Listar currículos do usuário |
| GET | /v1/resumes/{id} | Protegida | Obter currículo por ID |
| PUT | /v1/resumes/{id} | Protegida | Atualizar currículo (substituir arquivo) |
| DELETE | /v1/resumes/{id} | Protegida | Excluir currículo |

### Handlers

- `ResumeHandler` — `internal/apresentation/handlers/resume_handlers.go`
  - `Create(w, r)` — `POST /v1/resumes`
  - `List(w, r)` — `GET /v1/resumes`
  - `GetByID(w, r)` — `GET /v1/resumes/{id}`
  - `Update(w, r)` — `PUT /v1/resumes/{id}`
  - `Delete(w, r)` — `DELETE /v1/resumes/{id}`

### Banco de Dados

Nova tabela `resumes`:

```sql
CREATE TABLE IF NOT EXISTS resumes (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    original_name TEXT NOT NULL,
    raw_text TEXT NOT NULL,
    uploaded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

---

## Configuração

- Tamanho máximo de upload: 10 MB (definido como constante em `resume_handlers.go`)
- O limite é aplicado via `http.MaxBytesReader` no handler + `r.ParseMultipartForm`

---

## Decisões Técnicas

- **PDF extraction**: `github.com/ledongthuc/pdf` — biblioteca pura Go, simples, sem dependências externas, licença MIT
- **DOCX extraction**: `github.com/nguyenthenguyen/docx` — biblioteca pura Go para leitura de arquivos DOCX, extrai texto dos parágrafos
- **Upload via multipart/form-data**: uso do stdlib `r.ParseMultipartForm` + `r.FormFile("file")` para receber o arquivo
- **Validação de tipo**: verificação da extensão do nome original combinada com validação de magic bytes para maior segurança
- **Descarte do arquivo**: após extrair o texto, o `io.ReadCloser` é fechado e o arquivo não é persistido em disco ou banco
- **Isolamento por usuário**: todos os métodos de `ResumeRepository` filtram por `user_id` quando aplicável, garantindo que um usuário nunca acesse dados de outro
- **Sem RawText na listagem**: `ResumeSummaryResponse` exclui o campo `RawText` para evitar respostas grandes em listas com muitos currículos
- **Mensagens de erro em português** seguindo o padrão do projeto (ex: "formato de arquivo não suportado", "arquivo muito grande", "currículo não encontrado")

---

## Dependências a adicionar

- `github.com/ledongthuc/pdf`
- `github.com/nguyenthenguyen/docx`

---

## Fluxo de upload

```
POST /v1/resumes (multipart/form-data com campo "file")
  → ResumeHandler.Create
    → AuthMiddleware (já garante userID no contexto)
    → valida extensão do arquivo
    → valida magic bytes
    → http.MaxBytesReader limita a 10 MB
    → ResumeServices.Create
      → TextExtractor.ExtractText(filename, file) → rawText
      → ResumeRepository.Create(resume)
    → ResumeResponse { id, userId, originalName, rawText, uploadedAt }
```
