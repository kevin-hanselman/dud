# A Brief Tour of Dud

The goal of this page is to walk through the core capabilities of Dud. First we'll create a Dud project, then we'll store and version a large file with Dud, and finally we'll create a reproducible data pipeline. If you want to follow along, [install Dud]({{< ref "install/_index.md" >}}) before getting started.

## Create a project

First, let's create a new Dud project:

    $ mkdir ~/cifar && cd ~/cifar

    $ dud init
    Dud project initialized.
    See .dud/config.yaml and .dud/rclone.conf to customize the project.

As Dud tells us, `dud init` creates a `.dud` directory and some config files. Dud is designed to be ready to use out of the box, so don't worry about these files for now.

Next, let's download the CIFAR-10 computer vision dataset.

    $ curl -O https://www.cs.toronto.edu/~kriz/cifar-10-python.tar.gz
      % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                     Dload  Upload   Total   Spent    Left  Speed
    100  162M  100  162M    0     0  13.2M      0  0:00:12  0:00:12 --:--:-- 13.9M

    $ ls
    cifar-10-python.tar.gz

## Store and version large files

Now we have a tarball of the dataset. This file is fairly large, so it's not a good idea to store it in Git. Dud, however, is purpose-built for large files. (In fact, this file is quite small for Dud; it can comfortably handle files in the hundreds of Gigabytes.)

To have Dud store and version this file, we need to create a stage. In Dud, a stage is basically a record of one or more files (or directories) that Dud should track. Stages are written in YAML, and Dud provides a command-line interface to help you with the stage syntax. Here's the command to create the stage file for our dataset tarball:

    $ dud stage gen -o cifar-10-python.tar.gz | tee cifar.yaml
    working-dir: .
    outputs:
      cifar-10-python.tar.gz: {}

`dud stage gen -o cifar-10-python.tar.gz` tells Dud: "Generate the stage YAML to track the file `cifar-10-python.tar.gz`, and print the YAML to standard output." Because `dud stage gen` prints to stdout, in true UNIX fashion we can redirect its output anywhere we like. In this example, I use `tee` to copy the YAML to `cifar.yaml` while keeping stdout intact. This is just so we can quickly see the contents of the file in this walkthrough. I could've just as easily used `> cifar.yaml`.

Now we have a stage file, but we need to register it with Dud. We do that with `dud stage add`:

    $ dud stage add cifar.yaml
`dud stage add` adds a stage to Dud's index. The index is a simple text file (located at `.dud/index`) that tells Dud where all the stages are defined in the project. Don't worry too much about the index for now. It's enough to know that the index makes commands like `dud status` find and load our stage files:

    $ dud status
    cifar.yaml   (stage definition not checksummed)
      cifar-10-python.tar.gz  uncommitted

`dud status` gives us an overview of the Dud project. Here we can see our new stage, `cifar.yaml`, and the file it tracks, `cifar-10-python.tar.gz`. Dud tells us that the tarball is "uncommitted." This means Dud isn't storing this version of the file yet. Let's fix that by committing the stage:

    $ dud commit
    committing stage cifar.yaml
      cifar-10-python.tar.gz  162.60 MiB / 162.60 MiB  100%  ?/s  68ms total

`dud commit` goes through all of our stages (in this case, just `cifar.yaml`) and copies their files/directories to the Dud cache. The cache is a directory that holds all versions of all files and directories owned by Dud. By default, the cache lives at `.dud/cache/`, but it's location is configurable (see `.dud/config.yaml`).

To get a better sense of what `dud commit` did, let's look at the directory structure of the project:

    $ tree -an
    .
    ├── cifar-10-python.tar.gz -> /home/user/cifar/.dud/cache/fe/3d11c475ae0f6fec91f3cf42f9c69e87dc32ec6b44a83f8b22544666e25eea
    ├── cifar.yaml
    └── .dud
        ├── cache
        │   └── fe
        │       └── 3d11c475ae0f6fec91f3cf42f9c69e87dc32ec6b44a83f8b22544666e25eea
        ├── config.yaml
        ├── index
        └── rclone.conf

    3 directories, 6 files

Our tarball has been replaced with a link to a file in Dud's cache. That cached file is our original tarball, but it's named after the checksum of its contents. This is called [content-addressed storage](https://en.wikipedia.org/wiki/Content-addressable_storage), and it allows Dud to track any number of files (and any number of *versions* of files) with reasonable assurances against conflicts or duplication.

But how do we make sure we don't corrupt the cached version of the tarball? What happens if we accidentally modify our dataset?

    $ echo 'accidental overwrite' > cifar-10-python.tar.gz
    /usr/sbin/sh: line 1: cifar-10-python.tar.gz: Permission denied

Dud makes it very difficult to accidentally modify committed files. When Dud commits a file, it makes the link to the cache read-only.

The tarball isn't the only thing that's changed. Let's look at our stage file:

    $ cat cifar.yaml
    checksum: 59f5cc183a8f5784433aaf6b36eea327f53dc03a8953b89b55063a6630c902ea
    working-dir: .
    outputs:
      cifar-10-python.tar.gz:
        checksum: fe3d11c475ae0f6fec91f3cf42f9c69e87dc32ec6b44a83f8b22544666e25eea

Dud recorded the tarball's checksum in the stage file. (It also checksummed the stage file itself; more on that later.) With a copy of this stage file and the Dud cache, we can easily get this specific version of the tarball back. Let's see how that works by deleting the link to the tarball and asking Dud to `checkout` our stage:

    $ rm cifar-10-python.tar.gz

    $ readlink -v cifar-10-python.tar.gz
    readlink: cifar-10-python.tar.gz: No such file or directory

    $ dud checkout
    checking out stage cifar.yaml

    $ readlink -v cifar-10-python.tar.gz
    /home/user/cifar/.dud/cache/fe/3d11c475ae0f6fec91f3cf42f9c69e87dc32ec6b44a83f8b22544666e25eea

`dud checkout` goes through all of our stages, finds their checksummed files in the cache, and makes the appropriate links in our workspace.

Before we move on, let's check `dud status` again:

    $ dud status
    cifar.yaml   (stage definition up-to-date)
      cifar-10-python.tar.gz  up-to-date (link)

Dud tells us that `cifar-10-python.tar.gz` is committed and up-to-date, and it's available in our working directory as a read-only link to the cache.

We also see that `cifar.yaml` has its "stage definition up-to-date." Don't worry about this for this walkthrough. But if you're curious, this is what the top-level checksum in `cifar.yaml` buys us: Dud can detect when we've changed the stage (for example, added another file to track) to better inform its actions.

## Creating a basic data pipeline

Okay, let's extract the tarball already! We do that with `tar`, but it turns out [`tar` is hard to use](https://xkcd.com/1168/). We could write a quick shell script so we don't have to remember the `tar` flags, but we can do one better with Dud.

Now we'll glimpse the full power of Dud stages. As we've seen already, Dud stages own files, but they can also own the *command that generates those files*. On top of that, stages can *depend on other stages*. Let's see this in action by creating a stage to extract the CIFAR-10 tarball:

    $ mkdir cifar-10-batches-py

    $ dud stage gen \
        -d cifar-10-python.tar.gz \
        -o cifar-10-batches-py/ \
        -- tar -xvf cifar-10-python.tar.gz \
        | tee extract_cifar.yaml
    command: tar -xvf cifar-10-python.tar.gz
    working-dir: .
    dependencies:
      cifar-10-python.tar.gz: {}
    outputs:
      cifar-10-batches-py:
        is-dir: true

There's only two lines in this `dud stage gen` command that are new to us. `-d cifar-10-python.tar.gz` declares the tarball as a dependency to this stage. If the tarball changes (for example, if new images are added), Dud will know this stage should be re-run. The `--` tells Dud to stop looking for command-line flags (like `-d` and `-o`) and to treat the rest of the arguments as a shell command. In this case, our command is the `tar` command to extract the contents of the archive.

In the YAML output, notice that Dud can own entire directories, not just files. `dud stage gen` knew to add `is-dir: true` because we created the directory ahead of time.

Let's add the stage and check our status:

    $ dud stage add extract_cifar.yaml

    $ dud status
    cifar.yaml   (stage definition up-to-date)
      cifar-10-python.tar.gz  up-to-date (link)
    extract_cifar.yaml   (stage definition not checksummed)
      cifar-10-batches-py  uncommitted

This looks as expected. Our tarball is committed, but we haven't extracted it yet. We do that using `dud run`:

    $ dud run
    nothing to do for stage cifar.yaml
    running stage extract_cifar.yaml
    cifar-10-batches-py/
    cifar-10-batches-py/data_batch_4
    cifar-10-batches-py/readme.html
    cifar-10-batches-py/test_batch
    cifar-10-batches-py/data_batch_3
    cifar-10-batches-py/batches.meta
    cifar-10-batches-py/data_batch_2
    cifar-10-batches-py/data_batch_5
    cifar-10-batches-py/data_batch_1

As we can see, `dud run` was considering more than just our new stage. It notes that there is "nothing to do for stage cifar.yaml". This makes sense because we didn't define a command in that stage. Then, Dud begins running `extract_cifar.yaml`, and it sends all of `tar`'s output to the terminal.

Now we can check that the `tar` command worked:

    $ ls cifar-10-batches-py/
    batches.meta  data_batch_2  data_batch_4  readme.html
    data_batch_1  data_batch_3  data_batch_5  test_batch

Congrats on [defusing the bomb](https://xkcd.com/1168/)! Now that we know our pipeline worked, let's commit everything:

    $ dud commit
    committing stage cifar.yaml
    cifar-10-python.tar.gz up-to-date; skipping commit
    committing stage extract_cifar.yaml
      cifar-10-batches-py  177.59 MiB / 177.59 MiB  100%  ?/s  30ms total

Notice that Dud detected that the tarball from `cifar.yaml` hasn't changed, so it knew not to waste time committing it again.

Let's take another look at our project status for good measure:

    $ dud status
    cifar.yaml   (stage definition up-to-date)
      cifar-10-python.tar.gz  up-to-date (link)
    extract_cifar.yaml   (stage definition up-to-date)
      cifar-10-batches-py  up-to-date

Looks good! I bet you can guess what happens if we try re-running our pipeline:

    $ dud run
    nothing to do for stage cifar.yaml
    nothing to do for stage extract_cifar.yaml

Because both of our stages are committed and up-to-date, Dud detects that there's no sense in re-extracting the tarball. Excellent!

Let's finish up by visualizing our data pipeline. Dud can output its pipelines in DOT format to be visualized with [Graphviz](https://graphviz.org/). If we have Graphviz installed, we can pipe `dud graph` directly into `dot` to generate an image of the pipeline:

    $ dud graph | dot -Tpng -o pipeline.png

![png](../tour_48_0.png)

Dud shows our two stages as boxes with the files/directories they own as encapsulated ellipses. The arrow between stages shows that the `extract_cifar.yaml` stage depends on the tarball file owned by the `cifar.yaml` stage.

## Summary

In this page we walked through the core capabilities of Dud: storing and versioning large files and directories, and creating reproducible data pipelines. If haven't already, now's a good time to [install Dud]({{< ref "install/_index.md" >}}) and give it a go yourself! If you want to learn more, check out the [CLI Reference]({{< ref "cli/dud">}}) or reach out on [Github Discussions](https://github.com/kevin-hanselman/dud/discussions).