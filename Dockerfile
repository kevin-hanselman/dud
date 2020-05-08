FROM archlinux

RUN pacman -Sy --noconfirm base-devel git go hyperfine && pacman -Scc --noconfirm

RUN go get -u golang.org/x/lint/golint \
    && go get -u github.com/kisielk/godepgraph

ENV PATH=$PATH:~/go/bin/

RUN mkdir /src

WORKDIR /src

COPY go.mod go.sum ./

RUN go mod download \
    && rm *
