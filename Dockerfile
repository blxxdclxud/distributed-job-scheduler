FROM golang:1.24-alpine AS deps

# Create a working directory
WORKDIR /go/src/app

# Copy only the dependency files
COPY go.mod go.sum ./

# Download dependencies - this layer will be cached unless go.mod/go.sum changes
RUN go mod download

FROM golang:1.24-alpine AS builder

WORKDIR /go/src/app

# Copy cached dependencies from the deps stage
COPY --from=deps /go/pkg /go/pkg
COPY --from=deps /go/src/app/go.mod /go/src/app/go.sum ./

# Copy source code files
COPY . .

# Build the applications
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