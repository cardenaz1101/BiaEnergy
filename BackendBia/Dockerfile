FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM alpine:3.20
WORKDIR /app
RUN adduser -D -u 10001 app && mkdir -p /app/state && chown app /app/state
COPY --from=build /out/api /app/api
COPY --chown=app:app data/readings.csv data/events.csv data/meters.csv /app/data/
RUN chmod 755 /app/data && chmod 644 /app/data/*.csv
ENV ADDR=:8080 \
    DATA_DIR=/app/data \
    STATE_PATH=/app/state/state.json
USER app
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/api/health || exit 1
ENTRYPOINT ["/app/api"]
