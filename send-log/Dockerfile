FROM golang:1.14.2

ENV GO111MODULE=on
WORKDIR /server
COPY go.mod .
RUN go mod download

COPY . .
RUN go build -o send-log/send-log send-log/main.go