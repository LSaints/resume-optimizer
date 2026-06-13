# PLAN.md

## Existing Context

### Architecture

### Stack

- **Frontend**: React 19 + TypeScript 6 + Vite 8
- **Backend**: Go 1.26 (API REST em `/v1/*`)
- **Auth**: JWT via header `Authorization: Bearer <token>`

### Existing Conventions

- Versão de rota: `/v1/<resource>`
- Mensagens em português
- Dependências do frontend atualmente: apenas `react` e `react-dom`
- Sem roteador, sem cliente HTTP, sem componentes de UI atualmente

---

## Arquitetura

Aplicação Single Page Application (SPA) com React 19, utilizando roteador cliente (`react-router-dom`) para navegação entre páginas. O estado de autenticação será gerenciado via React Context. A comunicação com a API será feita através de uma camada de serviço centralizada usando `fetch` nativo.

---

## Diretrizes de Design

Seguindo o skill `frontend-design`, a interface deverá:

- **Tom visual**: editorial/profissional com personalidade — tipografia marcante, espaçamento generoso, paleta de alto contraste. Evitar estética genérica de IA (gradientes roxos, Inter, layouts simétricos previsíveis).
- **Tipografia**: uma fonte display distinta para títulos (ex: `DM Serif Display`, `Syne`, `Sora`) combinada com uma fonte de corpo refinada (ex: `Literata`, `IBM Plex Serif`, `STIX Two Text`). Carregadas via Google Fonts.
- **Paleta**: fundo escuro ou tom sóbrio com acentos vibrantes. Cores com propósito semântico claro.
- **Movimento**: transições suaves entre páginas, microfinterações em botões e cards, loading states com skeleton ou shimmer.
- **Componentes**: biblioteca própria de componentes, sem dependências de UI prontas (sem shadcn, MUI, Chakra). Botões, inputs, cards, modais e selects estilizados com CSS Modules ou CSS puro com variáveis.
- **Responsivo**: grid flexível, navegação adaptativa (drawer em mobile).

---

## Estrutura Técnica

### Árvore de Diretórios

```
src/
├── main.tsx
├── App.tsx
├── index.css
├── assets/
├── pages/
│   ├── LoginPage.tsx
│   ├── RegisterPage.tsx
│   ├── DashboardPage.tsx
│   ├── ResumeListPage.tsx
│   ├── ResumeUploadPage.tsx
│   ├── JobListPage.tsx
│   ├── JobFormPage.tsx
│   ├── OptimizePage.tsx
│   └── TypstViewerPage.tsx
├── components/
│   ├── Layout.tsx
│   ├── ProtectedRoute.tsx
│   ├── Header.tsx
│   ├── Sidebar.tsx
│   ├── FileUpload.tsx
│   ├── TypstRenderer.tsx
│   ├── ResumeCard.tsx
│   ├── JobCard.tsx
│   ├── OptimizationCard.tsx
│   ├── Button.tsx
│   ├── Input.tsx
│   ├── Select.tsx
│   ├── Modal.tsx
│   └── LoadingSpinner.tsx
├── contexts/
│   └── AuthContext.tsx
├── services/
│   ├── api.ts
│   ├── authService.ts
│   ├── resumeService.ts
│   ├── jobService.ts
│   └── optimizationService.ts
├── hooks/
│   ├── useAuth.ts
│   └── useApi.ts
├── types/
│   ├── user.ts
│   ├── resume.ts
│   ├── job.ts
│   └── optimization.ts
└── utils/
    └── storage.ts
```

### Páginas

| Página | Rota | Autenticação | Descrição |
|---|---|---|---|
| Login | `/login` | Pública | Formulário de login (email + senha) |
| Register | `/register` | Pública | Formulário de registro (nome + email + senha) |
| Dashboard | `/` | Protegida | Visão geral com resumo de currículos, vagas e otimizações recentes |
| Meus Currículos | `/resumes` | Protegida | Lista de currículos enviados com ações (ver, excluir) |
| Enviar Currículo | `/resumes/new` | Protegida | Upload de arquivo PDF/DOCX |
| Minhas Vagas | `/jobs` | Protegida | Lista de vagas cadastradas com ações (editar, excluir) |
| Nova Vaga | `/jobs/new` | Protegida | Formulário de cadastro de vaga |
| Editar Vaga | `/jobs/:id/edit` | Protegida | Formulário de edição de vaga |
| Otimizar Currículo | `/optimize` | Protegida | Seleção de currículo + vaga + disparo de otimização |
| Visualizador Typst | `/optimizations/:id` | Protegida | Exibição do resultado Typst da otimização |

### Componentes

#### Layout

- `Layout.tsx` — estrutura principal com `Header`, `Sidebar` (ou navbar inferior em mobile) e área de conteúdo. Verifica autenticação e redireciona se necessário.
- `Header.tsx` — logo, navegação principal, avatar/nome do usuário, botão de logout
- `Sidebar.tsx` — links de navegação: Dashboard, Currículos, Vagas, Otimizar
- `ProtectedRoute.tsx` — wrapper que verifica `AuthContext` e redireciona para `/login` se não autenticado

#### Dados

- `ResumeCard.tsx` — card de currículo na listagem (nome, data, ações)
- `JobCard.tsx` — card de vaga na listagem (título, preview da descrição, data, ações)
- `OptimizationCard.tsx` — card de otimização no histórico (vaga associada, data)

#### Formulários

- `FileUpload.tsx` — componente de upload com drag-and-drop, validação de tipo e tamanho
- `Button.tsx`, `Input.tsx`, `Select.tsx` — componentes de formulário estilizados, com estados de erro, disabled, loading

#### Visualização

- `TypstRenderer.tsx` — componente que recebe código Typst como string e o renderiza visualmente. Inicialmente fará uma renderização textual formatada (como preview estilizado). Poderá evoluir para usar `@typst/typst` ou WebAssembly. Inclui botão de copiar código.

### Serviços

#### `api.ts` — Cliente HTTP base

- `apiClient` — função `fetch` wrapper que:
  - Prefixa URLs com `http://localhost:8080/v1`
  - Injeta header `Authorization: Bearer <token>` automaticamente
  - Gerencia Content-Type (`application/json` ou `multipart/form-data`)
  - Lida com erros HTTP e os traduz para mensagens em português
  - Retorna JSON tipado

#### `authService.ts`

- `login(email, password): LoginResponse` — `POST /v1/auth/login`
- `register(name, email, password): UserResponse` — `POST /v1/users`

#### `resumeService.ts`

- `list(): ResumeResponse[]` — `GET /v1/resumes`
- `get(id): ResumeResponse` — `GET /v1/resumes/{id}`
- `upload(file: File): ResumeResponse` — `POST /v1/resumes` (multipart)
- `delete(id): void` — `DELETE /v1/resumes/{id}`

#### `jobService.ts`

- `list(): JobResponse[]` — `GET /v1/jobs`
- `get(id): JobResponse` — `GET /v1/jobs/{id}`
- `create(data): JobResponse` — `POST /v1/jobs`
- `update(id, data): JobResponse` — `PUT /v1/jobs/{id}`
- `delete(id): void` — `DELETE /v1/jobs/{id}`

#### `optimizationService.ts`

- `optimize(resumeId, jobId): OptimizeResponse` — `POST /v1/resumes/{resumeID}/optimize`
- `listByResume(resumeId): OptimizeSummaryResponse[]` — `GET /v1/resumes/{resumeID}/optimizations`
- `getByID(resumeId, optimizationId): OptimizeResponse` — `GET /v1/resumes/{resumeID}/optimizations/{optimizationID}`

### Contextos

#### `AuthContext.tsx`

Gerencia:

- `user` — dados do usuário logado (ou `null`)
- `token` — JWT armazenado
- `isAuthenticated` — booleano
- `login(email, password)` — chama `authService.login`, armazena token no `localStorage`, atualiza estado
- `register(name, email, password)` — chama `authService.register`, automaticamente faz login
- `logout()` — limpa token do `localStorage`, redireciona para `/login`
- `loading` — estado de carregamento inicial (verificando token salvo)

O contexto é inicializado verificando se há um token salvo no `localStorage` e validando-o com a API.

### Tipos TypeScript

```typescript
// user.ts
interface UserResponse {
  id: string
  name: string
  email: string
}

// auth.ts
interface LoginRequest {
  email: string
  password: string
}
interface LoginResponse {
  token: string
  user: UserResponse
}

// resume.ts
interface ResumeResponse {
  id: string
  originalName: string
  uploadedAt: string
}

// job.ts
interface JobRequest {
  title: string
  rawDescription: string
}
interface JobResponse {
  id: string
  title: string
  rawDescription: string
  createdAt: string
  updatedAt: string
}

// optimization.ts
interface OptimizeRequest {
  jobId: string
}
interface OptimizeResponse {
  id: string
  resumeId: string
  jobId: string
  typstContent: string
  createdAt: string
}
interface OptimizeSummaryResponse {
  id: string
  resumeId: string
  jobId: string
  createdAt: string
}
```

---

## Fluxos

### Autenticação

```
Usuário → /login → preenche email+senha → AuthContext.login()
  → authService.login() → POST /v1/auth/login
  → recebe token + user → salva token no localStorage
  → redireciona para /
  → Header exibe nome do usuário + logout
```

### Upload de Currículo

```
Usuário → /resumes/new → seleciona arquivo (PDF/DOCX) → submit
  → FileUpload valida tipo (.pdf, .docx) e tamanho (< 10MB)
  → resumeService.upload(file) → POST /v1/resumes (multipart)
  → redireciona para /resumes com confirmação
```

### Otimização

```
Usuário → /optimize
  → carrega lista de currículos e vagas do usuário (via services)
  → seleciona um currículo e uma vaga em selects
  → clica "Otimizar"
  → optimizationService.optimize(resumeId, jobId)
  → POST /v1/resumes/{resumeID}/optimize
  → exibe loading durante a chamada (pode levar até 60s)
  → redireciona para /optimizations/{id} com resultado Typst
```

### Visualização Typst

```
Usuário → /optimizations/{id}
  → optimizationService.getByID(resumeId, optimizationId)
  → TypstRenderer recebe typstContent como string
  → exibe conteúdo formatado visualmente
  → botão "Copiar código" copia typstContent para clipboard
```

---

## Dependências a adicionar

- `react-router-dom` — roteamento SPA
- Fonte display via Google Fonts (ex: `DM Serif Display`, `Syne`)
- Fonte de corpo via Google Fonts (ex: `Literata`, `STIX Two Text`, `IBM Plex Serif`)
- Sem dependências de UI — componentes próprios

---

## Decisões Técnicas

- **Sem biblioteca de componentes UI**: componentes próprios estilizados com CSS Modules, garantindo identidade visual única e evitando inchaço de dependências
- **fetch nativo em vez de axios**: uma dependência a menos; `fetch` é suficiente com um wrapper simples
- **Context API em vez de Zustand/Redux**: estado global mínimo (apenas autenticação); o restante é estado local de página
- **localStorage para token**: simplicidade; o token JWT já tem expiração própria (24h)
- **CSS Modules ou CSS puro com variáveis**: sem runtime CSS-in-JS para evitar custo de performance; variáveis CSS garantem consistência temática
- **TypstRenderer textual inicialmente**: exibição formatada do código Typst com realce de sintaxe e estrutura visual limpa. Em versão futura pode integrar engine Typst via WASM
- **Loading states explícitos**: botões com spinner durante chamadas à API (especialmente otimização que leva até 60s)
