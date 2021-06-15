## Concepts

### Artifact

An Artifact is a file or directory that is tracked by Dud. Artifacts are usually
stored in the Cache, but it isn't strictly necessary.

### Stage

A Stage is a group of Artifacts, or an operation that consumes and/or produces
a group of Artifacts. Stages are defined by the user in YAML files and should
be tracked with source control. The Stage YAML file format is described in
[`dud stage --help`]({{<ref "cli/dud_stage">}}).

### Index

The Index is the comprehensive group of Stages in a project. It is stored in
a plain text file at `.dud/index`. The Index forms a dependency graph of Stages,
enabling the user to define data pipelines.

### Cache

The Cache is a local directory where Dud stores and versions the contents of
Artifacts. The Cache is content-addressed, which (among other things)
facilitates storing all versions of all Artifacts without conflicts or
duplication.
