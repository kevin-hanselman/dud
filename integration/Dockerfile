FROM ubuntu:20.04

RUN apt update && apt install -y software-properties-common \
    && add-apt-repository ppa:longsleep/golang-backports \
    && apt update \
    && apt install -y \
        curl \
        git \
        golang-go \
        graphviz \
        jq \
        parallel \
        python3-pip \
        sudo \
        tree \
        unzip \
    && rm -rf /var/lib/apt/lists/* \
    && ln -sv "$(which python3)" /usr/local/bin/python \
    && git config --system --add safe.directory /dud
# safe.directory allows Git to parse and operate on the mounted source repo,
# even if it is owned by a different user (which is likely, unless the UIDs
# happen to match).
# See: https://git-scm.com/docs/git-config#Documentation/git-config.txt-safedirectory

COPY integration/install_hyperfine_deb.sh ./
RUN ./install_hyperfine_deb.sh

RUN curl https://rclone.org/install.sh | bash

RUN useradd --no-log-init -m user -G sudo \
    && echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers

USER user

WORKDIR /home/user

# Create a directory to mount a Docker volume to. If we don't create the mount
# point now as the user, Docker will create it with root permissions when it
# creates the container.
RUN mkdir ~/dud-data

ENV PATH=$PATH:/home/user/go/bin:/home/user/.local/bin

RUN pip install --no-cache --user dvc notebook \
    && dvc config --global core.analytics false \
    && dvc config --global core.check_update false \
    && dvc config --global cache.type symlink

# Pre-download the Go dependencies for Dud.
COPY --chown=user go.mod go.sum ./

RUN go mod download \
    && rm go.* \
    && go install mvdan.cc/gofumpt@latest \
    && go install golang.org/x/tools/cmd/goimports@latest \
    && go install honnef.co/go/tools/cmd/staticcheck@latest \
    && rm -rf /home/user/go/src

WORKDIR /dud
