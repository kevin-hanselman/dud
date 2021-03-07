FROM archlinux

RUN pacman -Sy --noconfirm --needed \
        base-devel \
        diffutils \
        git \
        # Make sure glibc is up-to-date. Sometimes it isn't,
        # and Go tools (e.g. go vet) can have problems.
        glibc \
        go \
        graphviz \
        hugo \
        hyperfine \
        jq \
        jupyter \
        parallel \
        python \
        python-pip \
        rclone \
        sudo \
        time \
        tokei \
        tree \
        xxhash \
    && pacman -Scc --noconfirm

RUN groupadd sudo && useradd --no-log-init -m user -G sudo

RUN echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers

USER user

WORKDIR /home/user

# Create a directory to mount a Docker volume to. If we don't create the mount
# point now as the user, Docker will create it with root permissions when it
# creates the container.
RUN mkdir ~/dud-data

ENV PATH=$PATH:/home/user/go/bin:/home/user/.local/bin

RUN pip install --no-cache --user dvc \
    && dvc config --global core.analytics false \
    && dvc config --global cache.type symlink

# Pre-download the Go dependencies for Dud.
COPY --chown=user go.mod go.sum ./

RUN go mod download \
    && rm go.{mod,sum} \
    && go install mvdan.cc/gofumpt@latest \
    && go install golang.org/x/lint/golint@latest \
    && rm -rf /home/user/go/src

WORKDIR /dud
