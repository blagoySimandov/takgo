FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /out/server  ./cmd/server && \
    CGO_ENABLED=0 go build -o /out/migrate ./cmd/migrate

FROM alpine:3.21

COPY --from=builder /out/server  /usr/local/bin/server
COPY --from=builder /out/migrate /usr/local/bin/migrate

WORKDIR /data
VOLUME  ["/data"]

EXPOSE 8080

CMD migrate up && server
