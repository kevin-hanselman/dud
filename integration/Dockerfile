FROM ubuntu:20.04

ARG GO_VERSION=1.21.5
ARG HUGO_VERSION=0.121.1
ARG HYPERFINE_VERSION=1.15.0

RUN apt update \
    && DEBIAN_FRONTEND=noninteractive apt install -y --no-install-recommends \
        build-essential \
        curl \
        gawk \
        git \
        graphviz \
        jq \
        parallel \
        python3-dev \
        python3-pip \
        python3-venv \
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

RUN set -x \
    && export ARCH=$(dpkg --print-architecture) \
    && rm -rvf /usr/local/go \
    && curl -fL "https://go.dev/dl/go${GO_VERSION}.linux-${ARCH}.tar.gz" \
        | tar -C /usr/local -xzvf - \
    && curl -fLO "https://github.com/gohugoio/hugo/releases/download/v${HUGO_VERSION}/hugo_extended_${HUGO_VERSION}_linux-${ARCH}.deb" \
    && dpkg --install hugo*.deb \
    && rm -v hugo*.deb \
    && curl -fLO "https://github.com/sharkdp/hyperfine/releases/download/v${HYPERFINE_VERSION}/hyperfine_${HYPERFINE_VERSION}_${ARCH}.deb" \
    && dpkg --install hyperfine*.deb \
    && rm -v hyperfine*.deb \
    && curl -f 'https://rclone.org/install.sh' | bash \
    && rm -rf /tmp/*

RUN useradd --no-log-init -m user -G sudo \
    && echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers

USER user

WORKDIR /home/user

ENV PATH=$PATH:/home/user/go/bin:/home/user/.local/bin:/usr/local/go/bin

RUN python3 -m pip install --no-cache-dir --user pipx \
    && pipx install notebook --pip-args='--no-cache-dir' \
    && pipx install dvc --pip-args='--no-cache-dir' \
    && dvc config --global core.analytics false \
    && dvc config --global core.check_update false \
    && dvc config --global cache.type symlink

# Create a directory to mount a Docker volume to. If we don't create the mount
# point now as the user, Docker will create it with root permissions when it
# creates the container.
RUN mkdir ~/dud-data

# Pre-download the Go dependencies for Dud.
COPY --chown=user go.mod go.sum ./

RUN go mod download \
    && rm go.* \
    && go install mvdan.cc/gofumpt@latest \
    && go install golang.org/x/tools/cmd/goimports@latest \
    && go install honnef.co/go/tools/cmd/staticcheck@latest \
    && rm -rf /home/user/go/src

WORKDIR /dud
