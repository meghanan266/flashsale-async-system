FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY processor/main.go ./processor/
RUN CGO_ENABLED=0 GOOS=linux go build -o processor ./processor/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/processor .

CMD ["./processor"]