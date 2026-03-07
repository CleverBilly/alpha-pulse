FROM golang:1.22-alpine AS builder

WORKDIR /app

# 先复制 go.mod 以利用 Docker 缓存。
COPY backend/go.mod ./go.mod
RUN go mod download

# 复制后端源码并编译。
COPY backend ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /output/server ./cmd/server

FROM alpine:3.20

WORKDIR /app
RUN adduser -D appuser
COPY --from=builder /output/server /app/server

USER appuser
EXPOSE 8080
CMD ["/app/server"]
