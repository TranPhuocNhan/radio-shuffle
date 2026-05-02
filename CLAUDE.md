# CLAUDE.md

Behavioral guidelines to reduce common LLM coding mistakes. Merge with project-specific instructions as needed.

**Tradeoff:** These guidelines bias toward caution over speed. For trivial tasks, use judgment.

## 1. Think Before Coding

**Don't assume. Don't hide confusion. Surface tradeoffs.**

Before implementing:

- State your assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them - don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

## 2. Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

## 3. Surgical Changes

**Touch only what you must. Clean up only your own mess.**

When editing existing code:

- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it - don't delete it.

When your changes create orphans:

- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.

The test: Every changed line should trace directly to the user's request.

## 4. Goal-Driven Execution

**Define success criteria. Loop until verified.**

Transform tasks into verifiable goals:

- "Add validation" → "Write tests for invalid inputs, then make them pass"
- "Fix the bug" → "Write a test that reproduces it, then make it pass"
- "Refactor X" → "Ensure tests pass before and after"

For multi-step tasks, state a brief plan:

```
1. [Step] → verify: [check]
2. [Step] → verify: [check]
3. [Step] → verify: [check]
```

Strong success criteria let you loop independently. Weak criteria ("make it work") require constant clarification.

---

**These guidelines are working if:** fewer unnecessary changes in diffs, fewer rewrites due to overcomplication, and clarifying questions come before implementation rather than after mistakes.

---

## 5. Project: Radio Shuffle — Go / Gin / SQLC

### Stack

- **API**: Gin (HTTP framework)
- **DB queries**: SQLC (type-safe SQL → Go code generation)
- **Validation**: `binding` struct tags (Gin's built-in validator)
- **DB**: PostgreSQL (SQLite for tests)
- **Auth**: JWT via `golang-jwt/jwt/v5`
- **Migrations**: golang-migrate (SQL up/down files)
- **Testing**: Go standard `testing` + `testify` + `httptest`

### Architecture — Modular Monolith

```
cmd/
└── server/
    └── main.go                 ← Entry point, wires all modules
internal/
├── module/
│   ├── user/                   ← Auth & user management
│   │   ├── handler.go
│   │   ├── service.go
│   │   ├── repository.go       ← Interface (port)
│   │   ├── repository_sqlc.go  ← SQLC-backed implementation
│   │   ├── dto.go
│   │   └── module.go           ← NewModule() — DI wiring
│   ├── station/                ← Same layout per module
│   ├── track/
│   ├── playlist/
│   └── stream/
└── platform/
    ├── config/                 ← Env/config loading
    ├── database/               ← DB connection + pool
    ├── middleware/              ← Auth, logging, recovery
    └── response/               ← Shared error/success helpers
db/
├── migration/                  ← golang-migrate SQL files
├── query/                      ← SQLC query files (one per module)
├── schema.sql                  ← Full schema reference for SQLC
└── sqlc.yaml                   ← SQLC config
pkg/
└── dbsqlc/                     ← SQLC generated code (do not edit)
```

### Module Isolation Rules

- Modules **never** import each other's packages directly
- Cross-module communication uses interfaces defined in the consuming module
- Each module exposes `NewModule()` that returns a struct with `RegisterRoutes(*gin.RouterGroup)`
- `cmd/server/main.go` is the only place that wires modules together

### SQLC Workflow

```
1. Write SQL in db/query/{module}.sql with SQLC annotations
2. Run `sqlc generate` to produce type-safe Go in pkg/dbsqlc/
3. Repository implementations wrap the generated Queries struct
```

### Vibe Coding Workflow

Always follow this loop for any new feature:

```
plan-feature → generate-crud → review-code → generate-test → go test
```

1. **Plan first** — use `plan-feature` skill; no code until plan is approved
2. **Scaffold** — use `generate-crud` to create all layers at once
3. **Review** — use `review-code` before considering a feature done
4. **Test** — use `generate-test`, then run `go test ./...`; don't ship untested endpoints

### Hard Rules

- All SQL lives in `db/query/*.sql` — never write inline SQL strings in Go code
- Every new endpoint needs at least one test
- Never put business logic in a handler function
- Never call the DB directly from a handler — go through the service → repository chain
- Never edit files in `pkg/dbsqlc/` — they are generated by `sqlc generate`
- Never import one module from another — use interfaces for cross-module deps
