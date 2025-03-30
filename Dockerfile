FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o dns-monitor ./cmd


FROM alpine:latest
WORKDIR /app
RUN mkdir -p /data && chmod 777 /data
COPY --from=builder /app/dns-monitor .
VOLUME ["/data"]
CMD ["./dns-monitor"]