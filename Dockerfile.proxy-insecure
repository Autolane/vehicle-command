FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /tesla-http-proxy-insecure ./cmd/tesla-http-proxy-insecure

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /tesla-http-proxy-insecure /usr/local/bin/
EXPOSE 8080
ENTRYPOINT ["tesla-http-proxy-insecure"]
CMD ["--host", "0.0.0.0", "--port", "8080"]
