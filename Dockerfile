FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download
RUN CGO_ENABLED=0 go install github.com/michurin/human-readable-json-logging/cmd/pplog@v0.0.0-20251207065006-861035fdcf92

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/main.go

FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates \
    && addgroup -S app \
    && adduser -S -G app app

COPY --from=builder /out/app ./app
COPY --from=builder /go/bin/pplog /usr/local/bin/pplog
COPY pplog.env ./pplog.env

USER app

EXPOSE 8080

ENTRYPOINT ["pplog", "./app"]
