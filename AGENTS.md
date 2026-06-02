# forge — the forge platform console (web UI + gateway)

Flagship console (split from the monorepo; renamed from 'foundry'). The GitHub
repo is `fromforgesoftware/forge`; the old monorepo is archived as
`fromforgesoftware/forge-monorepo`.

## Layout
- `server/` — Go (`forge-server`): API gateway, `/apps` discovery (from a runtime
  configmap), and the **live-only** platform topology (nodes from workload labels,
  edges from `forge.dev/connects-to` annotations — no declared catalog, no go/manifest).
- `frontend/` — Vue 3 SPA host; consumes ts-kit + vue-kit + forge-console-plugin.

## Commands
- Server: `cd server && go build ./... && go vet ./... && go test ./...`
- Frontend: `cd frontend && npm install --legacy-peer-deps && npm run build`

## Conventions / Boundaries
- One-line conventional commits, ≤72 chars. REST is JSON:API. NEVER commit secrets.
  No dependabot. Don't hand-edit generated code.

## In progress
- Runtime Module Federation plugin loader + relocating per-app console plugins
  (`frontend/src/app/features/console/plugins/*`) into their service repos as MF
  remotes. The contract + federation preset live in forge-console-plugin.
