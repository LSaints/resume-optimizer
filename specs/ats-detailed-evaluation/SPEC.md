# SPEC.md

## Feature

Detalhamento Avançado de Pontuação ATS

---

## Objetivo

Enriquecer o resultado da avaliação ATS com notas desagregadas por critério, palavras-chave encontradas e ausentes, e recomendações acionáveis, permitindo que o usuário entenda exatamente como a pontuação foi calculada e quais ações tomar para melhorar a compatibilidade do currículo.

---

## Requisitos de negócio

- O sistema deve retornar, além da pontuação geral, um detalhamento com notas individuais para cada critério ATS:
  - Correspondência de palavras-chave
  - Compatibilidade técnica
  - Experiência profissional
  - Impacto e resultados
  - Legibilidade ATS
- O sistema deve listar as palavras-chave da vaga que foram encontradas no currículo
- O sistema deve listar as palavras-chave da vaga que estão ausentes no currículo
- O sistema deve gerar recomendações acionáveis para aumentar a compatibilidade ATS
- O sistema deve exibir visualmente o detalhamento dos critérios em formato de barra de progresso ou gráfico de barras horizontais
- O usuário deve conseguir identificar rapidamente os critérios com melhor e pior desempenho
- As palavras-chave encontradas e ausentes devem ser exibidas como tags ou listas com distinção visual (encontradas em verde, ausentes em vermelho)
- As recomendações devem ser exibidas como uma lista ordenada numerada
- A interface deve deixar claro que a pontuação máxima de cada subcritério corresponde ao seu peso na nota final

---

## Restrições

- A pontuação máxima da avaliação geral continua sendo 10.0
- A soma dos subtotais dos critérios deve totalizar a pontuação geral
- Os pesos de cada critério devem ser exibidos ao lado da nota:
  - Correspondência de palavras-chave: peso 30% (máx 3.0)
  - Compatibilidade técnica: peso 25% (máx 2.5)
  - Experiência profissional: peso 20% (máx 2.0)
  - Impacto e resultados: peso 15% (máx 1.5)
  - Legibilidade ATS: peso 10% (máx 1.0)
- Avaliações antigas (criadas antes desta feature) não precisam ser retroativamente enriquecidas — os novos campos podem vir vazios para registros existentes
- As palavras-chave e recomendações devem ser geradas pela IA, não inventadas pelo sistema
- Cada recomendação deve ser específica e acionável, não genérica

---

## Critérios de Aceitação

- Ao avaliar um currículo, o sistema retorna breakdown com 5 subnotas que somam a nota geral
- O sistema retorna lista de palavras-chave encontradas no currículo
- O sistema retorna lista de palavras-chave ausentes no currículo
- O sistema retorna lista de recomendações acionáveis
- A página de resultado exibe o breakdown com indicadores visuais de progresso (barras)
- A página exibe palavras-chave encontradas em destaque verde e ausentes em destaque vermelho
- A página exibe recomendações como lista numerada
- Avaliações existentes (sem breakdown) continuam sendo exibidas sem erro — os novos campos aparecem vazios
