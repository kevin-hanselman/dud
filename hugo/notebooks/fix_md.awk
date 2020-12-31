BEGIN {
    inBash = 0
    numEmpty = 0
    rmNextIfEmpty = 0
    inPython = 0
}

{
    # Remove all trailing whitespace and carriage returns.
    sub(/[ \t\r]+$/, "")
}

/%%bash/ || /%%writefile/ {
    next
}

/\w+_files\// {
    sub(/\w+_files\//, "../")
}

/^```\s*python/ {
    inPython = $0
    next
}

/^```$/ {
    if (inBash) {
        inBash = 0
        rmNextIfEmpty = 1
        next
    }
    if (inPython) {
        inPython = 0
    }
}

/^$/ {
    if (rmNextIfEmpty) {
        rmNextIfEmpty = 0
    } else if (++numEmpty == 1) {
        print
    }
    next
}

{
    numEmpty = 0
    rmNextIfEmpty = 0
    if (inPython) {
        # If the first line in the code block starts with a bang, this is
        # actually a shell command. In this case, we remove the code fence
        # posts (and with them the Python syntax highlighting), change the bang
        # to a dollar sign, and indent the entire block (see inBash clause
        # below). If there is no bang, this is actually a Python block, so we
        # add the starting fence back.
        if (/^!/) {
            inBash = 1
            inPython = 0
            sub(/^!/, "$ ", $0)
        } else {
            print inPython
        }
    }
    if (inBash) {
        print "    "$0
    } else {
        print
    }
}
