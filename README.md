# Duc

Duc is a tool that facilitates storing, versioning, and reproducing large files
alongside source code.

## Design Principles

Duc aims to be simple, fast, and transparent.

### Simple

Duc should Do One Thing Well.

Duc should never get in your way (unless you're about to do something stupid).

Duc should strive to be stateless.

### Fast

Duc should prioritize speed while maintaining sensible assurances of data
integrity.

Duc should isolate time-intensive operations to keep the majority of the UX
as fast as possible.

Duc should scale to datasets in the hundreds of gigabytes and/or millions of
files.

### Transparent

Duc should explain itself early and often.

Duc should maintain its state in a human-readable (and ideally human-editable)
form.
