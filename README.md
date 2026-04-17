# plextrac-go

[![test](https://github.com/cresco/plextrac-go/actions/workflows/test.yml/badge.svg)](https://github.com/cresco/plextrac-go/actions/workflows/test.yml)
[![lint](https://github.com/cresco/plextrac-go/actions/workflows/lint.yml/badge.svg)](https://github.com/cresco/plextrac-go/actions/workflows/lint.yml)
[![security](https://github.com/cresco/plextrac-go/actions/workflows/security.yml/badge.svg)](https://github.com/cresco/plextrac-go/actions/workflows/security.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/cresco/plextrac-go.svg)](https://pkg.go.dev/github.com/cresco/plextrac-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/cresco/plextrac-go)](https://goreportcard.com/report/github.com/cresco/plextrac-go)

Go SDK and CLI for the [Plextrac](https://plextrac.com) pentest reporting platform.

- **SDK**: `github.com/cresco/plextrac-go/pkg/plextrac` — typed client for flaws, clients, reports, assets, writeups, attachments, templates, tags, users, and exports.
- **CLI**: `plextrac` — auth, CRUD, and bulk import of findings from a JSON file into a report.

## Install

### SDK

```bash
go get github.com/cresco/plextrac-go/pkg/plextrac@latest
```

### CLI

Download a release binary from the [Releases page](https://github.com/cresco/plextrac-go/releases), or build from source:

```bash
go install github.com/cresco/plextrac-go/cmd/plextrac@latest
```

## SDK quick start

```go
package main

import (
    "context"
    "log"

    "github.com/cresco/plextrac-go/pkg/plextrac"
)

func main() {
    c, err := plextrac.New(
        "https://tenant.kevlar.plextrac.com",
        plextrac.WithPasswordAuth("alice@example.com", "hunter2", plextrac.EnvMFA{}),
    )
    if err != nil {
        log.Fatal(err)
    }
    ctx := context.Background()
    iter := c.Flaws.List(ctx, "1234", "5678", plextrac.ListOpts{})
    for iter.Next(ctx) {
        f := iter.Value()
        log.Printf("%s  %s  %s", f.ID, f.Severity, f.Title)
    }
    if err := iter.Err(); err != nil {
        log.Fatal(err)
    }
}
```

See [examples/](examples/) for more.

## CLI usage

```bash
export PLEXTRAC_URL="https://tenant.kevlar.plextrac.com"
export PLEXTRAC_USERNAME="alice@example.com"
export PLEXTRAC_PASSWORD="hunter2"
export PLEXTRAC_MFA="123456"
export PLEXTRAC_CLIENT_ID="1234"
export PLEXTRAC_REPORT_ID="5678"

plextrac flaws list
plextrac flaws get <flaw-id>
plextrac flaws import findings.json              # upsert based on custom_fields.audit_id
plextrac flaws import findings.json --mode create
plextrac flaws import findings.json --dry-run

plextrac clients list
plextrac reports list --client 1234
plextrac assets list --client 1234 --report 5678

plextrac auth login                              # print a JWT for scripting
plextrac exports start --format pdf              # kick off a PDF export
```

Run `plextrac --help` or `plextrac <command> --help` for full options.

### Configuration

| Env var              | Description                                    | Default |
| :------------------- | :--------------------------------------------- | :------ |
| `PLEXTRAC_URL`       | Tenant base URL (required)                     | —       |
| `PLEXTRAC_USERNAME`  | Username for password auth                     | —       |
| `PLEXTRAC_PASSWORD`  | Password                                       | —       |
| `PLEXTRAC_MFA`       | TOTP/MFA code (if MFA enabled)                 | —       |
| `PLEXTRAC_API_KEY`   | API key (bypasses password/MFA)                | —       |
| `PLEXTRAC_TOKEN`     | Pre-obtained JWT                               | —       |
| `PLEXTRAC_CLIENT_ID` | Default client ID for commands                 | —       |
| `PLEXTRAC_REPORT_ID` | Default report ID                              | —       |
| `LOG_LEVEL`          | `error`, `warn`, `info`, `debug`               | `info`  |

Secrets are read from env or stdin only, never from argv.

## Features

- Password + MFA, API key, and raw JWT auth
- Typed models for every resource (no `map[string]any` in user code)
- Lazy `Iter[T]` pagination
- Context-aware, rate-limited, retried HTTP (respects `Retry-After`)
- Bulk upsert with bounded concurrency
- Pluggable `logr.Logger`
- `mdhtml` sub-package for Plextrac-safe markdown → HTML
- Findings JSON importer (schema in [docs/findings-schema.md](docs/findings-schema.md))

## Security

- TLS ≥ 1.2 by default; HTTPS-only unless `WithAllowHTTP`
- Input IDs validated against `^[0-9a-zA-Z_-]+$` before URL interpolation
- `Authorization` and password headers never logged
- Password memory is zeroed after token exchange
- Response bodies are **not** included in error messages by default — opt in with `WithDebugBody(true)`
- Supply chain: `govulncheck` + `trivy` on every push, daily `dependabot`, signed releases via `goreleaser`

## Development

```bash
make test         # go test -race -coverprofile=coverage.out ./...
make lint         # golangci-lint
make bench        # benchmarks
make build        # compile the CLI
make update       # go get -u ./... && go mod tidy
```

Go version is pinned in [`.github/go/Dockerfile`](.github/go/Dockerfile) and consumed by every CI job.

## Versioning

Semantic versioning via [`go-semantic-release`](https://github.com/go-semantic-release/semantic-release). Commit messages follow [Conventional Commits](https://www.conventionalcommits.org/).

## License

MIT — see [LICENSE](LICENSE).
