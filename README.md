# Duc

Duc is a tool that facilitates storing, versioning, and reproducing large files
alongside the code that uses them.

## Design Principles

Duc aims to be simple, fast, and transparent.

### Simple

Duc should Do One Thing Well.

Duc should require only trivial configuration.

### Fast

Duc should prioritize speed.

Duc should isolate time-intensive operations to keep the rest of the UI fast.
Currently, Duc's UI is designed such that only `duc commit` is time-intensive.

Duc should scale to datasets in the hundreds of gigabytes, ideally low
terabytes.

### Transparent

Duc should explain itself early and often. Inspecting Duc's inner-workings
should be easy and encouraged.
