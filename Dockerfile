# ===== Builder =====
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o pr-service ./cmd/app

# ===== Final image =====
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/pr-service ./pr-service
COPY ./configs ./configs
COPY ./migrations ./migrations


CMD ["./pr-service"]