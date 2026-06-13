# TASKS.md

## Task 1

Adicionar dependências externas

### Objetivo

Adicionar os pacotes `golang-jwt/jwt/v5` e `golang.org/x/crypto` ao `go.mod`

### Validação

- `go mod tidy` executa sem erros
- `go mod vendor` ou verificação de que as dependências constam no `go.sum`
- Os pacotes podem ser importados em código Go

---

## Task 2

Criar DTOs de autenticação

### Objetivo

Criar `LoginRequest` e `LoginResponse` seguindo os padrões de nomenclatura e localização dos DTOs existentes

### Validação

- Arquivo `internal/application/requests/login_request.go` existe com `Email` e `Password`
- Arquivo `internal/application/responses/login_response.go` existe com `Token` e `ExpiresAt`
- Structs seguem o padrão de tags `json:"camelCase"`

---

## Task 3

Implementar busca de usuário por email no repositório

### Objetivo

Adicionar método `GetByEmail(email string)` ao `UserRepository` para consultar usuário pelo email

### Validação

- Método `GetByEmail(email string) (*entities.User, error)` existe em `UserRepository`
- Retorna `nil, erro` se email não encontrado
- Retorna `*User` preenchido se email existe
- Query SQL usa `WHERE email = ?`

---

## Task 4

Aplicar hash bcrypt na criação de usuário

### Objetivo

Modificar `UserServices.CreateUser` para aplicar hash bcrypt na senha antes de enviar ao repositório. A senha em texto puro não deve mais ser armazenada no banco.

### Validação

- `UserServices.CreateUser` importa `golang.org/x/crypto/bcrypt`
- A senha armazenada no banco é um hash, não o texto puro
- O hash gerado é diferente para cada chamada (salt aleatório)
- Usuários existentes criados antes dessa mudança continuam funcionando (hash é aplicado apenas em novos cadastros)

---

## Task 5

Criar AuthServices com geração e validação de token JWT

### Objetivo

Implementar `AuthServices` com métodos `Login(email, password)` e `ValidateToken(tokenString)`. O login deve buscar o usuário, comparar a senha com bcrypt e gerar um token JWT assinado com HS256.

### Validação

- `AuthServices` injeta `UserRepository` via construtor
- `Login` com email/senha corretos retorna `(LoginResponse, nil)`
- `Login` com email inexistente retorna erro
- `Login` com senha incorreta retorna erro
- `ValidateToken` com token válido retorna claims contendo `userID`
- `ValidateToken` com token inválido retorna erro
- `ValidateToken` com token expirado retorna erro
- `ValidateToken` com token de algoritmo diferente retorna erro

---

## Task 6

Criar AuthHandler com endpoint de login

### Objetivo

Implementar `AuthHandler` com método `Login(w, r)` que aceita `POST /v1/auth/login`, decodifica o corpo, chama `AuthServices.Login` e retorna o token.

### Validação

- `POST /v1/auth/login` com `{ "email": "valido@email.com", "password": "senha" }` retorna `200` com `{ "token": "...", "expiresAt": "..." }`
- `POST /v1/auth/login` com email inexistente retorna `401`
- `POST /v1/auth/login` com senha incorreta retorna `401`
- `POST /v1/auth/login` com corpo inválido retorna `400`

---

## Task 7

Criar middleware de autenticação JWT

### Objetivo

Implementar `AuthMiddleware` que extrai o token do header `Authorization: Bearer <token>`, valida via `AuthServices.ValidateToken` e injeta o `userID` no contexto da requisição.

### Validação

- Requisição sem header `Authorization` retorna `401`
- Requisição com `Authorization: Bearer token_invalido` retorna `401`
- Requisição com `Authorization: Bearer token_expirado` retorna `401`
- Requisição com `Authorization: Bearer token_valido` passa adiante com `userID` no contexto
- Requisição com `Authorization: Bearer` (sem token) retorna `401`
- Requisição com `Authorization: Basic ...` (Bearer ausente) retorna `401`

---

## Task 8

Registrar rotas e conectar middleware

### Objetivo

Atualizar `routes.go` para:
- Registrar `POST /v1/auth/login` como rota pública
- Aplicar `AuthMiddleware` nas rotas protegidas existentes (`GET /v1/users`, `GET /v1/users/{id}`, `POST /v1/users`)
- A injeção de dependências deve incluir `AuthServices` e `AuthHandler`

### Validação

- `POST /v1/auth/login` funciona sem token
- `GET /v1/users` sem token retorna `401`
- `GET /v1/users/{id}` sem token retorna `401`
- `POST /v1/users` sem token retorna `401`
- Todos os endpoints protegidos funcionam com token válido
- A estrutura de injeção de dependências permanece legível e seguindo o padrão manual existente

---

## Task 9

Testar fluxo completo de autenticação

### Objetivo

Validar a feature ponta a ponta

### Validação

- Criar usuário via `POST /v1/users` resulta em senha hasheada no banco
- Fazer login com o usuário criado → token recebido com sucesso
- Usar token recebido para acessar `GET /v1/users` → sucesso
- Usar token inválido para acessar `GET /v1/users` → `401`
- Usar token expirado para acessar `GET /v1/users` → `401`
- Fazer login com senha errada → `401`
- Fazer login com email inexistente → `401`
- Acessar `POST /v1/auth/login` com corpo malformado → `400`
