# PLAN.md

## Existing Context

### Architecture

Layered Architecture (Clean Architecture)

```
cmd/api/
internal/
  apresentation/   -> HTTP layer (handlers, routes)
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

---

## Arquitetura

A autenticação JWT será implementada como uma camada de middleware sobre o `net/http` existente, seguindo o padrão de handlers do Go. Um serviço de autenticação será responsável por gerar e validar tokens. O hash de senha será integrado ao fluxo existente de criação de usuários e verificação de login.

---

## Estrutura Técnica

### Entidades

Nenhuma nova entidade. A entidade `User` existente será modificada para suportar hash de senha.

### Serviços

- `AuthServices` — localizado em `internal/application/services/auth_services.go`
  - `Login(email, password)` → valida credenciais e retorna token JWT
  - `ValidateToken(tokenString)` → valida token e retorna claims

### DTOs

- `LoginRequest` — `internal/application/requests/login_request.go`
  - `Email string json:"email"`
  - `Password string json:"password"`
- `LoginResponse` — `internal/application/responses/login_response.go`
  - `Token string json:"token"`
  - `ExpiresAt string json:"expiresAt"`

### Handlers

- `AuthHandler` — `internal/apresentation/handlers/auth_handlers.go`
  - `Login(w, r)` — POST /v1/auth/login

### Middleware

- `AuthMiddleware` — `internal/apresentation/middleware/auth_middleware.go`
  - Extrai token do header `Authorization: Bearer <token>`
  - Valida o token via `AuthServices.ValidateToken`
  - Injeta claims no contexto da requisição
  - Retorna 401 se token ausente, inválido ou expirado

### Rotas

- `POST /v1/auth/login` — pública
- Demais endpoints `/v1/users/*` — protegidos pelo middleware

### Modificações em entidades existentes

- `User.Password` passará a armazenar hash bcrypt em vez de texto puro
- `CreateUserRequest` permanece igual (recebe senha em texto puro)
- `UserRepository.CreateUser` deve aplicar hash antes de persistir

### Repositórios

Nenhum novo repositório. O `UserRepository` existente será estendido com:
- `GetByEmail(email)` — buscar usuário por email para autenticação

---

## Configuração

- Chave secreta JWT definida via variável de ambiente `JWT_SECRET`
- Tempo de expiração do token: 24 horas
- Custo do bcrypt: 10 (default)

---

## Decisões Técnicas

- **JWT library**: `github.com/golang-jwt/jwt/v5` — biblioteca padrão para JWT em Go
- **Bcrypt library**: `golang.org/x/crypto/bcrypt` — pacote oficial para hash de senhas
- **Sem interface**: manter consistência com o código existente, usando structs concretas
- **Claims no contexto**: usar `context.WithValue` para disponibilizar `userID` nos handlers protegidos
- **Middleware em cadeia**: usar função `AuthMiddleware(next http.Handler) http.Handler` compatível com `net/http`
- **Hash no repositório**: o hash será aplicado no service (`UserServices`), não no repositório, mantendo a separação de responsabilidades

---

## Dependências a adicionar

- `github.com/golang-jwt/jwt/v5`
- `golang.org/x/crypto` (bcrypt)

---

## Fluxo de autenticação

```
POST /v1/auth/login { email, password }
  → AuthHandler.Login
  → AuthServices.Login
    → UserRepository.GetByEmail(email)
    → bcrypt.CompareHashAndPassword(hash, password)
    → jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
  → LoginResponse { token, expiresAt }

GET /v1/users (com Authorization: Bearer <token>)
  → AuthMiddleware
    → extrai token do header
    → AuthServices.ValidateToken(token)
    → injeta userID no context
  → UserHandler.GetUsers (lê userID do context)
```
