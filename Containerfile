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

FROM docker.io/library/node:24.22.3-slim AS agents
ARG CLAUDE_CODE_VERSION=2.1.327
ARG CODEX_VERSION=0.169.1
ARG TARGETARCH
WORKDIR /pkg
RUN case "${TARGETARCH:-amd64}" in \
      amd64) cpu=x64; triple=x86_64-unknown-linux-musl ;; \
      arm64) cpu=arm64; triple=aarch64-unknown-linux-musl ;; \
      *) echo "unsupported arch ${TARGETARCH}" >&2; exit 1 ;; \
    esac \
    && npm pack --silent "@anthropic-ai/claude-code-linux-${cpu}-musl@${CLAUDE_CODE_VERSION}" \
      "@openai/codex@${CODEX_VERSION}-linux-${cpu}" \
    && mkdir -p claude codex /out/codex \
    && tar -xzf anthropic-ai-claude-code-*.tgz -C claude \
    && tar -xzf openai-codex-*.tgz -C codex \
    && install -m 755 claude/package/claude /out/claude \
    && cp -a "codex/package/vendor/${triple}/." /out/codex/

FROM docker.io/library/alpine:3.24
RUN apk add --no-cache git ca-certificates libstdc++ libgcc ripgrep bubblewrap \
    && adduser -D -H -u 10001 drift \
    && install -d -o drift -g drift /workspace /var/lib/drift/agents
COPY --from=agents /out/claude /usr/local/bin/claude
COPY --from=agents /out/codex /opt/codex
RUN ln -s /opt/codex/bin/codex /usr/local/bin/codex
COPY --from=app /out/drift /usr/local/bin/drift
ENV DRIFT_WORKSPACE_ROOT=/workspace \
    DRIFT_AGENTS_DIR=/var/lib/drift/agents \
    DISABLE_AUTOUPDATER=1 \
    USE_BUILTIN_RIPGREP=0
USER drift
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/drift"]
