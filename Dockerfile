FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY config ./config
COPY docs ./docs
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/kanban-server ./cmd/app

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /out/kanban-server ./kanban-server
COPY --from=builder /app/config ./config

ENV CONFIG_PATH=/app/config/config.docker.yaml
EXPOSE 8003

CMD ["./kanban-server"]
