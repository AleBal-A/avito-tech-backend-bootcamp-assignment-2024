FROM golang:1.22.5

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o estate-service ./cmd/estate-service/main.go
RUN go build -o migrator ./cmd/migrator/main.go

ENV CONFIG_PATH="./config/cfg.yaml"

CMD ["./estate-service"]