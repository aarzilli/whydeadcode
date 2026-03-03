package main

import (
	"os"

	"github.com/aarzilli/whydeadcode/_fixtures/pkg1"
)

var f = (*pkg1.Astruct).ReflectMethodByName

func main() {
	f(&pkg1.Astruct{ 10 }, os.Args[1])
}
