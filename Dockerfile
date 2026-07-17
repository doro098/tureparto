# ============================================================
# tureparto - Multi-stage Docker Build
# ============================================================
# Servidor de webhook para WhatsApp Cloud API (Meta).
# ============================================================

# --- Etapa 1: Compilación ---
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Cache de dependencias
COPY go.mod go.sum ./
RUN go mod download

# Código fuente
COPY . .

# Compilar binario estático
# modernc.org/sqlite es Go puro, no necesita CGO
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /tureparto .

# --- Etapa 2: Imagen final ---
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

RUN adduser -D -h /app -s /sbin/nologin appuser

WORKDIR /app

COPY --from=builder /tureparto /app/tureparto

# Directorio para la base de datos SQLite
RUN mkdir -p /app/data && chown -R appuser:appuser /app

USER appuser

EXPOSE 3000

VOLUME ["/app/data"]

ENTRYPOINT ["/app/tureparto"]
