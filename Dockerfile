# syntax=docker/dockerfile:1.7

# --- Frontend build ---
FROM node:22-alpine AS frontend-builder
WORKDIR /frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend ./
RUN npm run build

# --- Go build ---
FROM golang:1.25-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY services ./services

# Drop the dev placeholder and embed the real Vite build instead.
RUN rm -rf ./services/api-gateway/internal/handlers/web
COPY --from=frontend-builder /frontend/dist ./services/api-gateway/internal/handlers/web

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /out/server \
    ./cmd/server

# --- Runtime ---
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/server /server
EXPOSE 8080
USER nonroot
ENTRYPOINT ["/server"]
