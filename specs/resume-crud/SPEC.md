# SPEC.md

## Feature

CRUD de Currículos

---

## Objetivo

Permitir que o usuário autenticado envie seu currículo nos formatos PDF ou DOCX, tenha o texto extraído automaticamente e gerencie (CRUD) os currículos cadastrados.

---

## Requisitos de negócio

- O usuário deve estar autenticado para gerenciar currículos
- O usuário pode enviar um arquivo de currículo nos formatos PDF ou DOCX
- O sistema deve extrair o conteúdo textual do arquivo enviado
- Apenas o texto extraído deve ser armazenado, nunca o arquivo original
- O usuário pode listar todos os seus currículos cadastrados
- O usuário pode visualizar o conteúdo completo de um currículo específico
- O usuário pode substituir um currículo existente por um novo arquivo (update)
- O usuário pode excluir um currículo
- Cada currículo pertence a exatamente um usuário

---

## Restrições

- Apenas arquivos com extensão `.pdf` e `.docx` são aceitos
- O tamanho máximo do arquivo enviado é de 10 MB
- O arquivo original é descartado imediatamente após a extração do texto
- O usuário só pode acessar, alterar ou excluir seus próprios currículos
- O campo `rawText` é obrigatório para criação e atualização via extração de arquivo

---

## Critérios de Aceitação

- Um usuário autenticado consegue enviar um PDF e ter o texto salvo no banco
- Um usuário autenticado consegue enviar um DOCX e ter o texto salvo no banco
- A listagem de currículos retorna apenas os currículos do usuário logado
- A listagem de currículos retorna metadados sem o texto completo
- Ao visualizar um currículo individual, o texto extraído é retornado
- Ao atualizar com novo arquivo, o texto antigo é substituído pelo novo
- Ao excluir, o currículo é removido permanentemente
- Um usuário não consegue acessar currículos de outro usuário
- Tentativa de upload de formato não suportado retorna erro 400
- Tentativa de upload de arquivo maior que 10 MB retorna erro 400
- Usuário não autenticado recebe 401 em todas as operações
