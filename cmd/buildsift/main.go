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
	jsonOutput, args := extractJSONOption(args)
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
			return runCommand(args[1:], jsonOutput)
		}
	}

	content, err := readInput(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "buildsift: %v\n", err)
		return 2
	}

	diagnosis := analyzer.Analyze(string(content))
	if err := writeDiagnosis(os.Stdout, diagnosis, jsonOutput); err != nil {
		fmt.Fprintf(os.Stderr, "buildsift: render diagnosis: %v\n", err)
		return 1
	}
	return 0
}

// 2026-07-20：仅在命令分隔符前识别 JSON 开关，避免改写被执行命令的参数。
func extractJSONOption(args []string) (bool, []string) {
	jsonOutput := false
	remaining := make([]string, 0, len(args))
	for index, arg := range args {
		if arg == "--" {
			remaining = append(remaining, args[index:]...)
			break
		}
		if arg == "--json" {
			jsonOutput = true
			continue
		}
		remaining = append(remaining, arg)
	}
	return jsonOutput, remaining
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

func runCommand(command []string, jsonOutput bool) int {
	var capture lockedBuffer
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdin = os.Stdin
	if jsonOutput {
		cmd.Stdout = io.MultiWriter(os.Stderr, &capture)
		cmd.Stderr = io.MultiWriter(os.Stderr, &capture)
	} else {
		cmd.Stdout = io.MultiWriter(os.Stdout, &capture)
		cmd.Stderr = io.MultiWriter(os.Stderr, &capture)
	}

	err := cmd.Run()
	if jsonOutput {
		if renderErr := writeDiagnosis(os.Stdout, analyzer.Analyze(capture.String()), true); renderErr != nil {
			fmt.Fprintf(os.Stderr, "buildsift: render diagnosis: %v\n", renderErr)
			if err == nil {
				return 1
			}
		}
	}
	if err == nil {
		return 0
	}

	if !jsonOutput {
		fmt.Fprintln(os.Stderr)
		fmt.Fprint(os.Stderr, analyzer.Render(analyzer.Analyze(capture.String())))
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	if errors.Is(err, exec.ErrNotFound) || strings.Contains(err.Error(), "executable file not found") {
		return 127
	}
	return 1
}

func writeDiagnosis(w io.Writer, diagnosis analyzer.Diagnosis, jsonOutput bool) error {
	if !jsonOutput {
		_, err := fmt.Fprint(w, analyzer.Render(diagnosis))
		return err
	}
	payload, err := analyzer.RenderJSON(diagnosis, version)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(payload))
	return err
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, `BuildSift finds the first actionable cause in noisy build logs.

Usage:
  buildsift -- <command>     stream and diagnose a failing command
  buildsift <logfile>        diagnose a saved log
  <command> | buildsift      diagnose piped input

Options:
  -h, --help                 show help
  -v, --version              show version
      --json                 emit a structured evidence package`)
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
