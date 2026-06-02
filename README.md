# forge

The forge console — the platform's flagship web UI and gateway.

- **`server/`** — Go backend (`forge-server`): API gateway, `/apps` discovery
  (from a runtime configmap), and the **live-only** platform topology (nodes from
  workload labels, edges from `forge.dev/connects-to` annotations — no declared
  catalog, no `go/manifest`).
- **`frontend/`** — Vue 3 SPA host. Renders app admin UIs via the
  `@fromforgesoftware/forge-console-plugin` contract.

## Plugin model (in progress)

The target is Grafana-style **runtime** plugin loading: the host federates each
app's console bundle (`@fromforgesoftware/<app>-console`) per the `/apps`
configmap (see `consolePluginRemote()` in forge-console-plugin). The current
frontend still bundles first-party plugins compile-time under
`frontend/src/app/features/console/plugins/` — relocating these to their service
repos as Module-Federation remotes + wiring the MF host is the remaining work.
