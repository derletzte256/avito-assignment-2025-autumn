FROM golang:1.25-bookworm

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

CMD ["go", "test", "-v", "./..."]
