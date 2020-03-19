# Data Under Control

Data Under Control is a tool for storing, versioning, and reproducing big data
files right alongside the source code that creates and/or uses it. It is heavily
inspired by Data Version Control and Git.


## Design Goals

### Seamlessly scale to Terabytes of data

The real utility of DUC lies in its ability to scale effortlessly to datasets
that measure in the Terabytes. DUC operations should remain quick (as can be
expected) and informative, no matter the size of the data.

### What Would Git Do?

DUC is designed to be a close companion to Source Control Management (SCM)
tools -- specifically Git. When possible, DUC's operation should mirror Git.
"If this was a source code file, what would Git do with it?"

### Prefer transparency over abstraction

DUC should explain itself early and often, and shouldn't shirk from sharing
details of its inner-workings with the user. Most of the mechanisms DUC uses to
do its job (such as files, links, and checksums) are intimately familiar to
those using Git, and inspecting the inner-workings should be encouraged.

### Use sensible defaults

DUC should excel at its job (i.e. meet all of the goals above) with only trivial
configuration.
