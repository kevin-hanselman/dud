# Constributor's Guide

Please review and abide by the [Code of Conduct](CODE_OF_CONDUCT.md).


## TLDR

`make docker-integration-test` should complete without error. Read
the Makefile (or just try running it!) to see what goes into that command.

If you're making user-facing changes, you should probably update the
website/documentation. The `docs` Make rule will generate the project website
which is stored in the `docs` directory in the main branch. You should run this
rule (preferably as `make docker-docs`, see below) and commit everything in the
`hugo` and `docs` directories as part of your work.


## Development Environment

To build the code and run tests, you will need Git, GNU Make, and Docker.

The Docker image defined in `integration/Dockerfile` is the official development
environment. You can start an interactive container from this image by running
`make docker`. Additionally, every rule in the Makefile can be run in the Docker
image by prefixing the rule with `docker-`. Containers started in either of
these manners will have the host's source repo mounted at `/dud`.


## End-to-end tests

A suite of end-to-end tests lives at `integration/tests`. These tests are simply
shell scripts (always named `run.sh`) that are meant to act as if a user was
typing commands into their shell. These tests are run by
`integration/run_tests.py`. This Python script creates isolated Dud projects for
each test, runs the shell scripts, and asserts that no errors occurred during
execution. The Python script also can diff each test's project file tree to
ensure things look right (see `expected_fs.txt` files in the tests, and
`run_tests.py` for how this file is generated). Tests can be grouped in
sequences using subdirectories (e.g. see
`integration/tests/basic_file_operations`), in which case the sub-tests share
a Dud project and are executed in lexicographic order.
