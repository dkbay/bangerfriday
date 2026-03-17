FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o /app/banger-friday ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite

WORKDIR /app
COPY --from=builder /app/banger-friday .

EXPOSE 8080

CMD ["./banger-friday"]