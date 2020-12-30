BEGIN {
    inBash = 0
    firstBash = 0
    numEmpty = 0
    rmNextIfEmpty = 0
}

/%%bash/, /%%writefile/ {
    next
}

/\w+_files\// {
    sub(/\w+_files\//, "../")
}

/^```\s*bash/ {
    firstBash = 1
    inBash = 1
    next
}

/^```\s*$/ {
    if (inBash) {
        inBash = 0
        rmNextIfEmpty = 1
    }
    next
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
    if (firstBash) {
        firstBash = 0
        print "    $ " $0
    } else if (inBash) {
        print "    " $0
    } else {
        print
    }
}
