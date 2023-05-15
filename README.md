Displays why deadcode elimination was not performed by the Go linker.

The Go linker will disable most deadcode elimination if it finds reachable calls to `reflect.Value.MethodByName` or `reflect.Value.Method`. This is done because, using these two methods it is possible to dynamically call any public method in the application. 
Whydeadcode uses the call graph produced by the linker to display why `reflect.Value.MethodByName` or `reflect.Value.Method` are reachable. Use it like this:

```
	go build -ldflags=-c your/package > call-graph.txt 2>&1
	whydeadcode < call-graph.txt
```
