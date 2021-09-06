# Dud

[![Build status](https://github.com/kevin-hanselman/dud/workflows/build/badge.svg)](https://github.com/kevin-hanselman/dud/actions?query=workflow%3Abuild)
[![Go report card](https://goreportcard.com/badge/github.com/kevin-hanselman/dud)](https://goreportcard.com/report/github.com/kevin-hanselman/dud)

[Website](https://kevin-hanselman.github.io/dud/)
| [Install](https://kevin-hanselman.github.io/dud/install)
| [Getting Started](https://kevin-hanselman.github.io/dud/getting_started)
| [Source Code](https://github.com/kevin-hanselman/dud)

Dud is a lightweight tool for versioning data alongside source code and building
data pipelines. In practice, Dud extends many of the benefits of source
control to large binary data.

With Dud, you can **commit, checkout, fetch, and push large files and
directories** with a simple command line interface. Dud stores recipes (a.k.a.
stages) for retrieving your data in small YAML files. These stages can be
stored in source control to **link your data to your code**. On top of that,
stages can **run the commands to generate the data**, sort of like
[Make](https://www.gnu.org/software/make/). Stages can be chained together to
**create data pipelines**. See the [Getting
Started](https://kevin-hanselman.github.io/dud/getting_started) guide for
a hands-on overview.

Dud is pronounced "duhd", not "dood". Dud is not an acronym.


## Motivation

Dud is heavily inspired by [DVC](https://dvc.org/). DVC addresses the need for
data versioning and reproducibility, but its implementation is not without
problems. My criticisms of DVC boil down to two things: speed and simplicity. By
speed, I mean throughput and responsiveness. By simplicity, I mean doing
less--both in project scope and amount of abstraction.

In terms of speed, Dud is [generally much
faster](https://kevin-hanselman.github.io/dud/benchmarks) than DVC. In terms of
simplicity, Dud has a [smaller, more focused
scope](https://kevin-hanselman.github.io/dud/cli/dud), and it is [distributed as
a single executable](https://github.com/kevin-hanselman/dud/releases).

To summarize with an analogy: Dud is to DVC what [Flask][1] is to [Django][1].
Both Dud and DVC have their strengths. If you want a "batteries included" suite
of tools for managing machine learning projects, DVC may be a good fit for you.
If data management is your main area of need and you want something lightweight
and fast, Dud may be a better fit.

[1]: https://hackr.io/blog/flask-vs-django

To get down to brass tacks, read on.

### Concrete differences with DVC

#### Dud does not manage experiments and/or metrics.

Dud is solely focused on versioning and reproducing data alongside source code.
DVC's scope has grown to encompass a large portion of a traditional machine
learning workflow. While an integrated suite of tools has its benefits, if UNIX
is any guide, the composition of smaller, more focused tools generally yield
more productivity than their monolithic counterparts. For example, there's no
reason you couldn't use [MLflow](https://mlflow.org/) or
[Aim](https://aimstack.io/) alongside Dud to track your experiments. Dud does
not prescribe any solution for experiment tracking, and it doesn't try to enter
the new, yet already crowded, marketplace for such tools.

Secondly, versioning data alongside source code is an incredibly useful concept
in its own right. Domains beyond machine learning and data science (e.g. game
development and digital design) may greatly benefit from this approach to data
management without being burdened by extra baggage carried by a specific domain.


#### Dud commits must always be explicitly invoked; they are never side effects.

For both Dud and DVC, committing data to the cache is one of the most expensive
operations that each tool undertakes (in terms of both run-time and I/O).
Because of this, Dud puts the user in absolute control of when to commit data.
In Dud, commits only happen in when you run `dud commit`.

In contrast, DVC often commits automatically on your behalf as a side effect of
other commands (for example, during `dvc add` and `dvc repro`). While DVC is
trying to be helpful, these implicit commits are often accidental commits.
For example, if you're rapidly iterating on a pipeline, you're likely running
`dvc repro` or `dvc run` repeatedly as you develop. However, DVC will
automatically commit the results each time you run `dvc repro` or `dvc
run`--even if you are just debugging something or tweaking your code. Such
accidental commits have a high cost; they turn "rapid development" into
"development", and they bloat your cache. (You can disable DVC's implicit
commits using the `--no-commit` flag, but you have to remember to type it each
time, and DVC does not support enabling this flag by default, e.g. via
configuration file.)


#### Dud checks out files as symbolic links by default.

When Dud checks out cached files into the workspace, it uses symbolic links
(a.k.a. symlinks) by default. Symlinks have a number of benefits that make them
an excellent choice for checkouts. First, symlinks require very little I/O to
create, so `dud checkout` usually completes almost instantaneously. Second,
symlinks transparently redirect to the cached files themselves, so data isn't
duplicated between the workspace and the cache, and your storage space is used
efficiently. Last but not least, symlinks make it trivial to check if a file is
up-to-date (by checking the link target), so `dud status` can also be extremely
fast.

By default, DVC checks out files as hard copies. (Technically, DVC tries to use
[reflinks][reflink] before copies, but very few filesystems support reflinks, so
copies are far more likely to be the default.) With hard copies, efficiencies
listed above are not possible, so checkouts and status checks are inefficient by
default. To its credit, DVC's cache can be configured to use symlinks, but
arguably DVC's default cache configuration is not sensible for projects of any
significant size.

[reflink]: https://en.wikipedia.org/wiki/Data_deduplication#reflink


#### Running a Dud pipeline never implicitly alters a stage's artifacts.

When you run a pipeline in DVC, DVC will remove all pipeline outputs before
running the pipeline's command(s). While this can help ensure reproducible
pipelines, it is another implicit behavior the user must consider, and it
prevents the user from deciding when stage outputs can safely be reused.

If you don't want DVC to automatically remove outputs for you, you need to
explicitly tell it each output you'd like to persist. However, by telling DVC to
persist an output, DVC may perform a new and different automatic behavior. If
you're using symbolic links (or hard links) for checkouts (which is generally
a good idea; see above), DVC will "unprotect" all output links by replacing them
with hard copies from the cache. Not only is this behavior surprising, it's also
very costly in both runtime and storage.

The result of these two behaviors in DVC means that, in a sensible
configuration, stages simply cannot reuse outputs efficiently; the user has
little choice but to accept DVC's limitations.

When you run a pipeline Dud, Dud doesn't do any implicit modification of
existing files. Dud defers all modification of workspace files to the user. If
you want a specific behavior, you should code it into your stage's command. For
example, if you want to clear all outputs of a stage prior to it running, you
can delete any outputs at the beginning of your command's script. If you want
to reuse outputs, you can check for preexisting outputs in your script and
choose not to recreate them. Dud's minimalist approach results in a stage's
command entirely owning it's own reproducibility; the responsibility is
not awkwardly shared between the stage and the tool.


#### Dud delegates remote cache management to Rclone.

[Rclone](https://rclone.org) is a very popular command-line tool which describes
itself as "The Swiss army knife of cloud storage." At the time of writing,
Rclone has more than 28,000 stars on Github. Rclone supports just about any
cloud storage provider you've possibly heard of. (S3, GCS, Dropbox, Backblaze,
to name a few.) This is all to say: Rclone is a top-tier choice for moving data
around the internet.

Dud internally calls Rclone for all of its remote cache functionality, such as
`dud fetch` and `dud push`. But Dud doesn't hide the Rclone abstraction
entirely. Dud exposes its Rclone configuration file, and it's expected and
encouraged that users will use Rclone directly to configure remote storage or
interact with their remote data. By using Rclone, Dud's remote cache interface
immediately gains the benefit of years of open-source development and a rich,
well-documented CLI. This is an example of how Dud embraces the UNIX philosophy
and the composition of single-focus tools, as stated above.

In contrast, DVC stiches together various Python packages to support a modest
assortment of cloud storage options. At the time of writing, DVC 2.6 supports
eleven cloud storage providers, and Rclone 1.56 supports more than fifty. But
the amount of cloud storage options isn't the critical disadvantage of DVC's
approach. (Both Dud and DVC support the biggest players, such as S3 and GCS.)
DVC's critical disadvantage is that they must develop and maintain most of their
remote data management stack themselves. If Rclone is any indication, cloud data
transfer is a very hard problem, and DVC has their work cut out for them.

In summary, Dud leverages the deep knowledge and effort of the Rclone developers
to provide a robust and familiar remote cache experience. DVC plots their own
course, and in doing so incurs a steep development cost.


#### Dud does not use analytics. (And it never will.)

By default, DVC enables [embedded
analytics](https://dvc.org/doc/user-guide/analytics#anonymized-usage-analytics).
I strongly disagree with this practice, especially in free and open-source
software. I will never embed analytics in Dud.


## Contributing

See
[CONTRIBUTING.md](https://github.com/kevin-hanselman/dud/blob/main/CONTRIBUTING.md).


## License

BSD-3-Clause. See
[LICENSE](https://github.com/kevin-hanselman/dud/blob/main/LICENSE).

