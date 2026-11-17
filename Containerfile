FROM docker.io/library/node:24.22.3-slim AS web
WORKDIR /src/web
RUN npm install -g pnpm@10.34.9
COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile
COPY web/ ./
RUN mkdir -p ../app/internal/frontend && pnpm build

FROM docker.io/library/golang:1.27.4-alpine AS app
WORKDIR /src/app
COPY app/go.mod app/go.sum ./
RUN go mod download
COPY app/ ./
COPY --from=web /src/app/internal/frontend/dist ./internal/frontend/dist
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/drift ./cmd/server

FROM docker.io/library/alpine:3.24
RUN adduser -D -H -u 10001 drift
COPY --from=app /out/drift /usr/local/bin/drift
USER drift
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/drift"]
