package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

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

	paths, ul := Whydeadcode(os.Stdin)

	for _, path := range paths {
		path.Print()
	}

	if !*ignoreUnrecognizedInput && len(ul) > 1 {
		fmt.Fprintf(os.Stderr, "Unrecognized input:\n\n")
		io.WriteString(os.Stderr, strings.Join(ul, "\n"))
		io.WriteString(os.Stderr, "\n")
		os.Exit(1)
	}

	if *fail && len(paths) > 0 {
		os.Exit(1)
	}
}

type Path []string

// Whydeadcode parses the output of the linker and returns a list of problematic paths.
func Whydeadcode(linkerOut io.Reader) (pathsToReflectMethods []Path, unrecognizedLines []string) {
	seen := map[string]bool{}
	first := true
	reflectMethods := map[string]bool{}
	callers := map[string][]string{}

	buf := bufio.NewScanner(linkerOut)
bufScanLoop:
	for buf.Scan() {
		line := buf.Text()
		fields := strings.Split(line, " -> ")
		if len(fields) != 2 && !strings.Contains(line, "go:string") {
			unrecognizedLines = append(unrecognizedLines, line)
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
			enum(&pathsToReflectMethods, []string{fields[1]}, seen, first, callers)
			first = false
		}
	}

	for reflectMethod := range reflectMethods {
		enum(&pathsToReflectMethods, []string{reflectMethod}, seen, first, callers)
		first = false
	}

	return pathsToReflectMethods, unrecognizedLines
}

func enum(pathsToReflectMethods *[]Path, path Path, seen map[string]bool, first bool, callers map[string][]string) bool {
	last := path[len(path)-1]
	if !first {
		if last == "type:reflect.Value" || last == "type:*reflect.rtype" || last == "type:*reflect.Value" {
			// these are almost always false positives so we skip them
			return false
		}
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
		*pathsToReflectMethods = append(*pathsToReflectMethods, slices.Clone(path))
		return true
	}
	r := false
	for _, next := range callers[last] {
		if seen[next] {
			continue
		}
		r2 := enum(pathsToReflectMethods, append(path, next), seen, first, callers)
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

func (path Path) Print() {
	fmt.Println(path[0], "reachable from:")
	for i := 1; i < len(path); i++ {
		fmt.Println("\t", path[i])
	}
	fmt.Println()
}
