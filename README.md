# SUMMARY

Displays why deadcode elimination was partially disabled by the Go linker.

The Go linker will disable most deadcode elimination if it finds reachable calls to `reflect.Value.MethodByName` (that don't pass a constant string), `reflect.Value.Method` or `reflect.Value.Methods`. This is done because, using these two methods it is possible to dynamically call any public method in the application.
Whydeadcode uses the call graph produced by the linker to display why `reflect.Value.MethodByName`, `reflect.Value.Method` or `refect.Value.Methods` (or a function that performs an inlined call to one of those methods) are reachable. Use it like this:

```
go build -ldflags=-dumpdep your/package |& whydeadcode
```

Needs Go 1.21 or later.

By default whydeadcode will only print the first path to one of those functions it finds: because of how `-dumpdep` works only one path such path can be identified with certainty. 
The flag `-a` can be specified to print all paths but anything beyond the first result  can be a false positive (i.e. things that look like they will affect deadcode elimination but won't if the first result is taken care of). This output can also have false negatives (i.e. things that will continue to keep deadcode elimination disabled if the first result is taken care of but aren't shown).

# USAGE

```console
$ whydeadcode -h
Usage of ./whydeadcode:
        go build -ldflags=-dumpdep ... |& whydeadcode
  -a    Show all results
  -fail
        Fail on non-empty findings
  -ignore-unrecognized-input
        Ignore unrecognized input
  -s string
        Instead of printing a path to one of the reflect functions, print a path to the specified symbol
```

# TALK

For a more detailed explanation see the [golab 2023 talk](https://youtu.be/EkG177eRcco) or the [slides](https://github.com/aarzilli/talks/blob/master/golab2023_deadcode.pdf) (but see the [talk update](talk_update.md)).

