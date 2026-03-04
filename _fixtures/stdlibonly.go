package main

import (
	"errors"
	"fmt"
	"os"
)

// stdlibonly exercises code paths that pull in stdlib <ReflectMethod> functions
// (e.g. fmt.Errorf -> fmt.(*pp).doPrintf -> reflect.TypeOf -> *reflect.rtype,
// and text/template via cobra-style usage templates) without any user-code
// reflection. whydeadcode should produce no findings for this binary.
func main() {
	if len(os.Args) > 1 {
		err := errors.New(os.Args[1])
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
