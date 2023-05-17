package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

var callers = map[string][]string{}
var reflectMethods = map[string]bool{}

func main() {
	reflectMethods["reflect.Value.MethodByName"] = true
	reflectMethods["reflect.Value.Method"] = true
	seen := map[string]bool{}
	first := true

	buf := bufio.NewScanner(os.Stdin)
bufScanLoop:
	for buf.Scan() {
		line := buf.Text()
		fields := strings.Split(line, " -> ")
		if len(fields) != 2 {
			continue
		}
		for i := range fields {
			var flags string
			fields[i], flags = splitFlags(fields[i])
			if fields[i] == "" {
				continue bufScanLoop
			}
			if strings.Contains(flags, "<ReflectMethod>") {
				reflectMethods[fields[i]] = true
			}
		}
		callers[fields[1]] = append(callers[fields[1]], fields[0])
		if reflectMethods[fields[1]] && first {
			first = false
			enum([]string{fields[1]}, seen)
		}
	}

	for reflectMethod := range reflectMethods {
		enum([]string{reflectMethod}, seen)
	}
}

func enum(path []string, seen map[string]bool) bool {
	last := path[len(path)-1]
	if last == "type:reflect.Value" || last == "type:*reflect.rtype" || last == "type:*reflect.Value" {
		// these are almost always false positives so we skip them
		return false
	}
	seen[last] = true
	defer func() {
		if !visitOnce(last) {
			delete(seen, last)
		}
	}()
	if len(callers[last]) == 0 {
		if last != "_" {
			return false
		}
		fmt.Println(path[0], "reachable from:")
		for i := 1; i < len(path); i++ {
			fmt.Println("\t", path[i])
		}
		fmt.Println()
		return true
	}
	r := false
	for _, next := range callers[last] {
		if seen[next] {
			continue
		}
		r2 := enum(append(path, next), seen)
		r = r || r2
		if r && visitOnce(last) {
			return r
		}
	}
	return r
}

func splitFlags(sym string) (string, string) {
	for i := len(sym) - 1; i >= 0; i-- {
		if sym[i] == ' ' {
			return sym[:i], sym[i+1:]
		}
		if (sym[i] < 'a' || sym[i] > 'z') && (sym[i] < 'A' || sym[i] > 'Z') && (sym[i] != '<') && (sym[i] != '>') {
			return sym, ""
		}
	}
	return sym, ""
}

func visitOnce(sym string) bool {
	slash := strings.Index(sym, "/")
	if slash < 0 {
		return false
	}
	return strings.Index(sym[:slash], ".") >= 0
}
