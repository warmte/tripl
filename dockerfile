
FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /tripl ./tripl.go

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata poppler-utils

WORKDIR /app

COPY --from=builder /tripl /app/tripl

CMD ["/app/tripl"]
