# SPEC.md

## Feature

Autenticação JWT

---

## Objetivo

Proteger os endpoints da API com autenticação baseada em token JWT, garantindo que apenas usuários autenticados possam acessar recursos do sistema.

---

## Requisitos de negócio

- O sistema deve permitir que um usuário cadastrado realize login informando email e senha
- O sistema deve gerar um token JWT válido quando o login for bem-sucedido
- O sistema deve rejeitar tentativas de login com credenciais inválidas
- O sistema deve validar o token JWT em toda requisição a endpoints protegidos
- O sistema deve recusar requisições sem token ou com token inválido/expirado
- O sistema deve retornar erros claros e padronizados quando a autenticação falhar
- A senha do usuário nunca deve ser armazenada ou transmitida em texto puro

---

## Restrições

- Deve utilizar apenas a biblioteca padrão do Go e as dependências já existentes no projeto, com exceção de pacotes para JWT e hash de senha
- Deve seguir a arquitetura em camadas já estabelecida no projeto
- Deve manter o padrão de nomenclatura e organização existente
- As mensagens e comentários devem permanecer em português
- O token JWT deve conter apenas informações não sensíveis (nunca a senha)
- Endpoints de login não devem exigir autenticação

---

## Critérios de Aceitação

- Um usuário com credenciais válidas consegue fazer login e recebe um token JWT
- Um usuário com credenciais inválidas recebe erro 401 Unauthorized
- Uma requisição sem token a um endpoint protegido recebe erro 401 Unauthorized
- Uma requisição com token inválido recebe erro 401 Unauthorized
- Uma requisição com token expirado recebe erro 401 Unauthorized
- Uma requisição com token válido a um endpoint protegido é processada com sucesso
- A senha é armazenada com hash bcrypt no banco de dados
- O hash da senha é gerado no momento da criação do usuário
