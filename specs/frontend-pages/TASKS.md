# TASKS.md

## Task 1

Configurar roteamento e dependências iniciais

### Objetivo

Adicionar `react-router-dom` ao projeto, configurar o roteador no `<App>` com todas as rotas planejadas (páginas placeholder inicialmente), e instalar fontes tipográficas definidas no PLAN.md.

### Validação

- `npm install react-router-dom` executa sem erros
- O arquivo `src/App.tsx` utiliza `BrowserRouter` com `Routes` e `Route`
- As rotas `/login`, `/register`, `/`, `/resumes`, `/resumes/new`, `/jobs`, `/jobs/new`, `/jobs/:id/edit`, `/optimize`, `/optimizations/:id` estão definidas
- Cada rota renderiza um componente placeholder
- As fontes escolhidas são carregadas via `@import` no `index.css` e aplicadas como variáveis CSS (`--font-display`, `--font-body`)
- `npm run dev` inicia sem erros e a navegação entre todas as rotas funciona

---

## Task 2

Criar tema CSS e design tokens

### Objetivo

Estabelecer o sistema de design visual do projeto: variáveis CSS para cores, tipografia, espaçamento, bordas e sombras. Criar a identidade visual com paleta definida, tipografia marcante e tokens de design consistentes.

### Validação

- `src/styles/tokens.css` (ou `index.css`) contém variáveis CSS no `:root`:
  - `--color-bg`, `--color-surface`, `--color-text`, `--color-primary`, `--color-accent`, `--color-error`, `--color-success`
  - `--font-display`, `--font-body` com as fontes carregadas
  - `--space-*` para espaçamento consistente (ex: `--space-xs` a `--space-xxl`)
  - `--radius-sm`, `--radius-md`, `--radius-lg` para bordas
  - `--shadow-sm`, `--shadow-md`, `--shadow-lg` para elevação
- O tema é aplicado globalmente e visível em qualquer componente
- A paleta foge do padrão genérico (sem roxo claro + branco) — tom sóbrio com acento vibrante

---

## Task 3

Criar componentes de UI base (Button, Input, Select, LoadingSpinner)

### Objetivo

Implementar os componentes de interface reutilizáveis seguindo os tokens de design definidos na Task 2. Cada componente deve suportar estados de erro, desabilitado, loading e focus, com microfinterações CSS.

### Validação

- `Button.tsx` — suporta variantes (`primary`, `secondary`, `ghost`), tamanhos (`sm`, `md`, `lg`), estado `disabled`, estado `loading` (com spinner interno), transição hover/active suave
- `Input.tsx` — suporta label, placeholder, mensagem de erro, estado `disabled`, focus ring customizado
- `Select.tsx` — suporta label, opções, mensagem de erro, estado `disabled`
- `LoadingSpinner.tsx` — animação CSS sutil, suporta tamanhos e cores via props
- Todos os componentes aceitam `className` para customização externa
- Os componentes podem ser visualizados isoladamente em uma página de teste temporária

---

## Task 4

Criar AuthContext, hook useAuth e serviço de autenticação

### Objetivo

Implementar o contexto de autenticação global (`AuthContext`) que gerencia estado do usuário, token JWT e provê funções de login, registro e logout. Criar `authService.ts` com chamadas para a API e `utils/storage.ts` para gerenciar token no `localStorage`.

### Validação

- `AuthContext` expõe: `user`, `token`, `isAuthenticated`, `loading`, `login()`, `register()`, `logout()`
- `login(email, password)` chama `POST /v1/auth/login`, armazena token no `localStorage`, atualiza estado
- `register(name, email, password)` chama `POST /v1/users`, armazena token (login automático), atualiza estado
- `logout()` limpa `localStorage` e redefine estado para valores iniciais
- Ao montar, `AuthContext` verifica se há token no `localStorage` e, se existir, recupera dados do usuário via API
- `storage.ts` exporta `getToken()`, `setToken(token)`, `removeToken()`
- `useAuth()` hook retorna o contexto com tipagem correta

---

## Task 5

Criar Layout, ProtectedRoute e Header

### Objetivo

Implementar a estrutura de layout compartilhada: `ProtectedRoute` que redireciona para `/login` se não autenticado, `Layout` com header e sidebar (ou navbar inferior em mobile), e `Header` com navegação e informações do usuário.

### Validação

- `ProtectedRoute` renderiza `<Navigate to="/login" />` se `isAuthenticated` for `false`
- `ProtectedRoute` renderiza `<Outlet />` se `isAuthenticated` for `true`
- `Layout` exibe `Header` + `<Outlet />` (via `react-router-dom`), com transição suave entre páginas
- `Header` exibe: logo/marca, links de navegação (Dashboard, Currículos, Vagas, Otimizar), nome do usuário, botão de logout
- Em mobile (< 768px), a navegação vira drawer ou navbar inferior
- Layout é responsivo e ocupa altura total da viewport

---

## Task 6

Criar página de Login e Register

### Objetivo

Implementar as páginas públicas de autenticação com formulários funcionais, validação inline e integração com `AuthContext`.

### Validação

- `/login` exibe formulário com campos de email e senha, botão "Entrar", link para "/register"
- `/register` exibe formulário com campos de nome, email e senha, botão "Criar conta", link para "/login"
- Validação de campos: email válido, senha com mínimo de 6 caracteres, nome não vazio
- Botão de submit mostra loading enquanto a requisição está em andamento
- Erro da API (credenciais inválidas, email duplicado) é exibido como mensagem amigável
- Após login/registro bem-sucedido, redireciona para `/`
- Se o usuário já está autenticado e acessa `/login` ou `/register`, redireciona para `/`
- Design visual marcante e consistente com os tokens definidos

---

## Task 7

Criar API service layer (api.ts, resumeService.ts, jobService.ts, optimizationService.ts)

### Objetivo

Implementar a camada de comunicação com a API backend: cliente HTTP base com injeção automática de token JWT, tratamento de erros, e serviços específicos para cada recurso.

### Validação

- `api.ts` exporta funções `get`, `post`, `put`, `del` que:
  - Prefixam URLs com `http://localhost:8080/v1`
  - Injetam header `Authorization: Bearer <token>` quando token existe
  - Definim `Content-Type: application/json` para requisições JSON
  - Não definem `Content-Type` para `multipart/form-data` (deixa o browser definir)
  - Traduzem erros HTTP para mensagens em português
  - Retornam dados parseados com tipo genérico
- `resumeService.ts` — `list()`, `get(id)`, `upload(file)`, `delete(id)`
- `jobService.ts` — `list()`, `get(id)`, `create(data)`, `update(id, data)`, `delete(id)`
- `optimizationService.ts` — `optimize(resumeId, jobId)`, `listByResume(resumeId)`, `getByID(resumeId, optimizationId)`
- Todos os serviços retornam Promises tipadas com as interfaces definidas no PLAN.md

---

## Task 8

Criar página Dashboard

### Objetivo

Implementar a página inicial do usuário autenticado com visão geral: cards com contagem de currículos, vagas e otimizações recentes.

### Validação

- `/` exibe saudação com nome do usuário
- Cards mostram quantidades: "X currículos enviados", "Y vagas cadastradas", "Z otimizações realizadas"
- Cada card tem link para a página correspondente
- Layout em grid responsivo (1 coluna mobile, 2-3 colunas desktop)
- Esqueletos de carregamento (skeleton) enquanto dados são carregados
- Design consistente com o tema

---

## Task 9

Criar páginas de Currículo (lista e upload)

### Objetivo

Implementar a listagem de currículos do usuário e a página de upload com componente de drag-and-drop.

### Validação

- `/resumes` lista todos os currículos do usuário com: nome original do arquivo, data de upload, botões de visualizar e excluir
- Ao excluir, exibe modal de confirmação; após confirmar, remove da lista com animação
- Se não há currículos, exibe estado vazio com link para upload
- `/resumes/new` exibe:
  - Área de upload com drag-and-drop (destacar ao arrastar arquivo)
  - Validação de tipo (apenas `.pdf`, `.docx`) e tamanho (máximo 10MB)
  - Barra de progresso ou spinner durante upload
  - Mensagem de erro se arquivo inválido
  - Redireciona para `/resumes` após sucesso com feedback visual
- Botão "Voltar" em ambas as páginas

---

## Task 10

Criar páginas de Vaga (lista, criação e edição)

### Objetivo

Implementar o CRUD de vagas com listagem, formulário de criação/edição e exclusão.

### Validação

- `/jobs` lista todas as vagas do usuário com: título, preview da descrição, data de criação, botões de editar e excluir
- Ao excluir, exibe modal de confirmação
- Estado vazio com link para criar nova vaga
- `/jobs/new` e `/jobs/:id/edit` exibem formulário com campos:
  - Título (obrigatório, mínimo 3 caracteres)
  - Descrição (obrigatória, textarea com altura ajustável)
- Botão de submit mostra loading e exibe erro da API se houver
- Salvar redireciona para `/jobs` com feedback visual
- Ao editar, os campos são pré-preenchidos com dados da vaga
- Botão "Cancelar" retorna para `/jobs`

---

## Task 11

Criar páginas de Otimização (seleção e resultado)

### Objetivo

Implementar a página de otimização com seleção de currículo e vaga, e a página de visualização do resultado Typst.

### Validação

- `/optimize` exibe:
  - Select de currículos (carregados via `resumeService.list()`)
  - Select de vagas (carregados via `jobService.list()`)
  - Botão "Otimizar" desabilitado até que ambos sejam selecionados
  - Botão mostra spinner com texto "Otimizando..." durante a chamada (pode levar até 60s)
  - Mensagens de erro amigáveis (currículo não encontrado, vaga não encontrada, API sem chave)
  - Redireciona para `/optimizations/{id}` após sucesso
- `/optimizations/:id` exibe:
  - Metadados: nome do currículo e título da vaga usados, data
  - Componente `TypstRenderer` com o conteúdo Typst formatado visualmente
  - Botão "Copiar código" que copia o `typstContent` para a área de transferência
  - Botão "Voltar" ou "Nova otimização"
  - Link para ver histórico de otimizações do currículo

---

## Task 12

Criar componente TypstRenderer

### Objetivo

Implementar o componente de visualização do código Typst que exibe o currículo otimizado de forma estruturada e legível, com realce visual e opção de copiar.

### Validação

- `TypstRenderer.tsx` recebe `content: string` como prop
- Renderiza o conteúdo Typst em um container estilizado com:
  - Tipografia monoespaçada ou serifada para o código
  - Preservação de quebras de linha e indentação
  - Rolagem vertical para conteúdos longos
  - Fundo contrastante com borda sutil
- Botão "Copiar código" utiliza `navigator.clipboard.writeText()` e exibe feedback visual ("Copiado!" por 2s)
- O container tem largura máxima e centralização na página
- Design do visualizador é refinado e profissional, remetendo a um documento editorado
- Responsivo: ocupa largura total em mobile

---

## Task 13

Refinar design responsivo e testar fluxo completo

### Objetivo

Revisar todas as páginas para garantir responsividade completa em mobile, tablet e desktop. Testar o fluxo ponta a ponta de todas as funcionalidades.

### Validação

- Todas as páginas funcionam em viewport de 375px (mobile), 768px (tablet) e 1440px (desktop)
- Navegação adapta para drawer em mobile
- Formulários são utilizáveis em telas pequenas
- Tipografia escala corretamente
- Fluxo completo de registro → login → upload de currículo → cadastro de vaga → otimização → visualização do Typst funciona sem erros
- Logout limpa sessão e redireciona para login
- Acessar rota protegida sem token redireciona para login
- Tratamento de erro da API exibe mensagens amigáveis em português
- Build de produção (`npm run build`) executa sem erros
