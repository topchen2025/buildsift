package analyzer

import (
	"encoding/json"
	"testing"
)

// 2026-07-20：锁定公开 JSON 契约，防止字段名或证据结构意外漂移。
func TestRenderJSON(t *testing.T) {
	diagnosis := Diagnosis{
		Found:      true,
		Tool:       "maven",
		Summary:    "target/pmd.xml was not found",
		Confidence: "high",
		Evidence:   []Evidence{{Line: 42, Text: "target/pmd.xml was not found"}},
		Cascades:   3,
		NextCheck:  "mvn -e -X",
	}

	payload, err := RenderJSON(diagnosis, "0.2.0-test")
	if err != nil {
		t.Fatal(err)
	}

	var report StructuredReport
	if err := json.Unmarshal(payload, &report); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if report.SchemaVersion != JSONSchemaVersion || report.Version != "0.2.0-test" {
		t.Fatalf("unexpected schema or tool version: %+v", report)
	}
	if report.Tool != "maven" || report.Confidence != "high" || report.RootCause != diagnosis.Summary {
		t.Fatalf("unexpected diagnosis fields: %+v", report)
	}
	if report.CascadeCount != 3 || report.NextCheck != "mvn -e -X" {
		t.Fatalf("unexpected action fields: %+v", report)
	}
	if len(report.Evidence) != 1 || report.Evidence[0].Line != 42 || report.Evidence[0].Text != diagnosis.Evidence[0].Text {
		t.Fatalf("unexpected evidence: %+v", report.Evidence)
	}
}

func TestRenderJSONUsesEmptyEvidenceArray(t *testing.T) {
	payload, err := RenderJSON(Diagnosis{Tool: "unknown", Confidence: "low"}, "dev")
	if err != nil {
		t.Fatal(err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(payload, &raw); err != nil {
		t.Fatal(err)
	}
	if string(raw["evidence"]) != "[]" {
		t.Fatalf("evidence = %s, want []", raw["evidence"])
	}
	if string(raw["root_cause"]) != `""` {
		t.Fatalf("root_cause = %s, want empty string", raw["root_cause"])
	}
}
