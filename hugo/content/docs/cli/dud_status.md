---
title: dud status
---
## dud status

Print the state of one or more stages

### Synopsis

Status prints the state of one or more stages.

For each stage file passed in, status will print the current state of the
stage.  If no stage files are passed in, status will act on all stages in the
index. By default, status will act recursively on all upstream stages (i.e.
dependencies).

```
dud status [flags] [stage_file]...
```

### Options

```
  -h, --help   help for status
```

### Options inherited from parent commands

```
      --profile   enable profiling
      --trace     enable tracing
```

### SEE ALSO

* [dud]({{< relref "dud.md" >}})	 - 

