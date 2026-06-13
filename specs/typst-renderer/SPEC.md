# SPEC.md

## Feature

Visualizador de Otimizações com Renderizador Typst

---

## Objetivo

Permitir que o usuário visualize as otimizações de currículo renderizadas visualmente como documento formatado (não apenas como código-fonte Typst), e navegue pelo histórico de otimizações em uma interface dedicada.

---

## Requisitos de negócio

- O usuário deve visualizar o currículo otimizado como um documento visual formatado
- O usuário deve poder alternar entre a visualização renderizada e o código-fonte Typst
- O usuário deve poder copiar o código-fonte Typst com um clique
- O usuário deve poder navegar pelo histórico completo de otimizações de um currículo
- O histórico deve exibir um preview renderizado (thumbnail) de cada otimização
- O usuário deve poder baixar o currículo otimizado como PDF
- O usuário deve poder acessar a visualização renderizada de uma otimização específica via URL direta
- O usuário deve poder excluir uma otimização do histórico
- A visualização renderizada deve refletir fielmente o código Typst gerado

---

## Restrições

- A renderização não pode depender de serviços externos de terceiros (deve ser processada localmente)
- A renderização não pode modificar o conteúdo Typst original armazenado no banco
- Apenas otimizações pertencentes a currículos do usuário logado podem ser acessadas
- O código-fonte Typst deve permanecer acessível independentemente da renderização
- O tempo de renderização não pode exceder 15 segundos por requisição

---

## Critérios de Aceitação

- Uma otimização é exibida como documento visual formatado na página de visualização
- O usuário alterna entre visualização renderizada e código-fonte com um clique
- O preview renderizado carrega dentro de 15 segundos
- O histórico de otimizações lista todas as versões de um currículo com thumbnail renderizado
- O código-fonte Typst é copiado para a área de transferência ao clicar em "Copiar"
- O PDF gerado mantém o mesmo layout e conteúdo da visualização renderizada
- Tentar acessar visualização de otimização de outro usuário retorna erro 404
- Uma otimização excluída desaparece do histórico e retorna erro 404 ao ser acessada
