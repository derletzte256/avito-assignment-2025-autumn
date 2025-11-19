# syntax=docker/dockerfile:1.7
FROM golang:1.25-bookworm AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /out/server ./cmd/app

FROM gcr.io/distroless/base-debian12:nonroot AS runner

WORKDIR /app
COPY --from=builder /out/server /app/server

EXPOSE 8080

ENTRYPOINT ["/app/server"]
