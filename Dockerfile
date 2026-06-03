# forge is the console control plane: a Go backend that serves BOTH its /api
# surface and the built Vue SPA from a single image (Grafana-style). Standalone
# repo: frontend/ + server/ at the root; the SPA consumes the published
# @fromforgesoftware/* packages from GitHub Packages, the server the published
# go-kit. Build context is the repo root.

ARG GO_VERSION=1.25

# --- SPA build (published @fromforgesoftware/* from GitHub Packages) ---
FROM node:22-alpine AS web
WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json ./
# @fromforgesoftware/* live in GitHub Packages (auth required even when public);
# the token is a build secret so it never lands in a layer.
RUN --mount=type=secret,id=npmtoken \
    printf '@fromforgesoftware:registry=https://npm.pkg.github.com\n//npm.pkg.github.com/:_authToken=%s\n' "$(cat /run/secrets/npmtoken)" > .npmrc \
 && npm ci \
 && rm -f .npmrc
COPY frontend/ ./
# No sibling kit checkouts in the image — resolve the published packages.
ENV FORGE_USE_PUBLISHED_KIT=1
RUN npx vite build

# --- Go backend build (server + migrator) ---
FROM golang:${GO_VERSION}-alpine AS server
WORKDIR /src/server
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./
ENV GOWORK=off
RUN CGO_ENABLED=0 go build -trimpath -o /out/server   ./cmd/server
RUN CGO_ENABLED=0 go build -trimpath -o /out/migrator ./cmd/migrator

# --- runtime ---
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=server /out/server   /app/server
COPY --from=server /out/migrator /app/migrator
COPY --from=web    /src/frontend/dist /app/public
ENV FOUNDRY_STATIC_DIR=/app/public
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/server"]
