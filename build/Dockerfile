FROM golang:1.21-alpine as builder

WORKDIR /app

COPY go.work go.work.sum ./
COPY download/go.mod download/go.sum ./download/
COPY scan/go.mod scan/go.sum ./scan/
COPY db/go.mod db/go.sum ./db/

RUN go mod download
COPY . .

RUN go build -o progdl ./download/cmd/download.go

