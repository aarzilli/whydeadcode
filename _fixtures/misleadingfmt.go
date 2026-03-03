package main

import (
	"fmt"
	"reflect"
)

type Astruct struct {
	n int
}

func (a *Astruct) Methods() int {
	return a.n + 1
}

func UseMethods() reflect.Value {
	return reflect.ValueOf(&Astruct{10}).MethodByName("Methods")
}

func main() {
	v := UseMethods()
	fmt.Println(v)
}
