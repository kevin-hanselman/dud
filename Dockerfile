FROM golang:alpine

RUN apk update && apk add --no-cache git

RUN go get -u golang.org/x/lint/golint

RUN mkdir /src

WORKDIR /src

COPY go.mod go.sum ./

RUN go mod download

RUN rm *
