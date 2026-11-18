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
ARG DRIFT_VERSION=dev
WORKDIR /src/app
COPY app/go.mod app/go.sum ./
RUN --mount=type=cache,id=drift-gomod,target=/go/pkg/mod \
    go mod download
COPY app/ ./
COPY --from=web /src/app/internal/frontend/dist ./internal/frontend/dist
RUN --mount=type=cache,id=drift-gomod,target=/go/pkg/mod \
    --mount=type=cache,id=drift-gobuild,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X github.com/e1berd/drift/app/internal/delivery.version=${DRIFT_VERSION}" -o /out/drift ./cmd/server

FROM docker.io/library/node:24.22.3-slim AS agents
ARG CLAUDE_CODE_VERSION=2.1.327
ARG CODEX_VERSION=0.169.1
ARG AGY_RELEASE=1.5.10-5253791805734912
ARG AGY_SHA512_AMD64=be7621488dd685cfd68c094fe5a4249342fb0ea2bf5eca43cf4dad7a47991faa7bc607521d677d8f141447813aa58371a7d9256beb1549493b4c6cbb069133be
ARG AGY_SHA512_ARM64=7a02312539551f049fb38363690bf315fb36704b3344b149147c4bca478edbb4096583b7057cdb1c56f80c6b50cdae55cf8e821754f6169b25c32c98e923dd9d
ARG TARGETARCH
WORKDIR /pkg
RUN case "${TARGETARCH:-amd64}" in \
      amd64) cpu=x64; triple=x86_64-unknown-linux-musl; agy=linux-x64-musl/cli_linux_x64_musl; agy_sha="${AGY_SHA512_AMD64}" ;; \
      arm64) cpu=arm64; triple=aarch64-unknown-linux-musl; agy=linux-arm-musl/cli_linux_arm64_musl; agy_sha="${AGY_SHA512_ARM64}" ;; \
      *) echo "unsupported arch ${TARGETARCH}" >&2; exit 1 ;; \
    esac \
    && node -e 'fetch(process.argv[1]).then((r) => r.ok ? r.arrayBuffer() : Promise.reject(new Error(r.statusText))).then((b) => require("fs").writeFileSync("agy.tar.gz", Buffer.from(b)))' \
      "https://storage.googleapis.com/antigravity-public/antigravity-cli/${AGY_RELEASE}/${agy}.tar.gz" \
    && echo "${agy_sha}  agy.tar.gz" | sha512sum -c - \
    && mkdir -p agy /out && tar -xzf agy.tar.gz -C agy antigravity \
    && install -m 755 agy/antigravity /out/agy \
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
COPY --from=agents /out/agy /usr/local/bin/agy
RUN ln -s /opt/codex/bin/codex /usr/local/bin/codex
COPY --from=app /out/drift /usr/local/bin/drift
ENV DRIFT_WORKSPACE_ROOT=/workspace \
    DRIFT_AGENTS_DIR=/var/lib/drift/agents \
    DISABLE_AUTOUPDATER=1 \
    USE_BUILTIN_RIPGREP=0
USER drift
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/drift"]
