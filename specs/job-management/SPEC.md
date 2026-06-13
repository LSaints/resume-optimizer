# SPEC.md

## Feature

Gerenciamento de Vagas

---

## Objetivo

Permitir que o usuário autenticado cadastre, edite, liste, visualize e exclua vagas (descrições de posições/oportunidades), servindo como base para futura vinculação com currículos.

---

## Requisitos de negócio

- O usuário deve estar autenticado para gerenciar vagas
- O usuário pode cadastrar uma vaga com título e descrição
- O usuário pode editar título e descrição de uma vaga existente
- O usuário pode listar todas as suas vagas cadastradas
- O usuário pode visualizar os detalhes de uma vaga específica
- O usuário pode excluir uma vaga
- Cada vaga pertence a exatamente um usuário
- Futuramente, uma vaga poderá ser vinculada a um ou mais currículos

---

## Restrições

- O título é obrigatório e não pode estar vazio
- A descrição é obrigatória e não pode estar vazia
- Título e descrição são campos de texto livre (sem formatação rica)
- O usuário só pode acessar, alterar ou excluir suas próprias vagas
- A exclusão deve impedir vínculos futuros com currículos (sem cascade neste momento)

---

## Critérios de Aceitação

- Um usuário autenticado consegue criar uma vaga com título e descrição
- Um usuário autenticado consegue editar título e descrição de uma vaga existente
- Um usuário autenticado consegue listar todas as suas vagas
- Um usuário autenticado consegue visualizar os detalhes de uma vaga específica
- Um usuário autenticado consegue excluir uma vaga
- Tentativa de criar vaga com título vazio retorna erro 400
- Tentativa de criar vaga com descrição vazia retorna erro 400
- Um usuário não consegue acessar vagas de outro usuário
- Usuário não autenticado recebe 401 em todas as operações
