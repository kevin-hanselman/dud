# Dud

Dud is a tool for storing, versioning, and reproducing large files alongside
source code.

![Go report card](https://goreportcard.com/badge/github.com/kevin-hanselman/dud)
![Build status](https://github.com/kevin-hanselman/dud/workflows/build/badge.svg)

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
