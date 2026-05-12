FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o pulse-api ./cmd/api/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/pulse-api .
COPY --from=builder /app/etc/api.yaml ./etc/api.yaml
EXPOSE 8000 8001
CMD ["./pulse-api", "-f", "etc/api.yaml"]
