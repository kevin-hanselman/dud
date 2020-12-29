---
title: dud stage gen
---
## dud stage gen

Generate Stage YAML using the CLI

### Synopsis

Gen generates a Stage YAML file and prints it to STDOUT.

The output of this command can be redirected to a file and modified further as
needed.

```
dud stage gen [flags] [--] [stage_command]...
```

### Examples

```
dud stage gen -o data/ python download_data.py > download.yaml
```

### Options

```
  -d, --dep strings       one or more dependent files or directories
  -h, --help              help for gen
  -o, --out strings       one or more output files or directories
  -w, --work-dir string   working directory for the stage's command
```

### Options inherited from parent commands

```
      --profile   enable profiling
      --trace     enable tracing
```

### SEE ALSO

* [dud stage]({{< relref "dud_stage.md" >}})	 - Commands for interacting with stages and the index

