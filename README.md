# Dud

Dud is a tool that facilitates storing, versioning, and reproducing large files
alongside source code.

## Design Principles

Dud aims to be simple, fast, and transparent.

### Simple

Dud should Do One Thing Well.

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
