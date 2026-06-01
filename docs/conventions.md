# Coding Conventions

## Formatting & Style

- Use `gofmt` / `goimports` — all code must be formatted before committing
- Use tabs for indentation (Go standard)

## Naming

| Element | Convention | Example |
|---|---|---|
| Exported types, interfaces, functions | `PascalCase` | `Service`, `NewHandler` |
| Unexported types, functions, locals | `camelCase` | `service`, `rowFromDB` |
| JSON tags, SQL columns, env vars | `snake_case` | `stream_url`, `DATABASE_URL` |
| Repository methods | verb-first | `Create`, `GetByID`, `List`, `Update`, `Delete` |
| Sentinel errors | `Err` prefix | `ErrNotFound`, `ErrEmailTaken` |
| Domain row structs | `XxxRow` suffix | `StationRow`, `UserRow` |
| Input structs | `XxxInput` suffix | `CreateInput`, `CreateUserInput` |
| HTTP DTOs | `XxxRequest` / `XxxResponse` | `CreateStationRequest`, `StationResponse` |

## Types & Interfaces

- `Service` interface in `service.go`, unexported `service` struct implementing it
- `Repository` interface in `repository.go`, unexported `sqlcRepository` struct implementing it
- Constructors return the interface type: `func NewService(repo Repository) Service`
- Domain structs and input types are defined in `repository.go`, not `service.go`
- Use `*string`, `*int64`, `*bool` for nullable/optional fields in both DTOs and domain types
- `PATCH` updates use pointer fields — `nil` means "keep the existing value"

## Error Handling

- Sentinel errors are package-level `var` using `errors.New()`:
  ```go
  var ErrNotFound = errors.New("station not found")
  ```
- Services translate `pgx.ErrNoRows` (and similar infra errors) to domain sentinel errors
- Handlers use `errors.Is()` to map sentinel errors to HTTP status codes:
  ```go
  if errors.Is(err, ErrNotFound) {
      response.NotFound(c, "station not found")
      return
  }
  ```
- Wrap errors with context: `fmt.Errorf("upsert batch: %w", err)`
- Never swallow errors silently

## Validation

- Handler-level: Gin `binding` struct tags handle format and presence:
  ```go
  Email string `json:"email" binding:"required,email"`
  ```
- Service-level: explicit checks returning sentinel errors for business rules:
  ```go
  if len(password) < 8 {
      return ErrWeakPassword
  }
  ```
- Do not duplicate — binding tags for format, services for business rules

## Logging

- Use `log/slog` everywhere — not `log`, not `fmt.Println`
- Structured key-value pairs: `slog.Error("sync failed", "err", err)`
- HTTP request logging is handled by `mw.LoggerStructured()` — do not add per-handler logs
- HTTP handler errors are logged centrally by `httperr.Handle`; handlers should return errors instead of logging them
- Internal server errors return a generic client message; log records carry the real `err`, `request_id`, route, status, and request metadata
- Do not log `Authorization`, cookies, access/refresh tokens, passwords, raw request bodies, or database URLs

## Testing

### File placement

- Tests live alongside source: `handler_test.go`, `service_test.go`
- White-box (same package): `package auth` — access to unexported types
- Black-box (exported API only): `package radiobrowser_test`

### Test doubles

**Stub service** (handler tests):
```go
type stubService struct {
    registerOut AuthOutput
    registerErr error
}
func (s *stubService) Register(_ context.Context, _ RegisterInput) (AuthOutput, error) {
    return s.registerOut, s.registerErr
}
```

**Fake repository** (service tests):
```go
type fakeRepo struct {
    upserted []UpsertInput
    err      error
}
func (f *fakeRepo) UpsertBatch(_ context.Context, stations []UpsertInput) (int64, error) {
    if f.err != nil { return 0, f.err }
    f.upserted = append(f.upserted, stations...)
    return int64(len(stations)), nil
}
```

**httptest server** (external client tests):
```go
srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    _, _ = w.Write([]byte(singleStationJSON))
}))
defer srv.Close()
client := radiobrowser.NewClient(srv.URL)
```

### Handler test skeleton

```go
func TestCreateHandler(t *testing.T) {
    gin.SetMode(gin.TestMode)
    stub := &stubService{createOut: StationRow{ID: 1, Name: "Test"}}
    h := NewHandler(stub)
    r := gin.New()
    r.POST("/stations", h.Create)

    payload := `{"name":"Test","stream_url":"http://stream.test"}`
    req := httptest.NewRequest(http.MethodPost, "/stations", bytes.NewBufferString(payload))
    req.Header.Set("Content-Type", "application/json")
    res := httptest.NewRecorder()

    r.ServeHTTP(res, req)
    require.Equal(t, http.StatusCreated, res.Code)
}
```

### Assertions

- Use `require` (testify) when failure should abort the test immediately
- Use `t.Fatalf` / `t.Errorf` where testify is not already used in the file
- Match the style already present in the file you're editing

### Coverage expectations

- Every new endpoint needs at least one handler test
- Non-trivial service logic gets a service test
- Test both success and error paths (sentinel error → HTTP status)
- Test pagination (empty list, meta fields)
- Test external API clients: empty response, transient errors, retry exhaustion
