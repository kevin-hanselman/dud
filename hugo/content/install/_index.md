---
weight: 1
title: Installation
---

# Installing Dud

## System Requirements

Dud requires a UNIX-based operating system. Dud does not run on Windows and is
only tested against Linux. That said, Dud _should_ work on [Windows Subsystem
for Linux](https://en.wikipedia.org/wiki/Windows_Subsystem_for_Linux) and macOS.
If you want to kick the tires on one of these environments, please [submit any
issues](https://github.com/kevin-hanselman/dud/issues/new/choose) you may
encounter. Thanks in advance!

You will also need the following software packages:

* git
* Go (version 1.14 or greater)
* GNU Make

## Build and install Dud

The only installation method at this time is to build Dud from source. Luckily,
this isn't as painful as it may sound.

First, clone the Dud code repository:

    git clone https://github.com/kevin-hanselman/dud

Now use Make to build the Dud executable and install it in your Go binaries
directory.

    make install

This command will compile Dud, run an automated test suite, and install Dud to
a standard location. It will show you where it installed Dud, and if you have
a standard Go environment, it will likely be installed at `~/go/bin/dud`. If
you haven't already, add this directory to your `$PATH` environment variable so
Dud is accessible from anywhere on your system. For example, assuming the
installation location `~/go/bin/dud` and a Bash shell:

    export PATH=$PATH:~/go/bin/

You can check that Dud is installed and in your `$PATH` by trying to run Dud:

    dud

If the command worked, you'll see Dud print its help text and exit.
