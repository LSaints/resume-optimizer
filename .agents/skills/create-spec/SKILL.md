---

name: spec-driven-development
description: Generate structured software specifications from feature descriptions. Use this when users describe a feature, module, API, system, workflow, business process, or software requirement and need implementation planning. Creates SPEC.md, PLAN.md and TASKS.md documents inside specs/<00-feature-name>/.
--------------------------------------

Transform feature descriptions into implementation-ready specifications.

This skill generates a specification package consisting of:

```text
specs/<00-feature-name>/
├── SPEC.md
├── PLAN.md
└── TASKS.md
```

The process happens in three stages:

1. Requirements Analysis (SPEC.md)
2. Technical Planning (PLAN.md)
3. Task Decomposition (TASKS.md)

---

# FEATURE ANALYSIS

When a user describes a feature, system, API, workflow, module, automation, platform, dashboard, integration, or business process:

Analyze and extract:

* Business goals
* User needs
* Functional requirements
* Non-functional requirements
* Constraints
* Acceptance criteria

The user description is the source of truth.

Infer missing details only when necessary to remove ambiguity.

Do not invent business rules that contradict the request.

---

# STEP 1 - GENERATE SPEC.md

Generate a business-oriented specification.

Output:

```markdown
# SPEC.md

## Feature

Feature name

---

## Objetivo

Short business objective

---

## Requisitos de negócio

- Requirement
- Requirement

---

## Restrições

- Restriction
- Restriction

---

## Critérios de Aceitação

- Acceptance criteria
- Acceptance criteria
```

## SPEC Rules

Focus on:

* Business behavior
* Rules
* Constraints
* Validation requirements

Avoid:

* Technologies
* Frameworks
* Database decisions
* Architecture discussions

SPEC.md must answer:

"What should the system do?"

not

"How should it be built?"

---

# STEP 2 - GENERATE PLAN.md

Generate a technical implementation plan.

Output:

```markdown
# PLAN.md

## Arquitetura

Architecture overview

---

## Estrutura Técnica

### Entidades

- Entity

### Serviços

- Service

### DTOs

- DTO

### Repositórios

- Repository

### Endpoints

- Endpoint

---

## Configuração

Configuration requirements

---

## Decisões Técnicas

- Decision
- Decision
```

## PLAN Rules

Translate business requirements into implementation structures.

Include when relevant:

* Entities
* Repositories
* Services
* DTOs
* Endpoints
* Events
* Background jobs
* Integrations
* External APIs

PLAN.md must answer:

"How should the feature be implemented?"

Avoid implementation code.

---

# STEP 3 - GENERATE TASKS.md

Generate implementation tasks.

Tasks must be:

* Small
* Incremental
* Independently verifiable
* Ordered logically

Output:

```markdown
# TASKS.md

## Task 1

Description

### Objetivo

Expected outcome

### Validação

How to verify
```

Each task must contain:

* Description
* Objective
* Validation

---

## Task Breakdown Guidelines

Prefer:

* Entity creation
* Repository creation
* DTO creation
* Service implementation
* Endpoint implementation
* Validation rules
* Integration implementation
* End-to-end validation

Avoid:

* Massive tasks
* Multiple responsibilities in the same task

The final task should always validate the complete flow.

Example:

```markdown
## Task N

Testar fluxo completo

### Objetivo

Validar a feature ponta a ponta

### Validação

- Cenário 1 funciona
- Cenário 2 funciona
- Cenário 3 funciona
```

---

# DIRECTORY GENERATION

Generate a slug for the feature name.

Examples:

```text
Gerenciamento de Usuários
→ user-management

Portal do Cliente
→ customer-portal

Solicitação de Materiais
→ material-request

Monitoramento de Equipamentos
→ equipment-monitoring
```

Create:

```text
specs/<slug>/
├── SPEC.md
├── PLAN.md
└── TASKS.md
```

---

# QUALITY REQUIREMENTS

The generated documents must:

* Be implementation-ready
* Be understandable by humans
* Be understandable by AI agents
* Have clear traceability between files
* Avoid ambiguity
* Avoid contradictory requirements
* Maintain consistency between SPEC, PLAN and TASKS

---

# AGENT EXECUTION FLOW

For every request:

1. Analyze feature description.
2. Identify business objectives.
3. Extract requirements.
4. Extract restrictions.
5. Generate SPEC.md.
6. Generate PLAN.md.
7. Generate TASKS.md.
8. Create output structure:

```text
specs/<00-feature-name>/
├── SPEC.md
├── PLAN.md
└── TASKS.md
```

9. Never generate implementation code.
10. Generate specification artifacts only.

---

# EXPECTED OUTPUT

Given:

```text
Criar CRUD de usuários.
```

Generate:

```text
specs/user-management/
├── SPEC.md
├── PLAN.md
└── TASKS.md
```

# PROJECT CONTEXT ANALYSIS

Before generating PLAN.md and TASKS.md, inspect the existing project.

The goal is to ensure generated specifications follow the project's established architecture, conventions and patterns.

---

## Required Analysis

Analyze the repository structure and identify:

### Architecture

* Clean Architecture
* Vertical Slice Architecture
* Layered Architecture
* MVC
* Modular Monolith
* Microservices
* Feature Based Structure
* Other patterns

---

### Technology Stack

Identify:

* Programming language
* Frameworks
* ORM
* Database
* Frontend framework
* Testing framework
* Messaging systems
* Authentication strategy

Examples:

```text
ASP.NET Core
Entity Framework Core
PostgreSQL
Blazor
React
RabbitMQ
Redis
```

---

### Existing Conventions

Identify conventions already used by the project.

Examples:

```text
Repositories
Services
Use Cases
CQRS
MediatR
Result Pattern
Unit Of Work
DTOs
Custom Exceptions
```

New plans must follow existing conventions whenever possible.

Avoid introducing new architectural patterns without strong justification.

---

### Existing Folder Structure

Analyze current folders and modules.

Examples:

```text
src/
├── Domain
├── Application
├── Infrastructure
└── Api
```

or

```text
src/
├── Features
│   ├── Users
│   └── Orders
```

The generated PLAN.md must align with the discovered structure.

---

### Existing Naming Standards

Identify naming patterns.

Examples:

```text
UserService
IUserRepository
CreateUserRequest
UpdateUserRequest
```

Generated plans must follow the same naming conventions.

---

### Existing Endpoint Patterns

Analyze API organization.

Examples:

```text
/api/users
/api/orders
```

or

```text
/v1/users
/v1/orders
```

Generated endpoints must remain consistent.

---

## Architecture Consistency Rules

When generating PLAN.md:

Prefer:

* Existing patterns
* Existing abstractions
* Existing dependency injection strategy
* Existing validation strategy
* Existing exception strategy

Avoid:

* Introducing new frameworks
* Introducing new architectural styles
* Replacing existing patterns

Unless explicitly requested by the user.

---

## Missing Context Handling

If project context cannot be determined:

1. Infer architecture from available files.
2. Follow industry-standard conventions.
3. Document assumptions inside PLAN.md.

Example:

```markdown
## Assumptions

- Project follows Repository pattern
- Project uses Entity Framework Core
- Project follows layered architecture
```

---

## Specification Enrichment

After analyzing the project, enrich PLAN.md with:

```markdown
## Existing Context

### Architecture

Layered Architecture

### Stack

- ASP.NET Core
- Entity Framework Core
- PostgreSQL

### Existing Conventions

- Repository Pattern
- Service Layer
- DTOs
```

This section should appear before the implementation plan.

---

## Task Generation Using Context

TASKS.md must be generated using discovered project conventions.

Example:

If the project uses:

```text
Features/
 ├── Users
 └── Orders
```

Generate tasks such as:

```text
Create User Feature
Create User DTOs
Create User Service
Create User Endpoints
```

Instead of:

```text
Create Controllers
Create Repository Layer
```

if those concepts do not exist in the project.

---

## Agent Priority Order

When generating specifications:

1. User requirements
2. Existing project architecture
3. Existing project conventions
4. Existing naming standards
5. Industry best practices

Never invert this priority order.

The generated specification must fit naturally into the existing codebase.


With all files fully populated and internally consistent.
