# Dud

[![Build status](https://github.com/kevin-hanselman/dud/workflows/build/badge.svg)](https://github.com/kevin-hanselman/dud/actions?query=workflow%3Abuild)
[![Go report card](https://goreportcard.com/badge/github.com/kevin-hanselman/dud)](https://goreportcard.com/report/github.com/kevin-hanselman/dud)

Dud is a tool for storing, versioning, and reproducing large files alongside
source code.

With Dud, you can **commit, checkout, fetch, and push large files and
directories** with a simple command line interface. Dud stores recipes (a.k.a.
stages) for retrieving your data in small YAML files. These recipes can be
stored in source control to **link your data to your code**. On top of that, the
recipes can **run the commands to generate the data**, sort of like
[Make](https://www.gnu.org/software/make/). Recipes can be chained together to
**create data pipelines**.

Dud is heavily inspired by [DVC](https://dvc.org/). If DVC is [Django][1], Dud
aims to be [Flask][1]. Dud is [much
faster](https://kevin-hanselman.github.io/dud/benchmarks) and has
a [significantly smaller and simpler
CLI](https://kevin-hanselman.github.io/dud/cli/dud).

[1]: https://hackr.io/blog/flask-vs-django

Check out [the whirlwind
tour](https://kevin-hanselman.github.io/dud/getting_started/tour/) to get a feel
for using Dud.

Oh, and Dud is pronounced "duhd". Don't ask it where your car is; it doesn't
know.


## Design Principles

Dud aims to be simple, fast, and transparent.

### Simple

Dud should do one thing well and be a good UNIX citizen.

Dud should never get in your way (unless you're about to do something stupid).

Dud should strive to be stateless.

### Fast

Dud should prioritize speed while maintaining sensible assurances of data
integrity.

Dud should isolate time-intensive operations to keep the majority of the UX
as fast as possible.

Dud should scale to datasets in the hundreds of gigabytes and/or millions of
files.

### Transparent

Dud should explain itself early and often.

Dud should maintain its state in a human-readable (and ideally human-editable)
form.


## Terminology

### Artifact

An Artifact is a file or directory that is tracked by Dud. Artifacts are usually
stored in the Cache, but it isn't strictly necessary.

### Stage

A Stage is a group of Artifacts, or an operation that consumes and/or produces
a group of Artifacts. Stages are defined by the user in YAML files and should be
tracked with source control.

### Index

The Index is the comprehensive group of Stages in a project. The Index forms
a dependency graph of Stages, enabling the user to define data pipelines.

### Cache

The Cache is a local directory where Dud stores and versions the contents of
Artifacts. The Cache is content-addressed, which (among other things)
facilitates storing all versions of all Artifacts without conflicts or
duplication.
