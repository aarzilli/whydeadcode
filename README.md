Displays why deadcode elimination was not performed by the Go linker.

The Go linker will disable most deadcode elimination if it finds reachable calls to `reflect.Value.MethodByName` or `reflect.Value.Method`. This is done because, using these two methods it is possible to dynamically call any public method in the application. 
Whydeadcode uses the call graph produced by the linker to display why `reflect.Value.MethodByName` or `reflect.Value.Method` are reachable. Use it like this:

```
	go build -ldflags=-dumpdep your/package |& whydeadcode
```

Needs Go 1.21 or later.

Because of how `-dumpdep` works only the first result output by whydeadcode is real. Because of how `-dumpdep` works anything beyond the first result can be a false positive (i.e. things that look like they will affect deadcode elimination but won't if the first result is taken care of) and it can also have false negatives (i.e. things that will continue to keep deadcode elimination disabled if the first result is taken care of).
