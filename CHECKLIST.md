Project Checklists
==================

This file collects practical checklists for development, CI, migrations, and releases.

1) Pre-commit / Local checks

- Run formatting and vetting:

```bash
gofmt -w .
go vet ./...
```

- Run unit tests fast:

```bash
go test ./... -short
```

- Run linting (if installed):

```bash
golangci-lint run
```

2) Pre-push checklist

- All tests pass:

```bash
go test ./... -v
```

- No formatting diffs:

```bash
test -z "$(gofmt -l .)"
```

- Run migrations locally (if needed):

```bash
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f migrations/0001_create_news.sql
```

3) CI checklist (what CI runs)

- Checkout code
- Setup Go toolchain
- Start Postgres service
- Wait for DB healthy
- Run migrations from `migrations/`
- Run `go vet` and `gofmt` checks
- Run `go test ./...`

4) PR checklist (before requesting review)

- [ ] Branch built locally: `go build ./...`
- [ ] Unit tests added / updated for new behavior
- [ ] No panic in library code; errors returned
- [ ] DB queries parameterized
- [ ] `.env` not modified or committed
- [ ] Update `migrations/` for DB schema changes

5) Migration checklist

- Create SQL migration in `migrations/` and name it sequentially
- Test migration locally against a fresh DB
- Add rollback migration if necessary (or document steps)
- Ensure migration run in CI before integration tests

6) Release / Deploy checklist

- Bump version / tag commit
- Run full test suite and integration tests
- Build binary: `go build -o bin/server cmd/server/main.go`
- (Optional) Build Docker image and push to registry
- Deploy to target environment and run smoke tests

7) Security checklist

- Secrets stored in env or secrets manager (do not commit `.env`)
- Rotate `JWT_SECRET` and database credentials periodically
- Use least privilege DB user for app access

8) Testing patterns & tips

- Unit tests: mock repositories (interfaces) using testify/mock or hand-written mocks
- Repo tests: use `github.com/DATA-DOG/go-sqlmock` to validate SQL and results
- Integration tests: gate with `INTEGRATION=1` env var or build tag

9) Useful commands

```bash
# run all tests
go test ./... -v

# run only unit tests (if tests are partitioned by naming)
go test ./... -run Test.* -v

# format project
gofmt -w .

# vet
go vet ./...
```

Keep this file updated when processes or CI change.
