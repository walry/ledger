# syntax=docker/dockerfile:1

# ---------- 构建阶段 ----------
FROM golang:1.24-bookworm AS builder

WORKDIR /src

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOFLAGS=-mod=mod

# 先复制依赖清单，利用镜像层缓存。
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY pkg ./pkg

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-s -w" -o /out/ledger ./cmd/api

# ---------- 运行阶段 ----------
FROM debian:12-slim

ENV TZ=Asia/Shanghai \
    HTTP_ADDR=:8080 \
    UPLOAD_DIR=/data/uploads

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl tzdata \
    && rm -rf /var/lib/apt/lists/* \
    && groupadd --system --gid 10001 ledger \
    && useradd --system --uid 10001 --gid ledger --create-home ledger \
    && mkdir -p /data/uploads \
    && chown -R ledger:ledger /data

WORKDIR /app

COPY --from=builder /out/ledger /app/ledger

USER ledger

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=20s --retries=3 \
    CMD curl -fsS http://127.0.0.1:8080/healthz || exit 1

ENTRYPOINT ["/app/ledger"]
