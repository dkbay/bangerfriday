FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o /app/banger-friday ./cmd/main.go

FROM alpine:3.20

RUN apk --no-cache add ca-certificates sqlite
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app
COPY --from=builder /app/banger-friday .
RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

CMD ["./banger-friday"]
