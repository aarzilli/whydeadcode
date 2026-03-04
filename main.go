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
	all := flag.Bool("a", false, "Show all results")
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
		if !*all {
			break
		}
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

type Path []callerEdge

type callerEdge struct {
	caller string
	cause  string
}

// Whydeadcode parses the output of the linker and returns a list of problematic paths.
func Whydeadcode(linkerOut io.Reader) (pathsToReflectMethods []Path, unrecognizedLines []string) {
	seen := map[string]bool{}
	first := true
	reflectMethods := map[string]bool{}
	callers := map[string][]callerEdge{}

	buf := bufio.NewScanner(linkerOut)
bufScanLoop:
	for buf.Scan() {
		line := buf.Text()
		fields := strings.Split(line, " -> ")
		if len(fields) != 2 && !strings.Contains(line, "go:string") {
			unrecognizedLines = append(unrecognizedLines, line)
			continue
		}
		var cause string
		const causeSep = " !! "
		if strings.Contains(fields[1], causeSep) {
			v := strings.Split(fields[1], causeSep)
			if len(v) != 2 && !strings.Contains(line, "go:string") {
				unrecognizedLines = append(unrecognizedLines, line)
				continue
			}
			fields[1] = v[0]
			cause = v[1]
			const causePfx = "cause: "
			if !strings.HasPrefix(cause, causePfx) {
				unrecognizedLines = append(unrecognizedLines, line)
				continue
			}
			cause = cause[len(causePfx):]
		}
		_ = cause
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
		callers[fields[1]] = append(callers[fields[1]], callerEdge{fields[0], cause})
		if reflectMethods[fields[1]] && first {
			enum(&pathsToReflectMethods, []callerEdge{callerEdge{fields[1], ""}}, seen, first, callers)
			first = false
		}
	}

	for reflectMethod := range reflectMethods {
		enum(&pathsToReflectMethods, []callerEdge{callerEdge{reflectMethod, ""}}, seen, first, callers)
		first = false
	}

	return pathsToReflectMethods, unrecognizedLines
}

func enum(pathsToReflectMethods *[]Path, path Path, seen map[string]bool, first bool, callers map[string][]callerEdge) bool {
	last := path[len(path)-1]
	if !first {
		if last.caller == "type:reflect.Value" || last.caller == "type:*reflect.rtype" || last.caller == "type:*reflect.Value" {
			// these are almost always false positives so we skip them
			return false
		}
	}
	seen[last.caller] = true
	defer func() {
		if !visitOnce(last.caller) {
			delete(seen, last.caller)
		}
	}()
	if len(callers[last.caller]) == 0 {
		if last.caller != "_" {
			return false
		}
		*pathsToReflectMethods = append(*pathsToReflectMethods, slices.Clone(path))
		return true
	}
	r := false
	for _, next := range callers[last.caller] {
		if seen[next.caller] {
			continue
		}
		r2 := enum(pathsToReflectMethods, append(path, next), seen, first, callers)
		r = r || r2
		if r && visitOnce(last.caller) {
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
	fmt.Printf("%s reachable from", path[0].caller)
	if len(path) > 0 && path[1].cause != "" {
		fmt.Printf(" (cause: %s)", path[1].cause)
	}
	fmt.Printf(":\n")
	for i := 1; i < len(path); i++ {
		fmt.Printf("\t%s", path[i].caller)
		if i+1 < len(path) && path[i+1].cause != "" {
			fmt.Printf(" (cause: %s)", path[i+1].cause)
		}
		fmt.Printf("\n")
	}
	fmt.Println()
}

func (edge callerEdge) String() string {
	if edge.cause == "" {
		return edge.caller
	}
	return fmt.Sprintf("%s cause: %s", edge.caller, edge.cause)
}
