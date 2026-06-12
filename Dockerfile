FROM golang:1.22-bookworm AS go-build

WORKDIR /src
COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ ./
RUN CGO_ENABLED=0 go build -o /bin/server ./cmd/server
RUN CGO_ENABLED=0 go build -o /bin/worker-pipeline ./cmd/worker-pipeline

FROM node:22-bookworm AS render-deps

WORKDIR /render
COPY worker-render/package.json ./
RUN npm install --omit=dev
COPY worker-render/ ./

FROM node:22-bookworm

RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    ffmpeg \
    chromium \
    fonts-liberation \
    libnss3 \
    libatk-bridge2.0-0 \
    libdrm2 \
    libxkbcommon0 \
    libgbm1 \
    ca-certificates \
  && rm -rf /var/lib/apt/lists/*

ENV PUPPETEER_EXECUTABLE_PATH=/usr/bin/chromium

COPY --from=go-build /bin/server /bin/worker-pipeline /usr/local/bin/
COPY --from=render-deps /render /app/worker-render
COPY migrations /app/migrations

ENV RENDER_DIR=/app/worker-render
ENV MIGRATIONS_DIR=/app/migrations
ENV STORAGE_DIR=/data/storage
ENV FFMPEG_PATH=ffmpeg

WORKDIR /app
