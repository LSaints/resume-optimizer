### Architecture

**Backend** (Go 1.26): Feature-based packages under `internal/`, not layered.
```
cmd/api/main.go          → entrypoint
internal/http/routes.go  → wires all deps manually, registers routes
pkg/data/db.go           → SQLite connection + auto-migrate on startup
pkg/ai/client.go         → Gemini AI client
pkg/text_extractor/      → PDF/DOCX text extraction
```

**Frontend** (React 19 + TS 6 + Vite 8): Standard Vite SPA.
```
src/services/api.ts  → fetch wrapper, BASE_URL = http://localhost:8080/v1
src/contexts/        → AuthContext (JWT-based)
src/pages/           → one file per route
```

### Conventions

- Go: `New<Type>` constructors, concrete structs only (no interfaces, no DI)
- Raw SQL queries with SQLite (`mattn/go-sqlite3`)
- `userID` injected into request context via `auth.AuthMiddleware`
- File naming: `snake_case.go`
- Error & log messages in **Portuguese**
- DTO files: `requests.go`, `responses.go` per package
- Frontend: one CSS module per component/page (`FileUpload.module.css`)

### Routes (Go 1.22+ pattern syntax)

```
POST   /v1/auth/login
POST   /v1/users
GET    /v1/users
GET    /v1/users/{id}
POST   /v1/resumes
GET    /v1/resumes
GET    /v1/resumes/{id}
PUT    /v1/resumes/{id}
DELETE /v1/resumes/{id}
POST   /v1/jobs
GET    /v1/jobs
GET    /v1/jobs/{id}
PUT    /v1/jobs/{id}
DELETE /v1/jobs/{id}
POST   /v1/resumes/{resumeID}/optimize
GET    /v1/resumes/{resumeID}/optimizations
GET    /v1/resumes/{resumeID}/optimizations/{optimizationID}
DELETE /v1/resumes/{resumeID}/optimizations/{optimizationID}
GET    /v1/optimizations/{optimizationID}/render
GET    /v1/optimizations/{optimizationID}/render/pdf
POST   /v1/resumes/{resumeID}/evaluate
GET    /v1/resumes/{resumeID}/evaluations
GET    /v1/resumes/{resumeID}/evaluations/{evaluationID}
```

Most routes protected by `authMiddleware` (except `POST /v1/users`, `POST /v1/auth/login`).

### Developer commands

```sh
# Backend
cd backend && go run ./cmd/api          # start server on :8080

# Frontend
cd frontend && npm run dev              # Vite dev server on :5173
cd frontend && npm run build            # tsc -b && vite build
cd frontend && npm run lint             # ESLint

# Both
docker compose up                       # backend :8080 + frontend nginx :3000
```

### Environment

Backend reads `.env` via `godotenv.Load()` (does not fail if missing).
```
GEMINI_API_KEY=...
GEMINI_MODEL=gemini-2.0-flash
FRONTEND_URL=http://localhost:5173
```

SQLite db auto-created as `app.db` in CWD. Tables + migrations run on every startup.

### Specs

Feature specs in `specs/<nnn>-<feature>/` (SPEC.md / PLAN.md / TASKS.md). Check before implementing.
