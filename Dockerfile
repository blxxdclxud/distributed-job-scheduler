FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/server ./cmd/server/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/worker ./cmd/worker/main.go

FROM alpine:latest AS server
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/bin/server /app/server
ENTRYPOINT ["/app/server"]

FROM alpine:latest AS worker
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/bin/worker /app/worker
ENTRYPOINT ["/app/worker"]