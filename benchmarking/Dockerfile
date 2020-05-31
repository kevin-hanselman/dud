FROM archlinux

ARG USER=user
RUN pacman -Sy --noconfirm base-devel git go python python-pip && pacman -Scc --noconfirm

RUN groupadd -r user && useradd --no-log-init --create-home -g user $USER
USER $USER
ENV HOME /home/$USER

RUN pip install dvc --no-cache --user

COPY . /src
WORKDIR /src
ENV GOBIN $HOME/.local/bin
RUN go install .

ENV PATH=$PATH:$HOME/.local/bin
WORKDIR $HOME
