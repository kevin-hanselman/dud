BEGIN {
    inShell = 0
    numEmpty = 0
    rmNextIfEmpty = 0
    inPython = 0
}

{
    # Remove all trailing whitespace and carriage returns.
    sub(/[ \t\r]+$/, "")
    # Remove all ANSI escape sequences (e.g. terminal colors and cursor movements).
    gsub(/\033\[[0-9;]*[a-zA-Z]/, "")
}

/\w+_files\// {
    gsub(/\w+_files\//, "../")
}

/^```\s*python/ {
    inPython = $0
    next
}

/^```$/ {
    if (inShell) {
        inShell = 0
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
        # to a dollar sign, and indent the entire block (see inShell clause
        # below). If there is no bang, this is actually a Python block, so we
        # add the starting fence back.
        if (/^!/) {
            inShell = 1
            sub(/^!/, "$ ", $0)
        } else {
            print inPython
        }
        # Stop processing after the first line.
        inPython = 0
    }
}

/%%bash/ { next }

/%%writefile/ {
    # Print the file name as a Python/shell comment at the top of the block.
    print "# " $2
    next
}

{
    if (inShell) {
        print "    "$0
    } else {
        print
    }
}
