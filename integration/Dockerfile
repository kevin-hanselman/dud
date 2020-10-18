FROM archlinux

RUN pacman -Sy --noconfirm \
        diffutils \
        sudo \
        python \
        python-pip \
        jq \
        tree \
        git \
        graphviz \
    && pacman -Scc --noconfirm

RUN groupadd sudo && useradd --no-log-init -m user -G sudo

RUN echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers

USER user

WORKDIR /home/user

ENV PATH=$PATH:/home/user/.local/bin
RUN pip install --no-cache --user dvc \
    && dvc config --global core.analytics false \
    && dvc config --global cache.type symlink

COPY duc /usr/bin/
