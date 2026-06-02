# Foundry is a control plane: a Go backend that serves BOTH its /api surface and
# the built Vue SPA from a single image (Grafana-style). Build context is the
# repo root so the npm workspace libs resolve; go/kit is fetched as a versioned
# module (go/kit/vX.Y.Z), go/manifest still resolves via replace.

ARG GO_VERSION=1.25

# --- SPA build ---
FROM node:20-alpine AS web
WORKDIR /src
# tsconfig.base.json: the libs' tsconfigs extend it, and esbuild reads them
# while transforming ts-kit/vue-kit source (decorators need experimentalDecorators).
COPY package.json package-lock.json nx.json tsconfig.base.json /src/
COPY libs/ts-kit/package.json  /src/libs/ts-kit/package.json
COPY libs/vue-kit/package.json /src/libs/vue-kit/package.json
COPY apps/forge/frontend/package.json /src/apps/forge/frontend/package.json
# --legacy-peer-deps: the workspace's nx dev-deps have a peer-version conflict
# (@nx/storybook vs @nx/web) unrelated to the forge runtime build.
RUN npm ci --legacy-peer-deps
COPY libs/ts-kit/  /src/libs/ts-kit/
COPY libs/vue-kit/ /src/libs/vue-kit/
COPY libs/shared/  /src/libs/shared/
COPY apps/forge/frontend/ /src/apps/forge/frontend/
WORKDIR /src/apps/forge/frontend
RUN npx vite build

# --- Go backend build (server + migrator) ---
FROM golang:${GO_VERSION}-alpine AS server
WORKDIR /src
COPY apps/forge/server/ /src/apps/forge/server/
WORKDIR /src/apps/forge/server
ENV GOWORK=off
RUN CGO_ENABLED=0 go build -trimpath -o /out/server   ./cmd/server
RUN CGO_ENABLED=0 go build -trimpath -o /out/migrator ./cmd/migrator

# --- runtime ---
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=server /out/server   /app/server
COPY --from=server /out/migrator /app/migrator
COPY --from=web    /src/apps/forge/frontend/dist /app/public
ENV FOUNDRY_STATIC_DIR=/app/public
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/server"]
