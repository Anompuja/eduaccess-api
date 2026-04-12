# ── Build stage ──────────────────────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Cache deps separately from source
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Generate swagger docs before building
RUN go install github.com/swaggo/swag/cmd/swag@latest && \
    swag init -g cmd/main.go --output docs

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/main.go

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM alpine:3.21

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Jakarta

WORKDIR /app
COPY --from=builder /server .

EXPOSE 8080

CMD ["./server"]
