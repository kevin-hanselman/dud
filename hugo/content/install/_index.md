---
weight: 1
title: Install
---

# Installing Dud

## System Requirements

Dud requires a UNIX-based operating system. Dud does not run on Windows and is
only tested on Linux with a 64-bit x86 CPU. That said, Dud _should_ work on
[Windows Subsystem for
Linux](https://en.wikipedia.org/wiki/Windows_Subsystem_for_Linux) and macOS
operating systems, as well as 32-bit x86 and ARM CPU architectures. If you want
to kick Dud's tires in any of these environments, great! If you encounter
a problem, please [submit a Github
issue](https://github.com/kevin-hanselman/dud/issues/new/choose). Thanks in
advance!

### rclone

Dud uses [rclone](https://rclone.org) to interact with remote storage. Rclone is
required for the `push` and `fetch` commands. Visit https://rclone.org for more
information and installation instructions.


## Installing Dud from a release

Dud releases are [published on
Github](https://github.com/kevin-hanselman/dud/releases). Select a release that
matches your operating system and CPU architecture. Please keep in mind the
notes in the System Requirements section above. **Builds besides Linux x84_64 are
provided for convenience, but they are not explicitly tested.**

First download the release tarball, then extract its contents, and finally copy
the `dud` executable to somewhere in your `$PATH`. The following shell command
accomplishes these steps for a given release asset URL, copying the `dud`
executable to the user's default Go binary path, `~/go/bin`.

    curl -fL 'https://github.com/kevin-hanselman/dud/releases/COPY_URL_FROM_GITHUB.tar.gz' \
        | tar -C ~/go/bin -zxvf - dud


## Building Dud from source

You will need the following software packages:

* git
* Go (version 1.16 or greater)
* GNU Make

First, clone the Dud code repository:

    git clone https://github.com/kevin-hanselman/dud

Now use Make to build the Dud executable and install it in your Go binaries
directory.

    make install

This command will compile Dud, run an automated test suite, and install Dud to
a standard location. It will show you where it installed Dud, and if you have
a standard Go environment, it will likely be installed at `~/go/bin/dud`. If you
haven't already, add this directory to your `$PATH` environment variable so Dud
is accessible from anywhere on your system.
