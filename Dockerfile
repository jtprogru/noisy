FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY noisy.go ./

RUN go build -o noisy .

FROM alpine:3.19

WORKDIR /opt/noisy

COPY --from=builder /app/noisy .
COPY config.json .

ENTRYPOINT ["/opt/noisy/noisy"]

CMD ["--config", "config.json"]

