FROM golang:1.22-alpine3.20 AS builder
WORKDIR /app
COPY . /app

RUN go build

FROM alpine:3.20

COPY --from=builder /app/cert-manager-webhook-inwx /
ENTRYPOINT ["/cert-manager-webhook-inwx"]
