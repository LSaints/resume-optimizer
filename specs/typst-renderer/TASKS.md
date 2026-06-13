# TASKS.md

## Task 1

Criar TypstRenderService no backend

### Objetivo

Implementar `TypstRenderService` em `internal/application/services/typst_render_service.go` responsável por converter código Typst em SVG e PDF utilizando o binário `typst` via `os/exec`

### Validação

- `NewTypstRenderService() *TypstRenderService` existe
- `RenderToSVG(typstContent string) (string, error)`:
  - Cria arquivo temporário com o conteúdo `.typ`
  - Executa `typst compile input.typ output.svg --format svg`
  - Retorna o conteúdo do SVG como string
  - Limpa arquivos temporários com `defer os.Remove`
  - Usa timeout de 15 segundos via `context.WithTimeout`
- `RenderToPDF(typstContent string) ([]byte, error)`:
  - Executa `typst compile input.typ output.pdf`
  - Retorna os bytes do PDF
- Retorna erro "renderizador typst nao disponivel" se `typst` não estiver no PATH
- Retorna erro "erro ao renderizar documento" se a compilação falhar (Typst inválido)
- O serviço não depende de banco de dados ou outras camadas

---

## Task 2

Criar RenderResponse DTO

### Objetivo

Criar struct de response para a renderização em `internal/application/responses/render_response.go`

### Validação

- `RenderResponse` com campo `SvgContent string` e tag `json:"svgContent"`
- Struct segue padrão de nomenclatura e localização do projeto
- Código compila sem erros

---

## Task 3

Criar RenderHandler com endpoints de renderização

### Objetivo

Implementar `RenderHandler` em `internal/apresentation/handlers/render_handlers.go` com métodos para renderizar SVG e servir PDF

### Validação

- `NewRenderHandler(optService *OptimizationServices, renderService *TypstRenderService) *RenderHandler` existe
- `RenderSVG(w, r)` — `GET /v1/optimizations/{optimizationID}/render`:
  - Lê `optimizationID` de `r.PathValue("optimizationID")`
  - Lê `userID` do contexto
  - Busca otimização via `OptimizationServices.GetByID` (herda validação de ownership)
  - Chama `TypstRenderService.RenderToSVG`
  - Retorna 200 com `RenderResponse` JSON
  - Se otimização não encontrada, retorna 404
  - Se renderizador indisponível, retorna 502
  - Se compilação falhar, retorna 502
- `RenderPDF(w, r)` — `GET /v1/optimizations/{optimizationID}/render/pdf`:
  - Mesma validação de ownership
  - Retorna PDF com `Content-Type: application/pdf`
  - Header `Content-Disposition: attachment; filename="curriculo-otimizado.pdf"`
- Content-Type `application/json` definido em respostas JSON

---

## Task 4

Registrar rotas de renderização e conectar dependências

### Objetivo

Atualizar `internal/apresentation/routes/routes.go` para instanciar `TypstRenderService` e `RenderHandler`, e registrar as novas rotas

### Validação

- `TypstRenderService`, `RenderHandler` instanciados em ordem no `RegisterRoutes`
- Rotas registradas protegidas pelo `AuthMiddleware`:
  - `GET /v1/optimizations/{optimizationID}/render`
  - `GET /v1/optimizations/{optimizationID}/render/pdf`
- A aplicação compila sem erros
- O servidor inicializa e aceita requisições nas novas rotas

---

## Task 5

Atualizar TypstRenderer no frontend com suporte a renderização SVG

### Objetivo

Substituir `frontend/src/components/TypstRenderer.tsx` para exibir o SVG renderizado, com alternância entre visualização renderizada e código-fonte

### Validação

- Props: `{ content: string, optimizationID?: string }`
- Quando `optimizationID` é fornecido:
  - Faz fetch para `GET /v1/optimizations/{optimizationID}/render`
  - Exibe loading spinner enquanto carrega
  - Exibe SVG inline após carregar
  - Em caso de erro, exibe mensagem e fallback para código-fonte
- Botão de alternância entre modo "renderizado" e "código-fonte"
- Botão "Copiar código" visível no modo código-fonte, copia `content` para clipboard
- Botão "Baixar PDF" visível no modo renderizado, navega para `/v1/optimizations/{id}/render/pdf`
- Mantém estilo existente (`.module.css`) atualizado para suportar ambos os modos
- Quando `optimizationID` não é fornecido, exibe apenas código-fonte (comportamento legado)

---

## Task 6

Criar renderService no frontend

### Objetivo

Criar `frontend/src/services/renderService.ts` com funções para buscar SVG e obter URL de download PDF

### Validação

- `getRenderSVG(optimizationID: string): Promise<string>`:
  - Faz GET para `/v1/optimizations/{optimizationID}/render`
  - Retorna o campo `svgContent` da resposta
- `getDownloadPDFURL(optimizationID: string): string`:
  - Retorna a URL absoluta para download do PDF
- Erros da API são propagados (401, 404, 502)

---

## Task 7

Criar tipos de renderização no frontend

### Objetivo

Criar `frontend/src/types/render.ts` com interface do response de renderização

### Validação

- `RenderResponse` com campo `svgContent: string`
- Interface exportada e utilizável pelos serviços

---

## Task 8

Criar OptimizationHistoryPage

### Objetivo

Criar `frontend/src/pages/OptimizationHistoryPage.tsx` listando todas as otimizações de um currículo com preview renderizado

### Validação

- Rota: `/resumes/:id/optimizations`
- Carrega lista de otimizações via `optimizationService.listByResume(resumeId)`
- Carrega metadados do currículo e vagas para exibir nomes
- Cada item do histórico exibe:
  - Thumbnail SVG renderizado (fetch via `renderService.getRenderSVG`)
  - Nome do currículo
  - Título da vaga
  - Data formatada (pt-BR)
- Clique no item navega para `/optimizations/:resumeId/:optimizationId`
- Botão "Nova otimização" navega para `/optimize`
- Botão "Excluir" remove a otimização (confirmação via modal ou confirmação nativa)
- Exclusão atualiza a lista sem recarregar a página
- Estado de loading exibe skeleton
- Estado vazio exibe mensagem "Nenhuma otimização encontrada" com link para nova otimização
- Estado de erro exibe mensagem apropriada

---

## Task 9

Atualizar TypstViewerPage para usar TypstRenderer aprimorado

### Objetivo

Modificar `frontend/src/pages/TypstViewerPage.tsx` para passar `optimizationID` ao `TypstRenderer`, habilitando a renderização SVG

### Validação

- `TypstRenderer` recebe `content={optimization.typstContent}` e `optimizationID={optimizationId}`
- O SVG renderizado aparece na página (não apenas código-fonte)
- Alternância entre renderizado e código-fonte funciona
- Botão "Baixar PDF" aparece e baixa o PDF corretamente
- Botão "Ver histórico" navega para `/resumes/:resumeId/optimizations`
- Metadados e ações existentes permanecem funcionando

---

## Task 10

Adicionar rota do histórico no frontend

### Objetivo

Registrar a nova rota `/resumes/:id/optimizations` em `frontend/src/App.tsx` e adicionar navegação no Header quando aplicável

### Validação

- Rota `/resumes/:id/optimizations` registrada dentro do `ProtectedRoute` > `Layout`
- `OptimizationHistoryPage` é importada e usada na rota
- Navegação para a página funciona (via botão "Ver histórico" no TypstViewerPage)
- A página renderiza corretamente ao acessar a URL diretamente

---

## Task 11

Testar fluxo completo de renderização e visualização

### Objetivo

Validar a feature ponta a ponta incluindo renderização, visualização, histórico e download

### Validação

- Cenário 1: Otimizar currículo → visualizar renderizado no TypstViewerPage → alternar para código-fonte → copiar código → baixar PDF → tudo funciona
- Cenário 2: Acessar histórico de otimizações de um currículo → ver thumbnails → clicar em uma otimização → abre página com renderização completa
- Cenário 3: Excluir otimização do histórico → otimização desaparece → tentar acessar URL direta → erro 404
- Cenário 4: Tentar acessar renderização de otimização de outro usuário → erro 404
- Cenário 5: Tentar acessar rota de renderização sem autenticação → erro 401
- Cenário 6: Tipst inválido no banco → renderização retorna erro 502 → frontend mostra fallback para código-fonte
