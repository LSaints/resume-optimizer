# SPEC.md

## Feature

Tela de Avaliação ATS de Currículo

---

## Objetivo

Prover uma interface web para que o usuário autenticado possa avaliar a compatibilidade de um currículo com uma vaga usando critérios ATS, visualizar o resultado com pontuação e detalhes, e consultar o histórico de avaliações realizadas.

---

## Requisitos de negócio

- O usuário autenticado deve poder selecionar um currículo e uma vaga e solicitar uma avaliação ATS
- O sistema deve exibir o resultado da avaliação com: pontuação de 0 a 10, resumo textual e detalhamento com pontos fortes e oportunidades de melhoria
- O usuário deve poder visualizar o histórico de avaliações de um currículo específico
- O usuário deve poder acessar uma avaliação específica para ver seus detalhes completos
- A interface deve informar o usuário sobre o tempo de processamento da avaliação (pode levar até 60 segundos)
- Mensagens de erro devem ser exibidas em português de forma clara e amigável

---

## Restrições

- O frontend deve se comunicar exclusivamente com a API existente em `/v1/*`
- Não deve haver lógica de negócio no frontend — apenas validação de formulário e apresentação
- O token JWT deve ser obtido do `localStorage` e enviado automaticamente
- Score deve ser exibido com no máximo uma casa decimal
- Design deve ser responsivo (mobile e desktop)

---

## Critérios de Aceitação

- Um usuário autenticado consegue selecionar um currículo e uma vaga e solicitar uma avaliação
- O resultado da avaliação exibe: pontuação numérica de 0 a 10, resumo textual e detalhamento
- O histórico de avaliações de um currículo lista todas as avaliações realizadas com pontuação e resumo
- Uma avaliação específica pode ser consultada com todos os detalhes
- Tentar avaliar com currículo ou vaga inexistentes exibe mensagem de erro amigável
- Se a chave da API não estiver configurada, o sistema exibe mensagem clara ao usuário
- Um usuário não autenticado é redirecionado para o login ao acessar rotas protegidas
