# template-go-vue

A GitHub template repository for a contract-first Go + Vue application in the
[JorisJonkers-dev](https://github.com/JorisJonkers-dev) estate: a Go API generated from an OpenAPI
document with ogen, a Vue SPA whose client is generated from the same document with
`@hey-api/openapi-ts`, Postgres through goose and sqlc, and one distroless image in which the Go
binary serves the embedded web build. A repository generated from it builds, tests and runs its
end-to-end suite green with no edits.

The estate's decisions on this shape live in `JorisJonkers-dev/workspace`:
[ADR-0016](https://github.com/JorisJonkers-dev/workspace/blob/main/docs/decisions/ADR-0016-go-vue-contract-first-standard.md)
(Go + Vue, contract first) and
[ADR-0018](https://github.com/JorisJonkers-dev/workspace/blob/main/docs/decisions/ADR-0018-repository-templates.md)
(the repository templates). [`docs/blueprints/`](docs/blueprints/README.md) holds the architecture the
code follows, and says which parts the template leaves for later. A Go service without a web app
starts from [`template-go`](https://github.com/JorisJonkers-dev/template-go) instead. Hygiene files
(licence, security policy, `CODEOWNERS`, Renovate, editor config, release flow) come from
[`repo-template`](https://github.com/JorisJonkers-dev/repo-template).

## What is in it

| Path | What it is |
|------|------------|
| `openapi/v1/openapi.yaml` | The contract, written first; both sides are generated from it |
| `openapi/` | Redocly `recommended-strict` and Spectral with the OWASP ruleset; vacuum runs beside them |
| `internal/platform/oas/` | The ogen server, generated (`go generate`) |
| `web/src/infrastructure/api/` | The TypeScript client, generated: types, a fetch SDK that zod-validates every response, vue-query options |
| `db/migrations/` | goose migrations, embedded and applied by the binary at startup; linted by squawk |
| `db/queries/`, `internal/platform/pg/queries/` | SQL, and the Go sqlc generates from it |
| `internal/notes/` | The sample bounded context: `domain` (pure), `app` (use cases), `adapters/persistence` and `adapters/web` |
| `internal/platform/` | Postgres (`pg`, with `pgtest` for a migrated database per test), `httpapi` (assembles the ogen server), `httpx`, `webui`, and `itest`, which runs the whole service over real HTTP for the integration tests |
| `internal/assembly/` | Wires each context's use cases and adapters on the store; `main` and `itest` both build the service here |
| `internal/server/` | Probes, `/api/` to the API, everything else to the SPA, graceful drain |
| `cmd/template-go-vue/` | The binary; reads `DATABASE_URL`, `ADDR` (default `:8080`) and `DEV_USER`, migrates, and serves what `assembly` builds |
| `web/` | The Vue 3 SPA, and `embed.go`, which embeds its build (`web/dist`) in the binary |
| `web/src/app/contract.ts` | Checks every answer a web test stubs against the contract, so a page test cannot pass on an answer the server would never send |
| `web/tests/e2e/` | Playwright with axe, on a desktop and a phone, against the built binary; each journey names its stages with `test.step` |
| `scripts/coverage.sh` | The Go unit and integration coverage gates |
| `mise.toml`, `Taskfile.yml` | The pinned toolchain, and every command: `task` lists them |
| `Dockerfile` | Builds the web app, embeds it, ships a static binary on `distroless/static:nonroot` |
| `.github/workflows/ci.yml` | One parallel job per `task check` task; the e2e suite as one build of the binary, then `task e2e:ci` per device and shard in the Playwright image with a Postgres service, merged into one HTML report; one job for `docker build`, the quarantine check, and the mutation report on pull requests; `Pipeline Complete`, the one required check, passes when all but the two reports do |
| `.github/workflows/release.yml`, `release-please-config.json` | release-please, as in the rest of the estate |
| `deploy/template-go-vue.project.yml` | A [deploy-kit](https://github.com/JorisJonkers-dev/deploy-kit) Project Intent: one Process, an authenticated host, a Postgres edge |

The sample resource is a list of notes, there only to show every layer once. Replace it.

## The loop

```bash
mise install    # the pinned toolchain: Go, Node, pnpm, task, linters
task check      # lint, gen:check, Go unit, Go integration and web tests with coverage, build, secret scan
task e2e        # builds, starts the compose Postgres, runs Playwright with axe
task dev        # Postgres, the API on :8080, Vite with hot reload on :5173
```

Tests follow the estate
[testing standard](https://github.com/JorisJonkers-dev/workspace/blob/main/docs/testing-standards.md):
each test checks one behaviour and is named as it. Go unit tests need nothing outside the process,
and plain `go test ./...` runs only them. Tests that need Postgres sit behind the `integration`
build tag, in `*_integration_test.go`. A context's behaviours are checked there through real HTTP
and real Postgres with `itest`, as `internal/notes/behaviour_integration_test.go` does.
`task test:integration` runs them, and needs Docker for testcontainers.

`task lint` runs `testlint` to hold test names and tables to the standard, and an ESLint rule that
makes a long Playwright journey name its stages. Tests never retry: a flaky test is fixed, or
skipped with `quarantine: <issue-url> until YYYY-MM-DD` for at most 14 days, which CI checks.

**Changing the API** is always the same motion: edit `openapi/v1/openapi.yaml`, run `task gen` (which
lints the contract first), then fix what no longer compiles on either side. `task gen:check` fails
when committed generated code is stale, and CI runs it.

**Coverage gates.** Go unit tests reach every statement, branch and condition (gobco) and every
function, excluding generated code, test tooling, and what needs Postgres. The integration run
reaches every function (`scripts/coverage.sh`). Web tests reach every line, branch, function and
statement, excluding generated code and `main.ts` (`web/vite.config.ts`).

**Identity.** The API requires an `X-User-Id` header, which the platform's forward-auth sets at the
edge (`audience: authenticated` in the project file). On a local run nothing sits in front of the
binary, so `DEV_USER` fills the header in. Never set `DEV_USER` in a deployment.

## Use it

1. **Create the repository** with "Use this template" on GitHub.
2. **Rename the module.** Replace `github.com/JorisJonkers-dev/template-go-vue` with the new module
   path in `go.mod`, `.golangci.yml` (the goimports prefix and the depguard rule) and every import.
3. **Rename the application.** Rename `cmd/template-go-vue/` to `cmd/<name>`, then replace
   `template-go-vue` in `Taskfile.yml` (`APP`), `Dockerfile`, `package.json`, `openapi/package.json`,
   `web/package.json`, `web/index.html`, `web/playwright.config.ts`, `openapi/v1/openapi.yaml`,
   `.github/workflows/ci.yml`, `scripts/coverage.sh` and `release.yml`.
4. **Rename the project file.** Move `deploy/template-go-vue.project.yml` to
   `deploy/<name>.project.yml` and change `project`, `owner`, the Application `id`, the Process
   `name` and `image`, and the host.
5. **Reset release state**: delete `CHANGELOG.md` if present and keep `.release-please-manifest.json`
   at `0.0.0`.
6. `mise install && task check && task e2e`.

Ruleset and project boarding are described in [`CONTRIBUTING.md`](CONTRIBUTING.md) and
[`VERSIONING.md`](VERSIONING.md); `add-to-project.yml` should be deleted in a private repository,
as its header explains.
