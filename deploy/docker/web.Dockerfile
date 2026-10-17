# syntax=docker/dockerfile:1

# ---------- 构建阶段 ----------
FROM node:24-bookworm AS builder

WORKDIR /app

# 先装依赖，利用镜像层缓存。
COPY frontend/package.json frontend/package-lock.json ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci

COPY frontend/ ./

# 生产环境通过 nginx 反代同源 /api，无需额外注入 VITE_API_BASE。
RUN npm run build

# ---------- 运行阶段 ----------
FROM nginx:stable-alpine

COPY deploy/nginx/default.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -q --spider http://127.0.0.1/healthz || exit 1
