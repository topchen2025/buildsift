package analyzer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAnalyzeFixtures(t *testing.T) {
	tests := []struct {
		name       string
		tool       string
		contains   string
		confidence string
	}{
		{name: "maven", tool: "maven", contains: "target/pmd.xml", confidence: "high"},
		{name: "gradle", tool: "gradle", contains: "Could not resolve", confidence: "high"},
		{name: "npm", tool: "npm", contains: "Cannot find module", confidence: "high"},
		{name: "docker", tool: "docker", contains: "context deadline exceeded", confidence: "high"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join("..", "..", "testdata", test.name+".log")
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			diagnosis := Analyze(string(content))
			if !diagnosis.Found {
				t.Fatal("expected a diagnosis")
			}
			if diagnosis.Tool != test.tool {
				t.Fatalf("tool = %q, want %q", diagnosis.Tool, test.tool)
			}
			if diagnosis.Confidence != test.confidence {
				t.Fatalf("confidence = %q, want %q", diagnosis.Confidence, test.confidence)
			}
			if !strings.Contains(diagnosis.Summary, test.contains) {
				t.Fatalf("summary %q does not contain %q", diagnosis.Summary, test.contains)
			}
			if len(diagnosis.Evidence) == 0 || len(diagnosis.Evidence) > 3 {
				t.Fatalf("unexpected evidence count: %d", len(diagnosis.Evidence))
			}
		})
	}
}

func TestRedact(t *testing.T) {
	input := "Authorization: Bearer ghp_abcdefghijklmnopqrstuvwxyz token=supersecret /Users/alice/work https://bob:hunter2@example.com/v2/"
	redacted := Redact(input)
	for _, secret := range []string{"abcdefghijklmnopqrstuvwxyz", "supersecret", "alice", "bob", "hunter2"} {
		if strings.Contains(redacted, secret) {
			t.Fatalf("redacted output still contains %q: %s", secret, redacted)
		}
	}
	if !strings.Contains(redacted, "[REDACTED]") || !strings.Contains(redacted, "~/work") {
		t.Fatalf("unexpected redacted output: %s", redacted)
	}
}

func TestUnknownLogDoesNotGuess(t *testing.T) {
	diagnosis := Analyze("starting build\nall tasks pending\n")
	if diagnosis.Found {
		t.Fatalf("unexpected diagnosis: %+v", diagnosis)
	}
	if !strings.Contains(Render(diagnosis), "NO RELIABLE ROOT CAUSE") {
		t.Fatal("rendered output should explain that no reliable cause was found")
	}
}
