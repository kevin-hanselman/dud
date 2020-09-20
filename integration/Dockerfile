FROM archlinux

RUN pacman -Sy --noconfirm diffutils sudo python jq && pacman -Scc --noconfirm

RUN groupadd sudo && useradd -m user -G sudo

RUN echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers

USER user

WORKDIR /home/user
