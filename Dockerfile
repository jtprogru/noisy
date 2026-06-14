FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY noisy.go ./

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o noisy .

FROM alpine:3.19

RUN addgroup -S noisy && adduser -S -G noisy noisy

WORKDIR /opt/noisy

COPY --from=builder /app/noisy .

USER noisy

# Конфиг не вшивается в образ — монтируйте его как volume:
#   docker run -v ./config.json:/opt/noisy/config.json noisy --config config.json
ENTRYPOINT ["/opt/noisy/noisy"]

CMD ["--config", "config.json"]
