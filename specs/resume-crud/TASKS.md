# TASKS.md

## Task 1

Adicionar dependências de extração de texto

### Objetivo

Adicionar os pacotes `github.com/ledongthuc/pdf` e `github.com/nguyenthenguyen/docx` ao `go.mod`

### Validação

- `go mod tidy` executa sem erros
- Os pacotes constam no `go.sum`
- É possível importar ambos os pacotes em código Go

---

## Task 2

Criar tabela `resumes` no banco de dados

### Objetivo

Adicionar a criação da tabela `resumes` na função `createTables` em `internal/infrastructure/data/db.go`

### Validação

- A função `createTables` cria a tabela `resumes` com as colunas `id`, `user_id`, `original_name`, `raw_text`, `uploaded_at`
- A chave estrangeira `user_id` referencia `users(id)`
- A aplicação inicializa sem erros com a nova tabela

---

## Task 3

Criar ResumeRepository

### Objetivo

Implementar `ResumeRepository` em `internal/infrastructure/repositories/resume_repository.go` com métodos CRUD usando raw SQL

### Validação

- `NewResumeRepository(db *sql.DB) *ResumeRepository` existe
- Método `Create(resume entities.Resume) error` insere registro no banco
- Método `GetByID(id string) (entities.Resume, error)` retorna currículo por ID
- Método `GetByUserID(userID string) ([]entities.Resume, error)` retorna currículos do usuário
- Método `Update(resume entities.Resume) error` atualiza `original_name`, `raw_text` e `uploaded_at`
- Método `Delete(id string) error` remove registro do banco
- `GetByID` retorna erro se o ID não existir
- `GetByUserID` retorna slice vazio se usuário não tiver currículos

---

## Task 4

Criar TextExtractor para extração de texto de PDF e DOCX

### Objetivo

Implementar `TextExtractor` em `internal/application/services/text_extractor.go` capaz de extrair texto de arquivos PDF e DOCX

### Validação

- `NewTextExtractor() *TextExtractor` existe
- `ExtractText(filename string, file io.Reader) (string, error)` existe
- Extração de PDF retorna o texto contido no arquivo
- Extração de DOCX retorna o texto contido no arquivo
- Arquivo com extensão `.pdf` usa o parser PDF
- Arquivo com extensão `.docx` usa o parser DOCX
- Extensão não suportada retorna erro com mensagem em português
- O `io.Reader` é consumido corretamente (fecha ao final)

---

## Task 5

Criar DTOs de currículo

### Objetivo

Criar structs de request (multipart) e response para as operações de currículo

### Validação

- `CreateResumeRequest` e `UpdateResumeRequest` definidos como tipo multipart (sem struct, parâmetro `file` vindo do `r.FormFile`)
- `ResumeResponse` em `internal/application/responses/resume_response.go` possui campos: `ID`, `UserID`, `OriginalName`, `RawText`, `UploadedAt` com tags `json:"camelCase"`
- `ResumeSummaryResponse` no mesmo arquivo possui campos: `ID`, `UserID`, `OriginalName`, `UploadedAt` (sem `RawText`)
- Structs seguem o padrão de nomenclatura e localização do projeto

---

## Task 6

Criar ResumeServices

### Objetivo

Implementar `ResumeServices` em `internal/application/services/resume_services.go` orquestrando validação, extração de texto e persistência

### Validação

- `NewResumeServices(repo *ResumeRepository, extractor *TextExtractor) *ResumeServices` existe
- `Create(userID, originalName string, file io.Reader) (ResumeResponse, error)`:
  - Extrai texto do arquivo via `TextExtractor`
  - Persiste no banco via `ResumeRepository`
  - Retorna `ResumeResponse` com o texto extraído
  - Retorna erro se extração falhar
- `GetByID(userID, resumeID string) (ResumeResponse, error)`:
  - Retorna currículo apenas se o `userID` for o dono
  - Retorna erro se não encontrado ou se não pertence ao usuário
- `GetByUserID(userID string) ([]ResumeSummaryResponse, error)`:
  - Retorna lista resumida sem `RawText`
- `Update(userID, resumeID, originalName string, file io.Reader) (ResumeResponse, error)`:
  - Extrai novo texto, atualiza registro, retorna resposta completa
- `Delete(userID, resumeID string) error`:
  - Remove apenas se o currículo pertencer ao usuário

---

## Task 7

Criar ResumeHandler com endpoints

### Objetivo

Implementar `ResumeHandler` em `internal/apresentation/handlers/resume_handlers.go` com métodos para cada operação CRUD

### Validação

- `NewResumeHandler(service *ResumeServices) *ResumeHandler` existe
- `Create(w, r)` — lê `userID` do contexto, faz `r.ParseMultipartForm(10 << 20)`, obtém arquivo via `r.FormFile("file")`, valida extensão, chama `service.Create`
- `List(w, r)` — lê `userID` do contexto, chama `service.GetByUserID`, retorna array JSON
- `GetByID(w, r)` — lê `userID` do contexto e `id` da URL via `r.PathValue("id")`, chama `service.GetByID`
- `Update(w, r)` — lê `userID` do contexto e `id` da URL, faz `r.ParseMultipartForm`, chama `service.Update`
- `Delete(w, r)` — lê `userID` do contexto e `id` da URL, chama `service.Delete`
- Erros retornam mensagens em português com status HTTP adequado (400, 401, 404, 500)
- Content-Type `application/json` é definido em todas as respostas

---

## Task 8

Registrar rotas de currículo e conectar dependências

### Objetivo

Atualizar `internal/apresentation/routes/routes.go` para instanciar e registrar todas as dependências do módulo de currículos

### Validação

- `ResumeRepository`, `TextExtractor`, `ResumeServices` e `ResumeHandler` são instanciados em ordem
- Rotas são registradas protegidas pelo `AuthMiddleware`:
  - `POST /v1/resumes`
  - `GET /v1/resumes`
  - `GET /v1/resumes/{id}`
  - `PUT /v1/resumes/{id}`
  - `DELETE /v1/resumes/{id}`
- A aplicação compila sem erros
- O servidor inicia e aceita requisições nas novas rotas

---

## Task 9

Testar fluxo completo de CRUD de currículos

### Objetivo

Validar a feature ponta a ponta, incluindo cenários de sucesso e erro

### Validação

- Autenticar usuário e fazer upload de um PDF → texto extraído é retornado (200)
- Autenticar usuário e fazer upload de um DOCX → texto extraído é retornado (200)
- Listar currículos do usuário → retorna lista sem `rawText` (200)
- Visualizar currículo específico → retorna dados completos com `rawText` (200)
- Atualizar currículo com novo arquivo → texto antigo é substituído (200)
- Excluir currículo → registro removido (200), consulta posterior retorna (404)
- Tentar acessar currículo de outro usuário → erro (404 ou 403)
- Tentar upload de arquivo `.png` → erro (400)
- Tentar upload de arquivo maior que 10 MB → erro (400)
- Todas as operações sem token → erro (401)
