package analyzer

import (
	"fmt"
	"strings"
)

func Render(diagnosis Diagnosis) string {
	var output strings.Builder
	output.WriteString("BUILDSIFT DIAGNOSIS\n")
	output.WriteString("===================\n")
	if !diagnosis.Found {
		output.WriteString("NO RELIABLE ROOT CAUSE\n")
		output.WriteString("The log does not contain a failure pattern BuildSift can explain safely.\n")
		output.WriteString("\nNEXT CHECK\n")
		fmt.Fprintf(&output, "  %s\n", diagnosis.NextCheck)
		return output.String()
	}

	fmt.Fprintf(&output, "ROOT CAUSE [%s · %s]\n", strings.ToUpper(diagnosis.Confidence), strings.ToUpper(diagnosis.Tool))
	fmt.Fprintf(&output, "%s\n", diagnosis.Summary)
	output.WriteString("\nEVIDENCE\n")
	for _, evidence := range diagnosis.Evidence {
		fmt.Fprintf(&output, "  L%d  %s\n", evidence.Line, strings.TrimSpace(evidence.Text))
	}
	if diagnosis.Cascades > 0 {
		output.WriteString("\nCASCADE\n")
		fmt.Fprintf(&output, "  %d additional failure signal(s) folded\n", diagnosis.Cascades)
	}
	output.WriteString("\nNEXT CHECK\n")
	fmt.Fprintf(&output, "  %s\n", diagnosis.NextCheck)
	return output.String()
}
