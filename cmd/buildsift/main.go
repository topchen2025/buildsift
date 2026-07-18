package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/topchen2025/buildsift/internal/analyzer"
)

var version = "0.1.0-dev"

// 2026-07-18：首版坚持本地分析，避免构建日志和凭据离开用户机器。
func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) > 0 {
		switch args[0] {
		case "-h", "--help":
			printUsage(os.Stdout)
			return 0
		case "-v", "--version":
			fmt.Fprintf(os.Stdout, "buildsift %s\n", version)
			return 0
		case "--":
			if len(args) == 1 {
				fmt.Fprintln(os.Stderr, "buildsift: missing command after --")
				return 2
			}
			return runCommand(args[1:])
		}
	}

	content, err := readInput(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "buildsift: %v\n", err)
		return 2
	}

	diagnosis := analyzer.Analyze(string(content))
	fmt.Fprint(os.Stdout, analyzer.Render(diagnosis))
	return 0
}

func readInput(args []string) ([]byte, error) {
	if len(args) > 1 {
		return nil, errors.New("pass one log file, pipe stdin, or run a command after --")
	}
	if len(args) == 1 && args[0] != "-" {
		return os.ReadFile(args[0])
	}

	info, err := os.Stdin.Stat()
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeCharDevice != 0 {
		return nil, errors.New("no input; pass a log file, pipe stdin, or run a command after --")
	}
	return io.ReadAll(os.Stdin)
}

func runCommand(command []string) int {
	var capture lockedBuffer
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = io.MultiWriter(os.Stdout, &capture)
	cmd.Stderr = io.MultiWriter(os.Stderr, &capture)

	err := cmd.Run()
	if err == nil {
		return 0
	}

	fmt.Fprintln(os.Stderr)
	fmt.Fprint(os.Stderr, analyzer.Render(analyzer.Analyze(capture.String())))

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	if errors.Is(err, exec.ErrNotFound) || strings.Contains(err.Error(), "executable file not found") {
		return 127
	}
	return 1
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, `BuildSift finds the first actionable cause in noisy build logs.

Usage:
  buildsift -- <command>     stream and diagnose a failing command
  buildsift <logfile>        diagnose a saved log
  <command> | buildsift      diagnose piped input

Options:
  -h, --help                 show help
  -v, --version              show version`)
}

type lockedBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.String()
}
