# Frontend (Next.js) image. Build context is the repo root (see docker-compose.staging.yml).
# BACKEND_URL and NEXT_PUBLIC_* are baked in at build time (Next rewrites + client bundle).
FROM node:22-alpine AS builder
RUN corepack enable && corepack prepare pnpm@10 --activate
WORKDIR /app
COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile
COPY frontend/ ./

ARG BACKEND_URL=http://backend-service:8090
ARG NEXT_PUBLIC_API_URL=/api/v1
ARG NEXT_PUBLIC_WEBHOOK_URL=
ARG NEXT_PUBLIC_WS_URL=
ARG NEXT_PUBLIC_WS_SAME_ORIGIN=true
ENV BACKEND_URL=$BACKEND_URL \
    NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL \
    NEXT_PUBLIC_WEBHOOK_URL=$NEXT_PUBLIC_WEBHOOK_URL \
    NEXT_PUBLIC_WS_URL=$NEXT_PUBLIC_WS_URL \
    NEXT_PUBLIC_WS_SAME_ORIGIN=$NEXT_PUBLIC_WS_SAME_ORIGIN \
    NEXT_TELEMETRY_DISABLED=1
# No `pnpm prune --prod`: next start needs `typescript` (a devDependency) to load next.config.ts.
RUN pnpm build

# shortcut: ships all node_modules incl. dev deps; switch to output:"standalone" if image size matters.
FROM node:22-alpine
WORKDIR /app
ENV NODE_ENV=production NEXT_TELEMETRY_DISABLED=1 PORT=3000 HOSTNAME=0.0.0.0
COPY --from=builder --chown=node:node /app ./
USER node
EXPOSE 3000
CMD ["node", "node_modules/next/dist/bin/next", "start", "-p", "3000", "-H", "0.0.0.0"]
