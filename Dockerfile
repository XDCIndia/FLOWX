# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /tmp/fluxa-api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /tmp/fluxa-worker ./cmd/worker

# Runtime stage
FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /tmp/fluxa-api .
COPY --from=builder /tmp/fluxa-worker .
COPY db/migrations /app/db/migrations
EXPOSE 3000

# Run api by default; override with: docker run flowx ./fluxa-worker
CMD ["./fluxa-api"]
