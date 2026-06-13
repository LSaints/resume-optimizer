# SPEC.md

## Feature

Otimização de Currículo com IA

---

## Objetivo

Permitir que o usuário autenticado otimize um currículo existente utilizando a API do Google AI Studio (Gemini), gerando uma versão estruturada em Typst ajustada ao nível ATS da vaga alvo.

---

## Requisitos de negócio

- O usuário deve estar autenticado para solicitar otimização
- O usuário deve selecionar um currículo existente (previamente enviado) e uma vaga existente (previamente cadastrada)
- O sistema deve montar um prompt contendo:
  - System prompt definindo o papel da IA como otimizadora de currículos
  - Texto extraído do currículo selecionado
  - Descrição da vaga selecionada
- O resultado da IA deve ser um currículo reescrito em linguagem Typst
- O nível de maturidade ATS (nível de exigência) da descrição da vaga deve ser respeitado na reescrita
- O currículo otimizado deve ser armazenado e associado ao currículo original e à vaga
- O usuário pode consultar o histórico de otimizações de um currículo
- O usuário pode visualizar o conteúdo Typst gerado em uma otimização específica

---

## Restrições

- Apenas currículos e vagas pertencentes ao usuário logado podem ser usados
- A chave da API do Google AI Studio deve ser configurada via variável de ambiente
- O prompt enviado à IA não deve conter informações sensíveis além do próprio currículo e vaga do usuário
- O system prompt deve ser versionado e definido em código, nunca vindo de entrada do usuário
- O currículo otimizado deve ser armazenado como texto (Typst), nunca como arquivo binário

---

## Critérios de Aceitação

- Um usuário autenticado consegue otimizar um currículo existente com base em uma vaga existente
- O sistema retorna o currículo otimizado em formato Typst
- O nível ATS da vaga é refletido na profundidade e detalhamento do currículo gerado
- O histórico de otimizações de um currículo lista todas as versões geradas
- Uma otimização específica pode ser consultada individualmente
- Tentativa de otimizar com currículo ou vaga de outro usuário retorna erro 404
- Tentativa de otimizar sem autenticação retorna erro 401
- Se a chave da API não estiver configurada, o sistema retorna erro 500
- Se a API do Google retornar erro, o sistema retorna erro 502
