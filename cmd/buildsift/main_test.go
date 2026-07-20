package main

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/topchen2025/buildsift/internal/analyzer"
)

// 2026-07-20：从真实 CLI 入口验证文件输入可产生纯净、可解析的证据包。
func TestJSONEndToEnd(t *testing.T) {
	fixture := filepath.Join("..", "..", "testdata", "maven.log")
	cmd := exec.Command("go", "run", ".", "--json", fixture)
	payload, err := cmd.Output()
	if err != nil {
		t.Fatalf("run CLI: %v", err)
	}

	var report analyzer.StructuredReport
	if err := json.Unmarshal(payload, &report); err != nil {
		t.Fatalf("invalid CLI JSON %q: %v", payload, err)
	}
	if report.SchemaVersion != analyzer.JSONSchemaVersion || report.Tool != "maven" {
		t.Fatalf("unexpected report: %+v", report)
	}
	if report.RootCause == "" || len(report.Evidence) == 0 || report.Version == "" {
		t.Fatalf("incomplete report: %+v", report)
	}
}

func TestDefaultTextOutputRemainsCompatible(t *testing.T) {
	fixture := filepath.Join("..", "..", "testdata", "npm.log")
	cmd := exec.Command("go", "run", ".", fixture)
	payload, err := cmd.Output()
	if err != nil {
		t.Fatalf("run CLI: %v", err)
	}
	if !strings.Contains(string(payload), "BUILDSIFT DIAGNOSIS") || !strings.Contains(string(payload), "ROOT CAUSE") {
		t.Fatalf("unexpected text output: %s", payload)
	}
}
