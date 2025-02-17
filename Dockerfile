# Build stage
FROM golang:1.24 AS builder
WORKDIR /app
COPY . .
ENV CGO_ENABLED=0
RUN go mod tidy && go build -o bot .

# Minimal runtime image
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/bot .
CMD ["/root/bot"]
