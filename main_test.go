package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestWhydeadcode(t *testing.T) {
	for _, c := range []struct {
		fixtureName string
		tgt         []string
	}{
		{"valuemethod", []string{"main.F", "main.main"}},
		{"typemethod", []string{"main.F", "main.main"}},
		{"reachmeth", []string{"github.com/aarzilli/whydeadcode/_fixtures/pkg1.(*Astruct).ReflectMethodByName", "github.com/aarzilli/whydeadcode/_fixtures/pkg1.(*Astruct).ReflectMethodByName·f", "main.f", "main.main"}},
		{"misleadingfmt", []string{"reflect.(*rtype).Methods.func1", "reflect.(*rtype).Methods", "type:*reflect.rtype cause: main.UseMethods", "fmt.(*pp).printArg", "fmt.(*pp).doPrintln", "fmt.Fprintln", "main.main"}},
	} {
		t.Run(c.fixtureName, func(t *testing.T) {
			paths, _ := Whydeadcode(buildFixture(t, c.fixtureName))
			t.Logf("%q -> %q", c.fixtureName, paths[0])
			if len(paths[0]) < len(c.tgt) {
				t.Error("output path not long enough")
			}
			for i := range c.tgt {
				if c.tgt[i] != paths[0][i].String() {
					t.Errorf("mismatch at index %d (expected %q got %q)", i, c.tgt[i], paths[0][i])
					break
				}
			}
		})
	}
}

func buildFixture(t *testing.T, name string) io.Reader {
	t.Helper()
	cmd := exec.Command("go", "build", "-o", "_debug", "-ldflags=-dumpdep", filepath.Join("_fixtures", name)+".go")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("compilation failed for %q: %v", name, err)
	}
	os.Remove("_debug")
	return bytes.NewReader(out)
}
