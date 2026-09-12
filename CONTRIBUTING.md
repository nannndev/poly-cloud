# Contributing to Poly Cloud

Thanks for considering it. Bug reports, documentation fixes, and provider testing
are as welcome as code.

## Getting the stack running

You need Docker and Docker Compose. Nothing else is installed on your machine.

```sh
git clone https://github.com/nannndev/poly-cloud
cd poly-cloud
cp .env.example .env
make dev
```

Before starting, replace `TOKEN_ENC_KEY` and `RCLONE_RC_PASS` in `.env` with
random values. The defaults are placeholders and are not safe to run with.

| Service | Address |
|---|---|
| Frontend | http://localhost:3000 |
| API health | http://localhost:8080/healthz |
| Postgres | localhost:5432 |

If port 3000 is already taken on `::1`, open `http://127.0.0.1:3000` instead —
that origin is already in the default `CORS_ORIGINS`.

## Running the checks

```sh
make test      # gofmt + go vet + go test, via a Go container
make fmt       # gofmt -w, same way
```

The backend checks run inside a container, so a local Go toolchain is optional.

For the frontend:

```sh
cd apps/frontend
npx vue-tsc --noEmit -p .nuxt/tsconfig.json
npx nuxt build
```

Please make sure the backend tests and the frontend typecheck both pass before
opening a pull request.

## Repository layout

```
apps/backend    Go API — HTTP handlers, storage service, rclone engine, index
apps/frontend   Nuxt 4 app — the interface you use day to day
apps/landing    Public landing page, static, deployed separately
docs/           Architecture, data model, API spec, and ADRs
```

`docs/` is worth reading before a larger change. `docs/08-adr.md` records why
several things are the way they are — including decisions that look odd without
the context.

## Conventions

**Language.** User-facing strings are English. Code comments in the Go backend
and the Nuxt app are written in Indonesian, matching the existing files — follow
whichever the surrounding file uses.

**Comments.** Explain why something is done, not what the line does. A comment
that restates the code is noise; one that records a constraint or a rejected
alternative saves the next person an hour.

**Schema changes.** Migrations in `apps/backend/migrations/` run automatically,
but only when the Postgres volume is empty. Changing the schema after the volume
exists needs `docker compose down -v` (which destroys local data) or a manual
migration.

**Security.** The rclone RC API is all-or-nothing: anything that can reach it can
read every stored credential. Never expose the `rclone` service outside the
compose network — see ADR-011.

## Pull requests

- Branch off `main`.
- Keep the change focused. Two unrelated fixes are easier to review as two pull requests.
- Say what you tested. "Uploaded a 2 GB file to S3 and to Drive" tells a reviewer more than "works".
- If the change affects behaviour someone relies on, update `docs/` in the same PR.

## Reporting bugs

Useful reports include the steps to reproduce, what you expected, what happened,
and which providers were connected. Logs from `make logs` around the failure
help. Redact tokens and access keys before pasting anything.

## Security issues

Do not open a public issue for a vulnerability. Report it privately through
GitHub's security advisory form on the repository instead.
