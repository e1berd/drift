FROM docker.io/library/node:24.22.3-slim AS web
WORKDIR /src/web
RUN npm install -g pnpm@10.34.9
COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./
# Кеш store переживает пересборки, а меньше параллельных загрузок и больше
# ретраев спасают от ETIMEDOUT на нестабильной сети.
RUN --mount=type=cache,id=drift-pnpm,target=/pnpm/store \
    pnpm install --frozen-lockfile \
      --store-dir /pnpm/store \
      --package-import-method copy \
      --network-concurrency 8 \
      --fetch-retries 5 \
      --fetch-timeout 120000
COPY web/ ./
RUN mkdir -p ../app/internal/frontend && pnpm build

FROM docker.io/library/golang:1.27.4-alpine AS app
WORKDIR /src/app
COPY app/go.mod app/go.sum ./
RUN --mount=type=cache,id=drift-gomod,target=/go/pkg/mod \
    go mod download
COPY app/ ./
COPY --from=web /src/app/internal/frontend/dist ./internal/frontend/dist
RUN --mount=type=cache,id=drift-gomod,target=/go/pkg/mod \
    --mount=type=cache,id=drift-gobuild,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/drift ./cmd/server

FROM docker.io/library/alpine:3.24
RUN adduser -D -H -u 10001 drift
COPY --from=app /out/drift /usr/local/bin/drift
USER drift
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/drift"]
