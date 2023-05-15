package main

import (
	"os"
	"bufio"
	"strings"
	"fmt"
	"debug/gosym"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

var callers = map[string][]string{}

func main() {
	buf := bufio.NewScanner(os.Stdin)
	for buf.Scan() {
		line := buf.Text()
		fields := strings.Split(line, " calls ")
		if len(fields) != 2 {
			continue
		}
		fields[0] = removeClosures(fields[0])
		fields[1] = removeClosures(fields[1])
		callers[fields[1]] = append(callers[fields[1]], fields[0])
	}
	
	seen := map[string]bool{}
	enum([]string{ "reflect.Value.MethodByName" }, seen)
	enum([]string{ "reflect.Value.Method" }, seen)
}

func enum(path []string, seen map[string]bool) {
	last := path[len(path)-1]
	seen[last] = true
	if len(callers[last]) == 0 {
		if !isPublicMethod(last) {
			fmt.Println(path[0], "called by:")
			for i := 1; i < len(path); i++ {
				fmt.Println("\t", path[i])
			}
			fmt.Println()
		}
	}
	for _, next := range callers[last] {
		if seen[next] {
			continue
		}
		enum(append(path, next), seen)
	}
}

func removeClosures(p string) string {
	dot := strings.LastIndex(p, ".")
	if dot < 0 {
		return p
	}
	if !strings.HasPrefix(p[dot+1:], "func") {
		return p
	}
	for i := dot+5; i < len(p); i++ {
		if p[i] < '0' || p[i] > '9' {
			return p
		}
	}
	return p[:dot]
}

func isPublicMethod(p string) bool {
	sym := &gosym.Sym{ Name: p }
	if sym.ReceiverName() == "" {
		return false
	}
	bn := sym.BaseName()
	if bn == "" {
		return false
	}
	return bn[0] >= 'A' && bn[0] <= 'Z'
}
