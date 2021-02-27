# A Tour of Dud

Create a project.

    $ mkdir ~/cifar && cd ~/cifar

    $ dud init
    Dud project initialized.
    See .dud/config.yaml and .dud/rclone.conf to customize the project.

Add some data and tell Dud to track it.

    $ curl -sSO https://www.cs.toronto.edu/~kriz/cifar-10-python.tar.gz

    $ dud stage gen -o cifar-10-python.tar.gz | tee cifar.yaml
    working-dir: .
    outputs:
      cifar-10-python.tar.gz: {}

    $ dud stage add cifar.yaml

    $ dud status
    cifar.yaml   (stage definition not checksummed)
      cifar-10-python.tar.gz  uncommitted

Commit the data to the Dud cache.

    $ dud commit
    committing stage cifar.yaml
    ⠋ committing cifar-10-python.tar.gz (163 MB, 2470.318 MB/s)

    $ cat cifar.yaml
    checksum: 59f5cc183a8f5784433aaf6b36eea327f53dc03a8953b89b55063a6630c902ea
    working-dir: .
    outputs:
      cifar-10-python.tar.gz:
        checksum: fe3d11c475ae0f6fec91f3cf42f9c69e87dc32ec6b44a83f8b22544666e25eea

    $ dud status
    cifar.yaml   (stage definition up-to-date)
      cifar-10-python.tar.gz  up-to-date (link)

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

Add a stage to extract the data.

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

    $ dud stage add extract_cifar.yaml
Run our pipeline and commit the results.

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

    $ dud status
    cifar.yaml   (stage definition up-to-date)
      cifar-10-python.tar.gz  up-to-date (link)
    extract_cifar.yaml   (stage definition not checksummed)
      cifar-10-batches-py  uncommitted

    $ dud commit
    committing stage cifar.yaml
    cifar-10-python.tar.gz up-to-date; skipping commit
    committing stage extract_cifar.yaml
    ⠋ committing cifar-10-batches-py (178 MB, 5116.031 MB/s)

Notice that Dud detected the tarball from cifar.yaml was up-to-date and did not copy any data.

    $ dud status
    cifar.yaml   (stage definition up-to-date)
      cifar-10-python.tar.gz  up-to-date (link)
    extract_cifar.yaml   (stage definition up-to-date)
      cifar-10-batches-py  up-to-date

Visualize the pipeline using Graphviz.

    $ dud graph | dot -Tpng -o pipeline.png

![png](../tour_26_0.png)

