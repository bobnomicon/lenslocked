FROM node:latest AS tailwind-builder
WORKDIR /tailwind
COPY ./package.json ./package.json
COPY ./package-lock.json ./package-lock.json
COPY ./tailwind.css ./tailwind.css
COPY ./templates ./templates
RUN npm ci
RUN npm run build

FROM golang:1.26 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -v -o ./server ./cmd/server

FROM ubuntu:latest
WORKDIR /
RUN apt-get update && \
    apt-get install -y ca-certificates && \
    update-ca-certificates
COPY --from=tailwind-builder /tailwind/assets ./assets
COPY --from=builder /app/server ./server

CMD [ "./server" ]
