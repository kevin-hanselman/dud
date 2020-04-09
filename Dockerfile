FROM golang:alpine

RUN apk add --no-cache git

RUN go get -u golang.org/x/lint/golint \
    && go get -u github.com/kisielk/godepgraph

RUN mkdir /src

WORKDIR /src

COPY go.mod go.sum ./

RUN go mod download \
    && rm *
