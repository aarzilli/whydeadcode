package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

var callers = map[string][]string{}
var reflectMethods = map[string]bool{}

func main() {
	ignoreUnrecognizedInput := flag.Bool("ignore-unrecognized-input", false, "Ignore unrecognized input")
	fail := flag.Bool("fail", false, "Fail on non-empty findings")
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintf(out, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(out, "\tgo build -ldflags=-dumpdep ... |& whydeadcode\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	seen := map[string]bool{}
	first := true
	found := false
	ul := []string{}

	buf := bufio.NewScanner(os.Stdin)
bufScanLoop:
	for buf.Scan() {
		line := buf.Text()
		fields := strings.Split(line, " -> ")
		if len(fields) != 2 && !strings.Contains(line, "go:string") {
			ul = append(ul, line)
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
			found = enum([]string{fields[1]}, seen) || found
		}
	}

	for reflectMethod := range reflectMethods {
		found = enum([]string{reflectMethod}, seen) || found
	}

	if !*ignoreUnrecognizedInput && len(ul) > 1 {
		fmt.Fprintf(os.Stderr, "Unrecognized input:\n\n")
		io.WriteString(os.Stderr, strings.Join(ul, "\n"))
		io.WriteString(os.Stderr, "\n")
		os.Exit(1)
	}

	if *fail && found {
		os.Exit(1)
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
