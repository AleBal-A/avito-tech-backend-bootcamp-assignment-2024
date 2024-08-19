FROM golang:1.22.5

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o migrator ./cmd/migrator/main.go

CMD ./migrator --db-url="postgres://test_user:qwertest@test_db:5432/test_db" --migrations-path=./migrations