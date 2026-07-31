FROM golang:1.23-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG CGO_ENABLED=0
RUN CGO_ENABLED=${CGO_ENABLED} GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata curl
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /src/config ./config
EXPOSE 8080
HEALTHCHECK --interval=15s --timeout=3s --retries=3 CMD curl -f http://localhost:8080/health || exit 1
ENTRYPOINT ["./server"]