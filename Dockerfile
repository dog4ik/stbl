FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -v -o /app/stbl

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
  sqlite3 ca-certificates && \
  rm -rf /var/lib/apt/lists/*

WORKDIR /usr/local/bin

COPY --from=builder /app/stbl .

VOLUME ["/data"]
ENV DATABASE_URL=/data/database.sqlite

EXPOSE 4208
CMD ["./stbl"]
