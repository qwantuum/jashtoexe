package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: jashtoexe <input.jash> [output.exe]")
		os.Exit(1)
	}

	input := os.Args[1]
	output := "output.exe"
	if len(os.Args) > 2 {
		output = os.Args[2]
	}

	script, err := os.ReadFile(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %s\n", err)
		os.Exit(1)
	}

	absOut, err := filepath.Abs(output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid output path: %s\n", err)
		os.Exit(1)
	}

	tmpDir, err := os.MkdirTemp("", "jashtoexe")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create temp dir: %s\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	scriptB64 := base64.StdEncoding.EncodeToString(script)

	mainSrc := fmt.Sprintf(`package main

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/qwantuum/jash/pkg/evaluator"
	"github.com/qwantuum/jash/pkg/lexer"
	"github.com/qwantuum/jash/pkg/parser"
)

const scriptB64 = %s

func main() {
	data, err := base64.StdEncoding.DecodeString(scriptB64)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to decode script:", err)
		os.Exit(1)
	}
	source := string(data)

	l := lexer.New(source)
	tokens, errs := l.Tokenize()
	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, e)
		}
		os.Exit(1)
	}

	p := parser.New(tokens)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		for _, e := range p.Errors() {
			fmt.Fprintln(os.Stderr, e)
		}
		os.Exit(1)
	}

	env := evaluator.NewEnvironment()
	result := evaluator.Eval(program, env)

	if result != nil {
		switch r := result.(type) {
		case *evaluator.Error:
			fmt.Fprintln(os.Stderr, r.Message)
			os.Exit(1)
		default:
			if result.Type() != evaluator.NULL_OBJ && result.Type() != evaluator.RETURN_OBJ {
				fmt.Println(result.Inspect())
			}
		}
	}
}
`, "`"+scriptB64+"`")

	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainSrc), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write main.go: %s\n", err)
		os.Exit(1)
	}

	modPath := filepath.Dir(absOut)
	for {
		if _, err := os.Stat(filepath.Join(modPath, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(modPath)
		if parent == modPath {
			modPath = ""
			break
		}
		modPath = parent
	}

	goMod := []byte(fmt.Sprintf(`module app

go 1.21

require github.com/qwantuum/jash v0.0.0

replace github.com/qwantuum/jash => %s
`, filepath.ToSlash(modPath)))

	if modPath == "" {
		goMod = []byte(`module app

go 1.21

require github.com/qwantuum/jash v0.0.0

replace github.com/qwantuum/jash => C:/jash-dev
`)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), goMod, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write go.mod: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Building %s -> %s\n", input, absOut)
	cmd := exec.Command("go", "build", "-o", absOut, "-ldflags", "-s -w", ".")
	cmd.Dir = tmpDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Build failed: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("Done!")
}
