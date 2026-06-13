# SPEC.md

## Feature

Avaliação de Currículo por ATS

---

## Objetivo

Permitir que o usuário autenticado avalie a compatibilidade de um currículo existente com uma vaga específica, gerando uma pontuação de 0 a 10 baseada em critérios ATS (Applicant Tracking System), com detalhamento dos pontos fortes e oportunidades de melhoria.

---

## Requisitos de negócio

- O usuário deve estar autenticado para solicitar uma avaliação
- O usuário deve selecionar um currículo existente (previamente enviado) e uma vaga existente (previamente cadastrada)
- O sistema deve analisar o currículo contra a descrição da vaga e atribuir uma pontuação de 0 a 10
- A pontuação deve refletir critérios ATS: correspondência de palavras-chave, relevância de experiência, adequação de habilidades, formatação e legibilidade
- O sistema deve retornar além da pontuação:
  - Um resumo textual da avaliação
  - Um detalhamento com pontos fortes e pontos a melhorar
- O usuário pode consultar o histórico de avaliações de um currículo
- O usuário pode visualizar uma avaliação específica em detalhe

---

## Restrições

- Apenas currículos e vagas pertencentes ao usuário logado podem ser usados
- A pontuação deve ser estritamente numérica entre 0 e 10 (suportando uma casa decimal)
- A chave da API do Google AI Studio deve estar configurada
- O prompt enviado à IA não deve conter informações sensíveis além do próprio currículo e vaga do usuário
- O system prompt deve ser versionado e definido em código, nunca vindo de entrada do usuário
- Cada avaliação gera um novo registro — não há atualização de avaliações existentes

---

## Critérios de Aceitação

- Um usuário autenticado consegue avaliar um currículo existente com base em uma vaga existente e recebe uma pontuação de 0 a 10
- O sistema retorna a pontuação acompanhada de um resumo e um detalhamento da análise
- O histórico de avaliações de um currículo lista todas as avaliações realizadas
- Uma avaliação específica pode ser consultada individualmente com todos os detalhes
- Tentativa de avaliar com currículo ou vaga de outro usuário retorna erro 404
- Tentativa de avaliar sem autenticação retorna erro 401
- Se a chave da API não estiver configurada, o sistema retorna erro 500
- Se a API do Google retornar erro, o sistema retorna erro 502
