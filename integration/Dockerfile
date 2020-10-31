FROM archlinux

RUN pacman -Sy --noconfirm --needed \
        base-devel \
        diffutils \
        git \
        go \
        graphviz \
        hyperfine \
        jq \
        python \
        python-pip \
        rclone \
        sudo \
        tree \
        xxhash \
    && pacman -Scc --noconfirm

RUN groupadd sudo && useradd --no-log-init -m user -G sudo

RUN echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers

USER user

WORKDIR /home/user

ENV PATH=$PATH:/home/user/go/bin:/home/user/.local/bin

RUN pip install --no-cache --user dvc \
    && dvc config --global core.analytics false \
    && dvc config --global cache.type symlink

# Pre-download the Go dependencies for Dud.
COPY go.mod go.sum ./

RUN go mod download \
    && rm go.{mod,sum} \
    && go get golang.org/x/tools/cmd/goimports \
    && go get golang.org/x/lint/golint \
    && rm -rf /home/user/go/src

WORKDIR /dud
