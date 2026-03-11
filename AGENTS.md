# Repository Guidelines

This project must use Superpowers workflow.

Required skills:
- brainstorming
- writing-plans
- test-driven-development
- executing-plans
- verification-before-completion

## Project Structure & Module Organization
Alpha Pulse is a monorepo. `backend/` contains the Go API: `cmd/server/` is the entrypoint, `internal/` holds engines and services, `repository/` handles persistence, `router/` wires routes, and `models/` defines GORM schemas. `frontend/` contains the Next.js app: route code in `app/`, framework entry files in `pages/`, reusable UI in `components/`, state in `store/`, API access in `services/`, and shared types in `types/`. Use `frontend/test/` for fixtures and `frontend/tests/e2e/` for Playwright specs. `docker/` holds compose files; `scripts/` contains bootstrap and dev helpers; `docs/` stores architecture and API notes.

## Build, Test, and Development Commands
`./scripts/bootstrap.sh` installs Go and Node dependencies.
`./scripts/dev.sh` starts backend and frontend against local MySQL and Redis.
`USE_DOCKER_DEPS=1 ./scripts/dev.sh` starts MySQL and Redis in Docker, then runs the apps locally.
`cd docker && docker compose up --build` runs the full stack in containers.
`cd backend && go run ./cmd/server` starts only the API.
`cd backend && go test ./...` runs all backend tests.
`cd frontend && npm run dev|build|lint|test|test:e2e` covers local dev, production build, linting, unit tests, and browser tests.

## Coding Style & Naming Conventions
Format all Go code with `gofmt`; keep package names lowercase and place tests next to code as `*_test.go`. Follow the existing backend layering: orchestration in `internal/service/`, HTTP logic in `internal/handler/`, and database access in `repository/`. Frontend code uses TypeScript with 2-space indentation, PascalCase for components, camelCase for store actions and helpers, and `@/` imports for app modules. Prefer descriptive filenames such as `MarketSnapshotLoader.tsx` or `signal_service_cache_test.go`.

## Testing Guidelines
Backend tests use Go's built-in `testing` package. Frontend unit tests use Vitest and Testing Library; end-to-end coverage uses Playwright in `frontend/tests/e2e/*.spec.ts`. Add or update tests whenever behavior changes, especially around snapshots, signals, streaming, caching, and API responses. Run the affected test suite before opening a PR.

## Commit & Pull Request Guidelines
Recent history favors short imperative commit subjects, usually with prefixes such as `feat:`, `fix:`, or `refactor:`. Keep commits focused; for example, `feat: add market snapshot stream fallback`. PRs should include a short summary, linked issue or task, test evidence, and screenshots or recordings for UI changes. Call out schema, environment, or API contract changes explicitly.

## Security & Configuration Tips
Store secrets in `backend/.env` and `frontend/.env.local`; never commit them. Keep `frontend/.next/`, `frontend/test-results/`, `frontend/playwright-report/`, and `*.log` out of reviews unless a task specifically requires them.
