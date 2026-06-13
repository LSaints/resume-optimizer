### Architecture

Clean Architecture (Layered) (Backend)

```
cmd/api/
internal/
  apresentation/   -> HTTP layer (handlers, routes, middleware)
  application/     -> Use cases (services, DTOs)
  domain/          -> Business entities
  infrastructure/  -> Data access (db, repositories)
```

### Stack

- Go 1.26 (stdlib `net/http`, `database/sql`, `encoding/json`)
- SQLite via `mattn/go-sqlite3`
- UUID via `google/uuid`


### Existing Conventions

- Concrete structs with constructor functions (`New<Type>`)
- No interfaces, no DI framework
- Raw SQL queries
- Versioned routes: `/v1/<resource>`
- Folder naming: `apresentation/` (Portuguese spelling)
- Files: `snake_case.go`
- DTOs: `<Action><Entity>Request` / `<Entity>Response`
- Messages in Portuguese
- Dependencies wired manually in `routes.go`
- `userID` injetado no contexto da requisição via `AuthMiddleware`

---