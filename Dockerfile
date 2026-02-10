FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -v -o ./server ./cmd/server/

FROM ubuntu:latest
WORKDIR /
COPY ./assets ./assets
COPY --from=builder /app/server ./server

CMD [ "./server" ]
