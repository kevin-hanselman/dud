---
name: Bug report
about: Create a report to help us improve
title: ''
assignees: ''

---

**Please acknowledge the following**
- [ ] I have read about [Minimal Bug Reports][bugs] and what follows is my good faith attempt at creating one.

**Describe the bug**
A clear and concise description of what the bug is.

**System information**

Output of `dud version`:
```
dud version
# copy output here
```

Output of `uname -srmo`:
```
uname -srmo
# copy output here
```

If you're building from source, please provide your Go version.
```
go version
# copy output here
```

**Steps to Reproduce**
Steps to reproduce the behavior. Ideally this a copy-paste-able shell script (or
set of small scripts) that reproduces the problem.
```
# for example:
dud init
echo -e 'outputs:\n  foo.txt' > stage.yaml
dud stage add stage.yaml
# etc.
```

**Expected behavior**
A clear and concise description of what you expected to happen.


[bugs]: https://matthewrocklin.com/blog/work/2018/02/28/minimal-bug-reports
