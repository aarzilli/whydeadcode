package pkg1

import (
	"fmt"
	"reflect"
)

type Astruct struct {
	N int
}

func (a *Astruct) UnreachableMethod() {
	fmt.Println("UnreachableMethod")
}

func (a *Astruct) ReflectMethodByName(name string) {
	reflect.ValueOf(&Astruct{}).MethodByName(name).Call(nil)
}

func (a *Astruct) M() {
	fmt.Println("M", a.N)
}

