# Contributor's Guide

Please review and abide by the [Code of Conduct](CODE_OF_CONDUCT.md).


## Design Goals

When contributing to Dud, please keep in mind the following goals of the
project:

Dud should be simple, fast, and transparent.

### Simple

Dud should do one thing well and be a good UNIX citizen.

Dud should never get in your way (unless you're about to do something stupid).

Dud should be less magical, not more.

### Fast

Dud should prioritize speed while maintaining sensible assurances of data
integrity.

Dud should isolate time-intensive operations to keep the majority of the UX
as fast as possible.

Dud should scale to datasets in the hundreds of gigabytes and/or hundreds of
thousands of files.

### Transparent

Dud should explain itself early and often.

Dud should maintain its state in a human-readable (and ideally human-editable)
form.


## Testing TLDR

`make docker-integration-test` should complete without error. Read
the Makefile (or just try running it!) to see what goes into that command.


## Development environment

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


## Documentation and Website

This project uses [Hugo](https://gohugo.io/) to turn Markdown files under
`hugo/content/` into a static website served on Github Pages. The CLI docs are
generated automatically by the Cobra Go library (see `make cli-docs`), but all
other Markdown files require some level of manual intervention.

### Benchmarks

TODO: Expand this documentation.

See `make hugo/content/benchmarks/_index.md` for how the benchmarks page is
built.

### Jupyter Notebooks

This project uses Jupyter Notebooks to build executable and reproducible
documentation. To start the Jupyter Notebook server, run `make
docker-serve-jupyter`.

See `make hugo/notebooks/%.md` for how Jupyter Notebooks get converted to
Markdown, and then see `make hugo/content/%.md` for how that Markdown is cleaned
and added to the `hugo/content/` directory.

`make hugo/notebooks/%.md` doesn't execute the notebooks automatically; you will
have to execute them manually. The reason for this is mostly because notebook
execution can be an expensive operation (in time and network bandwidth), but
also because there are some inconsistencies in the output of `nbconvert
--execute` versus manual execution.
