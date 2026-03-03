package main

import (
	"fmt"
	"os"
	"reflect"
)

type Bstruct struct {
	n float64
}

func (b *Bstruct) M() {
	fmt.Println("hello world")
}

func F(name string) {
	reflect.ValueOf(&Bstruct{ 2.0 }).MethodByName(name).Call(nil)
}

func main() {
	F(os.Args[1])
}
