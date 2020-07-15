FROM archlinux

RUN pacman -Sy --noconfirm diffutils python && pacman -Scc --noconfirm

RUN useradd -m user

USER user

WORKDIR /home/user
