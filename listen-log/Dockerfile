FROM golang:1.14.2

ENV GO111MODULE=on
WORKDIR /server
COPY go.mod .
RUN go mod download

COPY . .
RUN go build -o listen-log/listen-log listen-log/main.go