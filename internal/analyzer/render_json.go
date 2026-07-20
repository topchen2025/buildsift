package analyzer

import "encoding/json"

const JSONSchemaVersion = "1"

type StructuredEvidence struct {
	Line int    `json:"line"`
	Text string `json:"text"`
}

type StructuredReport struct {
	SchemaVersion string               `json:"schema_version"`
	Tool          string               `json:"tool"`
	Confidence    string               `json:"confidence"`
	RootCause     string               `json:"root_cause"`
	Evidence      []StructuredEvidence `json:"evidence"`
	CascadeCount  int                  `json:"cascade_count"`
	NextCheck     string               `json:"next_check"`
	Version       string               `json:"version"`
}

// 2026-07-20：固定证据包字段和空数组表示，便于 CI 与智能体稳定消费。
func NewStructuredReport(diagnosis Diagnosis, version string) StructuredReport {
	evidence := make([]StructuredEvidence, 0, len(diagnosis.Evidence))
	for _, item := range diagnosis.Evidence {
		evidence = append(evidence, StructuredEvidence{Line: item.Line, Text: item.Text})
	}
	return StructuredReport{
		SchemaVersion: JSONSchemaVersion,
		Tool:          diagnosis.Tool,
		Confidence:    diagnosis.Confidence,
		RootCause:     diagnosis.Summary,
		Evidence:      evidence,
		CascadeCount:  diagnosis.Cascades,
		NextCheck:     diagnosis.NextCheck,
		Version:       version,
	}
}

func RenderJSON(diagnosis Diagnosis, version string) ([]byte, error) {
	return json.MarshalIndent(NewStructuredReport(diagnosis, version), "", "  ")
}
