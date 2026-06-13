# SPEC.md

## Feature

Páginas do Frontend

---

## Objetivo

Prover uma interface web funcional e visualmente marcante para o sistema de otimização de currículos, contemplando autenticação, gerenciamento de currículos e vagas, otimização com IA e visualização do resultado em Typst.

---

## Requisitos de negócio

### Autenticação

- O usuário deve poder se registrar informando nome, e-mail e senha
- O usuário deve poder fazer login com e-mail e senha
- O usuário deve permanecer autenticado entre sessões (token persistente)
- O usuário deve poder sair da conta (logout)
- Rotas protegidas devem redirecionar para o login se o usuário não estiver autenticado

### Gerenciamento de Currículos

- O usuário autenticado deve poder enviar um arquivo de currículo (PDF ou DOCX)
- O sistema deve exibir a lista de currículos enviados pelo usuário
- O usuário deve poder visualizar detalhes de um currículo específico
- O usuário deve poder excluir um currículo

### Gerenciamento de Vagas

- O usuário autenticado deve poder cadastrar uma vaga informando título e descrição
- O sistema deve exibir a lista de vagas cadastradas pelo usuário
- O usuário deve poder editar e excluir uma vaga

### Otimização de Currículo

- O usuário autenticado deve poder selecionar um currículo e uma vaga para otimização
- O sistema deve disparar a otimização via IA e exibir o resultado em formato Typst
- O usuário deve poder visualizar o histórico de otimizações de um currículo
- O usuário deve poder visualizar o conteúdo Typst de uma otimização específica em um visualizador dedicado

### Visualizador Typst

- O sistema deve exibir o conteúdo Typst gerado de forma visualmente estruturada
- O visualizador deve apresentar o currículo otimizado de maneira legível e profissional
- O usuário deve poder copiar o código Typst gerado

---

## Restrições

- O frontend deve se comunicar exclusivamente com a API existente (`/v1/*`)
- Não deve haver lógica de negócio no frontend — apenas validação de formulário e apresentação
- O token JWT deve ser armazenado no `localStorage` e enviado no header `Authorization: Bearer <token>`
- O upload de currículo deve respeitar o limite de 10MB definido pela API
- O design deve ser responsivo (mobile e desktop)
- As mensagens de erro devem ser amigáveis e em português

---

## Critérios de Aceitação

- Um novo usuário consegue se registrar e fazer login
- Um usuário autenticado consegue enviar um currículo (PDF ou DOCX) e vê-lo na lista
- Um usuário autenticado consegue cadastrar uma vaga e vê-la na lista
- Um usuário autenticado consegue selecionar um currículo e uma vaga e gerar uma otimização
- O resultado da otimização é exibido no visualizador Typst de forma legível
- O histórico de otimizações lista todas as versões geradas para um currículo
- Um usuário não autenticado é redirecionado para o login ao acessar rotas protegidas
- O usuário consegue copiar o código Typst do resultado da otimização
- A interface é responsiva e funciona em dispositivos móveis
