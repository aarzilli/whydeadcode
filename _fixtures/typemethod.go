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
	mmeth, _ := reflect.TypeOf(&Bstruct{}).MethodByName(name)
	mmeth.Func.Call(nil)
}

func main() {
	F(os.Args[1])
}
